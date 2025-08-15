from __future__ import annotations

import os
import re
from typing import List, Dict, Any

import numpy as np
import requests
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer, util

OLLAMA_URL = os.getenv("OLLAMA_URL", "http://ollama:11434").rstrip("/")
MODEL_GEN = os.getenv("GEN_MODEL", "qwen2.5:3b-instruct-q4_0")

EMBED_MODEL = os.getenv("EMBED_MODEL", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
SIM_THRESHOLD = float(os.getenv("SIM_THRESHOLD", "0.84"))
SIM_THRESHOLD_HARD = float(os.getenv("SIM_THRESHOLD_HARD", "0.88"))
HISTORY_LIMIT = int(os.getenv("HISTORY_LIMIT", "16"))
LLM_TEMPERATURE = float(os.getenv("LLM_TEMPERATURE", "0.3"))
LLM_MAX_TOKENS = int(os.getenv("LLM_MAX_TOKENS", "300"))
REQUEST_TIMEOUT = int(os.getenv("REQUEST_TIMEOUT", "120"))
ALWAYS_CHAT = os.getenv("ALWAYS_CHAT", "true").lower() == "true"
USE_MINI = os.getenv("USE_MINI", "true").lower() == "true"

app = FastAPI(title="Talapker NLP API", version="1.2-hybrid")

FILLERS = ("ну","типа","короче","вообще","вообщем","смотри","слушай","эй","ей","чё","че","ёу","ээ","аа")
FILLERS_RE = re.compile(r"\b(" + "|".join(map(re.escape, FILLERS)) + r")\b", re.I | re.U)
NON_WORDS_RE = re.compile(r"[^\w\s]+", re.U)
SPACES_RE = re.compile(r"\s+", re.U)

CANON = {"wkau":"wkatu","zkatu":"wkatu","жангир":"wkatu","жангір":"wkatu"}

def normalize(text: str) -> str:
    t = (text or "").lower().strip()
    t = FILLERS_RE.sub(" ", t)
    t = NON_WORDS_RE.sub(" ", t)
    for k, v in CANON.items():
        t = t.replace(k, v)
    return SPACES_RE.sub(" ", t).strip()

CANDIDATES = [
    {"slug": "smalltalk", "phrases": [
        "привет","хай","здорово","как ты","как дела","ты тут","салам","салем","ассалаумағалейкум","hi","hello"
    ]},
    {"slug": "programs", "phrases": [
        "образовательные программы","список программ","какие направления есть",
        "білім беру бағдарламалары","бағдарламалар тізімі","қандай мамандықтар бар"
    ]},
    {"slug": "documents", "phrases": [
        "документы для поступления","какие документы нужны","перечень документов",
        "құжаттар","қандай құжаттар керек","құжаттар тізімі"
    ]},
    {"slug": "grants", "phrases": [
        "гранты","стипендии","как получить грант","проходные баллы",
        "гранттар","стипендия","грантқа қалай түсемін","өту балы"
    ]},
    {"slug": "dorm", "phrases": [
        "общежитие","проживание","места в общаге","комнаты",
        "жатақхана","тұру","орын беру","жатын орын"
    ]},
    {"slug": "admissions", "phrases": [
        "приёмная комиссия","контакты приёмной","сроки приёма","стоимость обучения",
        "қабылдау комиссиясы","байланыс","қабылдау мерзімдері","оқу ақысы"
    ]},
    {"slug": "why-wkatu", "phrases": [
        "почему wkatu","зачем wkatu","почему именно wkatu","какие преимущества wkatu",
        "преимущества университета","плюсы университета","почему выбрать wkatu",
        "неге wkatu","неге дәл wkatu","артықшылықтары wkatu","жақсы жақтары wkatu"
    ]},
    {"slug": "campus", "phrases": [
        "какие развлечения","чем заняться в wkatu","студенческая жизнь","кружки и клубы",
        "мероприятия","досуг","студклуб","спортзал","спорт секции",
        "студенттік өмір","үйірмелер","іс-шаралар","демалыс","спорт секциялар"
    ]},
    {"slug": "dorm", "phrases": [
        "общежитие","проживание","места в общаге","комнаты",
        "жатақхана","тұру","орын беру","жатын орын",
        "очередь в общежитие","какая очередь","очередность","лист ожидания","мест нет"
    ]},

]

flat_phrases: List[str] = []
phrase_to_slug: List[str] = []
for c in CANDIDATES:
    for p in c["phrases"]:
        flat_phrases.append(p)
        phrase_to_slug.append(c["slug"])

model = SentenceTransformer(EMBED_MODEL)
embeds = model.encode(flat_phrases, normalize_embeddings=True, convert_to_tensor=True)

class AskReq(BaseModel):
    text: str

class AskResp(BaseModel):
    slug: str
    confidence: float
    best_phrase: str

class ChatReq(BaseModel):
    text: str
    history: List[Dict[str, str]] | None = None

class ChatResp(BaseModel):
    answer: str

class ChatPlusResp(BaseModel):
    intent_slug: str
    intent_confidence: float
    mini_answer: str | None
    llm_answer: str

@app.get("/health")
def health() -> Dict[str, Any]:
    return {
        "status": "ok",
        "embed_model": EMBED_MODEL,
        "gen_model": MODEL_GEN,
        "threshold": SIM_THRESHOLD,
        "threshold_hard": SIM_THRESHOLD_HARD,
        "always_chat": ALWAYS_CHAT,
        "use_mini": USE_MINI,
        "version": "1.2-hybrid",
    }

@app.post("/ask", response_model=AskResp)
def ask(req: AskReq) -> AskResp:
    if ALWAYS_CHAT:
        return AskResp(slug="unknown", confidence=0.0, best_phrase="")
    text = normalize(req.text or "")
    if not text:
        return AskResp(slug="unknown", confidence=0.0, best_phrase="")
    q_emb = model.encode([text], normalize_embeddings=True, convert_to_tensor=True)
    sims = util.cos_sim(q_emb, embeds).cpu().numpy()[0]
    idx = int(np.argmax(sims))
    score = float(sims[idx])
    slug = phrase_to_slug[idx]
    best = flat_phrases[idx]
    if score < SIM_THRESHOLD:
        return AskResp(slug="unknown", confidence=score, best_phrase=best)
    return AskResp(slug=slug, confidence=score, best_phrase=best)

KZ_CHARS = set("әғқңөұүіӘҒҚҢӨҰҮІ")
def detect_lang(s: str) -> str:
    if any(ch in KZ_CHARS for ch in s):
        return "kz"
    low = s.lower()
    if any(w in low for w in ["жатақхана","гранттар","құжат","бағдарлама","сәлем"]):
        return "kz"
    return "ru"

MINI: Dict[str, Dict[str, str]] = {
    "smalltalk": {
        "ru": "Привет! Чем помочь по WKATU: поступление, программы, гранты, общага?",
        "kz": "Сәлем! WKATU бойынша не керек: қабылдау, бағдарламалар, гранттар, жатақхана?",
    },
    "programs": {
        "ru": "Программы WKATU: агро, ветеринария, инж-тех, IT и др. Нужен список/профили?",
        "kz": "WKATU бағдарламалары: агро, ветеринария, инж-тех, IT және т.б. Тізімі керек пе?",
    },
    "documents": {
        "ru": "Документы: ID/паспорт, аттестат, ЕНТ, фото 3×4, мед-075У, заявление и др.",
        "kz": "Құжаттар: Жеке куәлік, аттестат, ҰБТ, 3×4 фото, 075-У, өтініш және т.б.",
    },
    "grants": {
        "ru": "Гранты: по ЕНТ и квотам, сроки и проходные — в приёмке. Подсказать по баллам?",
        "kz": "Гранттар: ҰБТ және квоталар бойынша, мерзім/өту балдары — қабылдауда.",
    },
    "admissions": {
        "ru": "Приёмка: контакты, сроки подачи и стоимость — могу подсказать детали.",
        "kz": "Қабылдау: байланыс, өтініс мерзімдері, оқу ақысы — мәлімет беремін.",
    },
    "why-wkatu": {
        "ru": "Почему WKATU: практика, сильные агро/инж направления, стипендии, общежитие.",
        "kz": "Неге WKATU: тәжірибе, мықты агро/инж бағыттар, стипендия, жатақхана.",
    },
    "campus": {
        "ru": "Кампус: клубы, спорт, мероприятия, волонтёрство. Что интересует?",
        "kz": "Кампус: клубтар, спорт, іс-шаралар, волонтёрлік. Не қызықты?",
    },
    "dorm": {
        "ru": "Общежитие: приоритет иногородним/льготникам, распределение по очереди. Очередь по заявкам приёмки: статус можно уточнить в деканате/общежитии по номеру заявки.",
        "kz": "Жатақхана: басымдық қаладан тыс/жеңілдік санаттарына, кезекпен беріледі. Кезек қабылдау өтініштері бойынша: мәртебені өтініш нөмірі арқылы нақтылаңыз.",
    },
}

def make_mini(slug: str, lang: str) -> str | None:
    if not USE_MINI:
        return None
    m = MINI.get(slug)
    if not m:
        return None
    return m.get(lang, m.get("ru"))

SYSTEM_PROMPT = (
    "Ты — официальный помощник TalapkerBot WKATU (ЗКАТУ им. Жангир хана). "
    "Отвечай дружелюбно и кратко на русском или казахском (ориентируйся на язык пользователя). "
    "Если вопрос связан с поступлением, программами, грантами, общежитием или студенческой жизнью — отвечай конкретно с акцентом на WKATU. "
    "Если спрашивают про другие университеты — отметь, что выбор зависит от критериев, и корректно укажи преимущества WKATU. "
    "Если вопрос не про образование/WKATU — всё равно дай краткий полезный ответ по сути и мягко предложи вернуться к теме WKATU."
)

def _trim_history(history: List[Dict[str, str]] | None) -> List[Dict[str, str]]:
    if not history:
        return []
    out: List[Dict[str, str]] = []
    for m in history[-HISTORY_LIMIT:]:
        role = m.get("role")
        content = (m.get("content") or "").strip()
        if role in ("user", "assistant") and content:
            out.append({"role": role, "content": content})
    return out

def _ollama_chat(messages: List[Dict[str, str]], temperature: float, num_predict: int) -> str:
    payload = {
        "model": MODEL_GEN,
        "messages": messages,
        "stream": False,
        "options": {"temperature": temperature, "num_predict": num_predict},
    }
    r = requests.post(f"{OLLAMA_URL}/v1/chat/completions", json=payload, timeout=REQUEST_TIMEOUT)
    r.raise_for_status()
    data = r.json()
    if isinstance(data, dict):
        choices = data.get("choices") or []
        if choices and isinstance(choices, list):
            msg = choices[0].get("message") or {}
            return (msg.get("content") or "").strip()
    return ""

@app.post("/chat", response_model=ChatResp)
def chat(req: ChatReq) -> ChatResp:
    prompt = (req.text or "").strip()
    if not prompt:
        return ChatResp(answer="Сұрағыңыз бос сияқты. Қысқаша жазыңызшы.")
    messages: List[Dict[str, str]] = [{"role": "system", "content": SYSTEM_PROMPT}]
    messages += _trim_history(req.history)
    messages.append({"role": "user", "content": prompt})
    try:
        answer = _ollama_chat(messages, LLM_TEMPERATURE, LLM_MAX_TOKENS)
        if not answer:
            answer = _ollama_chat(messages, min(LLM_TEMPERATURE + 0.2, 0.8), LLM_MAX_TOKENS)
        if not answer:
            answer = "Помогу по WKATU: поступление, программы, гранты, общежитие. Что интересно?"
        return ChatResp(answer=answer)
    except Exception:
        return ChatResp(answer="Қызмет уақытша қолжетімсіз. Кейінірек қайталап көріңіз.")

@app.post("/chat_plus", response_model=ChatPlusResp)
def chat_plus(req: ChatReq) -> ChatPlusResp:
    prompt = (req.text or "").strip()
    if not prompt:
        return ChatPlusResp(intent_slug="unknown", intent_confidence=0.0, mini_answer=None,
                            llm_answer="Сұрағыңыз бос сияқты. Қысқаша жазыңызшы.")

    lang = detect_lang(prompt)
    intent_slug, intent_conf = "unknown", 0.0
    mini_answer: str | None = None
    if USE_MINI:
        t = normalize(prompt)
        if t:
            q_emb = model.encode([t], normalize_embeddings=True, convert_to_tensor=True)
            sims = util.cos_sim(q_emb, embeds).cpu().numpy()[0]
            idx = int(np.argmax(sims))
            intent_conf = float(sims[idx])
            intent_slug = phrase_to_slug[idx]
            if intent_conf >= SIM_THRESHOLD_HARD:
                mini_answer = make_mini(intent_slug, lang)

    messages: List[Dict[str, str]] = [{"role": "system", "content": SYSTEM_PROMPT}]
    messages += _trim_history(req.history)

    if intent_slug != "unknown" and intent_conf >= SIM_THRESHOLD:
        hint = f"(Похоже, что вопрос про: {intent_slug}. Дай краткий точный ответ.)"
        messages.append({"role": "user", "content": f"{prompt}\n\n{hint}"})
    else:
        messages.append({"role": "user", "content": prompt})

    try:
        llm_answer = _ollama_chat(messages, LLM_TEMPERATURE, LLM_MAX_TOKENS)
        if not llm_answer:
            llm_answer = _ollama_chat(messages, min(LLM_TEMPERATURE + 0.2, 0.8), LLM_MAX_TOKENS)
        if not llm_answer:
            llm_answer = "Помогу по WKATU: поступление, программы, гранты, общежитие. Что интересно?"
        return ChatPlusResp(
            intent_slug=intent_slug,
            intent_confidence=intent_conf,
            mini_answer=mini_answer,
            llm_answer=llm_answer,
        )
    except Exception:
        return ChatPlusResp(
            intent_slug="unknown",
            intent_confidence=0.0,
            mini_answer=mini_answer,
            llm_answer="Қызмет уақытша қолжетімсіз. Кейінірек қайталап көріңіз.",
        )
