from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer, util
import numpy as np
import os, requests

OLLAMA_URL = os.getenv("OLLAMA_URL", "http://ollama:11434")
MODEL_GEN = os.getenv("GEN_MODEL", "qwen2.5:3b-instruct-q4_0")

app = FastAPI(title="Talapker NLP API", version="1.0")

MODEL_NAME = "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
model = SentenceTransformer(MODEL_NAME)

CANDIDATES = [
    {"slug": "smalltalk", "phrases": [
        "привет", "хай", "здорово", "как ты", "как дела",
        "ты тут", "салам", "салем", "ассалаумағалейкум",
        "hi", "hello"
    ]},
    {"slug": "programs", "phrases": [
        "образовательные программы", "список программ", "какие направления есть",
        "білім беру бағдарламалары", "бағдарламалар тізімі", "қандай мамандықтар бар"
    ]},
    {"slug": "documents", "phrases": [
        "документы для поступления", "какие документы нужны", "перечень документов",
        "құжаттар", "қандай құжаттар керек", "құжаттар тізімі"
    ]},
    {"slug": "grants", "phrases": [
        "гранты", "стипендии", "как получить грант", "проходные баллы",
        "гранттар", "стипендия", "грантқа қалай түсемін", "өту балы"
    ]},
    {"slug": "dorm", "phrases": [
        "общежитие", "проживание", "места в общаге", "комнаты",
        "жатақхана", "тұру", "орын беру", "жатын орын"
    ]},
    {"slug": "admissions", "phrases": [
        "приёмная комиссия", "контакты приёмной", "сроки приёма", "стоимость обучения",
        "қабылдау комиссиясы", "байланыс", "қабылдау мерзімдері", "оқу ақысы"
    ]},
]

flat_phrases, phrase_to_slug = [], []
for c in CANDIDATES:
    for p in c["phrases"]:
        flat_phrases.append(p)
        phrase_to_slug.append(c["slug"])

embeds = model.encode(flat_phrases, normalize_embeddings=True, convert_to_tensor=True)

class AskReq(BaseModel):
    text: str

class AskResp(BaseModel):
    slug: str
    confidence: float
    best_phrase: str

@app.get("/health")
def health():
    return {"status": "ok"}

@app.post("/ask", response_model=AskResp)
def ask(req: AskReq):
    text = (req.text or "").strip()
    if not text:
        return AskResp(slug="unknown", confidence=0.0, best_phrase="")
    q_emb = model.encode([text], normalize_embeddings=True, convert_to_tensor=True)
    sims = util.cos_sim(q_emb, embeds).cpu().numpy()[0]
    idx = int(np.argmax(sims))
    score = float(sims[idx])
    slug = phrase_to_slug[idx]
    best = flat_phrases[idx]
    if score < 0.87:
        return AskResp(slug="unknown", confidence=score, best_phrase=best)
    return AskResp(slug=slug, confidence=score, best_phrase=best)


class ChatReq(BaseModel):
    text: str
    history: list[dict] | None = None

class ChatResp(BaseModel):
    answer: str

@app.post("/chat", response_model=ChatResp)
def chat(req: ChatReq):
    prompt = (req.text or "").strip()
    if not prompt:
        return ChatResp(answer="Сұрағыңыз бос сияқты. Қысқаша жазыңызшы.")
    payload = {
        "model": MODEL_GEN,
        "messages": (req.history or []) + [
            {"role": "system", "content": "Ты помощник TalapkerBot WKATU. Отвечай кратко, дружелюбно, на русском или казахском, по языку пользователя. Если вопрос о поступлении/программах/грантах/общежитии — будь точным и по делу."},
            {"role": "user", "content": prompt}
        ],
        "stream": False
    }
    r = requests.post(f"{OLLAMA_URL}/v1/chat/completions", json=payload, timeout=120)
    if r.status_code != 200:
        return ChatResp(answer="Қызмет уақытша қолжетімсіз. Кейінірек қайталап көріңіз.")
    data = r.json()
    answer = data["choices"][0]["message"]["content"]
    return ChatResp(answer=answer)
