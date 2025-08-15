package handlers

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	contentclient "telegramBot/bot/internal/client"
	keyboard "telegramBot/bot/internal/keybord"
	"telegramBot/bot/internal/nlpclient"
)

const smalltalkWindow = 1
const classifyThreshold = 0.65

func stKey(chatID int64) string { return "st:" + strconv.FormatInt(chatID, 10) }

type Bot struct {
	API     *tgbotapi.BotAPI
	Store   sync.Map
	History sync.Map

	APICl *contentclient.Client
	NLP   *nlpclient.Client
}

func New(api *tgbotapi.BotAPI, apiCl *contentclient.Client, nlp *nlpclient.Client) *Bot {
	return &Bot{API: api, APICl: apiCl, NLP: nlp}
}

func (b *Bot) langOf(chatID int64) string {
	if v, ok := b.Store.Load(chatID); ok {
		if s, ok := v.(string); ok && (s == "ru" || s == "kz") {
			return s
		}
	}
	return "ru"
}

func (b *Bot) getHistory(chatID int64) []map[string]string {
	if v, ok := b.History.Load(chatID); ok {
		if h, ok := v.([]map[string]string); ok {
			return h
		}
	}
	return nil
}

func (b *Bot) pushUser(chatID int64, text string) {
	h := b.getHistory(chatID)
	h = append(h, map[string]string{"role": "user", "content": text})
	if len(h) > 8 {
		h = h[len(h)-8:]
	}
	b.History.Store(chatID, h)
}

func (b *Bot) pushAssistant(chatID int64, text string) {
	h := b.getHistory(chatID)
	h = append(h, map[string]string{"role": "assistant", "content": text})
	if len(h) > 8 {
		h = h[len(h)-8:]
	}
	b.History.Store(chatID, h)
}

func (b *Bot) enterSmalltalk(chatID int64) {
	b.Store.Store(stKey(chatID), smalltalkWindow)
}
func (b *Bot) inSmalltalk(chatID int64) bool {
	if v, ok := b.Store.Load(stKey(chatID)); ok {
		if n, ok := v.(int); ok && n > 0 {
			b.Store.Store(stKey(chatID), n-1)
			return true
		}
	}
	return false
}

var reHello = regexp.MustCompile(`(?i)\b(салам|салем|сәлем|прив(?:ет)?|здар(?:ова|ов)|эй|ей|хей|hey|hi|hello|ассалаумағалейкум|ассалам|сәлеметсіз бе|салям|салямалейкум|здоровченко)\b`)
var reHow = regexp.MustCompile(`(?i)\b(как\s*(ты|дела|поживаешь)|что\s*нового|чё\s*как|че\s*как|қалайсың|жағдай|how'?s?\s*it\s*going|whats?'?s?\s*up)\b`)
var reBuddy = regexp.MustCompile(`(?i)\b(друг|бро|брат|братик|дружище|ай|баур|бауыр|чел)\b`)

func isSmalltalk(text string) bool {
	t := strings.TrimSpace(text)
	return reHello.MatchString(t) || reHow.MatchString(t) ||
		(reBuddy.MatchString(t) && (reHello.MatchString(t) || reHow.MatchString(t)))
}

var reFiller = regexp.MustCompile(`(?i)\b(ну|вот|типа|короче|слушай|смотри|вообще|как бы|вообщем|шо|че|чё|ээ|аа|емае|ёмаё|ёмае|чел|бро)\b`)

func normalizeForClassify(text string) string {
	t := strings.ToLower(strings.TrimSpace(text))
	t = reFiller.ReplaceAllString(t, " ")
	t = strings.Join(strings.Fields(t), " ")
	return t
}

var badOutRe = regexp.MustCompile(`(?i)(зад|секс|эрот|порно|трах|сись|fuck|bitch)`)

func sanitizeAnswer(s string, lang string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "尽力帮", "")
	s = strings.ReplaceAll(s, "有什么我可以做的吗", "")
	s = strings.ReplaceAll(s, "mind relaxing", "")

	if badOutRe.MatchString(s) {
		if lang == "kz" {
			return "Кешіріңіз, әңгімені WKATU тақырыптарына бұрайық: қабылдау, бағдарламалар, гранттар, жатақхана. Неден бастаймыз?"
		}
		return "Давайте вернёмся к полезному: WKATU — поступление, программы, гранты или общежитие. Что именно интересно?"
	}
	return s
}

//	func friendlyPrefixed(text, lang string) string {
//		if lang == "kz" {
//			return "Достық һәм қысқа жауап бер. WKATU тақырыбына (қабылдау, бағдарламалар, гранттар, жатақхана) фокус жаса. Оффтоп болса, сыпайы түрде әңгімени WKATU-ға бұр. " + text
//		}
//		return "Отвечай дружелюбно и кратко. Держи фокус на WKATU (поступление, программы, гранты, общежитие). На оффтоп отвечай вежливо и мягко направляй к теме WKATU. " + text
//	}
func cannedSmalltalk(lang string) string {
	if lang == "kz" {
		return "Жақсы! Көмектесуге дайынмын. WKATU бойынша не қызықты: қабылдау, бағдарламалар, гранттар, жатақхана?"
	}
	return "Отлично! Готов помочь. Что по WKATU интересно: поступление, программы, гранты, общежитие?"
}

func (b *Bot) HandleMessage(upd tgbotapi.Update) {
	if upd.Message == nil {
		return
	}
	chatID := upd.Message.Chat.ID
	text := upd.Message.Text

	switch text {
	case "/start":
		b.Store.Delete(chatID)
		b.Store.Delete(stKey(chatID))
		b.History.Delete(chatID)
		msg := tgbotapi.NewMessage(chatID, "Тілді таңдаңыз / Выберите язык:")
		msg.ReplyMarkup = keyboard.LangKeyboard()
		b.API.Send(msg)
		return
	case "/help":
		b.API.Send(tgbotapi.NewMessage(chatID, "Выберите язык, затем раздел."))
		return
	case "🇷🇺 Русский":
		b.Store.Store(chatID, "ru")
		b.Store.Delete(stKey(chatID))
		b.History.Delete(chatID)
		msg := tgbotapi.NewMessage(chatID, "Выберите раздел:")
		msg.ReplyMarkup = keyboard.Menu(keyboard.MenuRU)
		b.API.Send(msg)
		return
	case "🇰🇿 Қазақша":
		b.Store.Store(chatID, "kz")
		b.Store.Delete(stKey(chatID))
		b.History.Delete(chatID)
		msg := tgbotapi.NewMessage(chatID, "Бөлімді таңдаңыз:")
		msg.ReplyMarkup = keyboard.Menu(keyboard.MenuKZ)
		b.API.Send(msg)
		return
	}

	lang := b.langOf(chatID)
	forceSmalltalk := b.inSmalltalk(chatID)

	switch text {
	case "🎓 Образовательные программы":
		b.sendFromAPI(chatID, "programs", lang)
		return
	case "📑 Документы":
		b.sendFromAPI(chatID, "documents", lang)
		return
	case "🎁 Гранты":
		b.sendFromAPI(chatID, "grants", lang)
		return
	case "🏠 Общежитие":
		b.sendFromAPI(chatID, "dorm", lang)
		return

	case "Почему WKATU?", "Преимущества WKATU", "Почему именно WKATU?", "Артықшылықтары WKATU":
		b.sendFromAPI(chatID, "why-wkatu", lang)
		return
	case "Студенческая жизнь", "Развлечения в WKATU", "Клубы и кружки", "Студенттік өмір":
		b.sendFromAPI(chatID, "campus", lang)
		return

	case "🎓 Білім беру бағдарламалары":
		b.sendFromAPI(chatID, "programs", lang)
		return
	case "📑 Құжаттар":
		b.sendFromAPI(chatID, "documents", lang)
		return
	case "🎁 Гранттар":
		b.sendFromAPI(chatID, "grants", lang)
		return
	case "🏠 Жатақхана":
		b.sendFromAPI(chatID, "dorm", lang)
		return
	}

	if (isSmalltalk(text) || forceSmalltalk) && b.NLP != nil {
		b.pushUser(chatID, text)
		if mini, llm, err := b.NLP.ChatPlus(text, lang, b.getHistory(chatID)); err == nil {
			if mini != nil && strings.TrimSpace(*mini) != "" {
				out := sanitizeAnswer(*mini, lang)
				b.API.Send(tgbotapi.NewMessage(chatID, out))
				b.pushAssistant(chatID, out)
				b.enterSmalltalk(chatID)
				return
			}
			if strings.TrimSpace(llm) == "" {
				llm = "Помогу по WKATU: поступление, программы, гранты, общежитие. Что интересно?"
			}
			out := sanitizeAnswer(llm, lang)
			b.API.Send(tgbotapi.NewMessage(chatID, out))
			b.pushAssistant(chatID, out)
			b.enterSmalltalk(chatID)
			return
		}
		cs := cannedSmalltalk(lang)
		b.API.Send(tgbotapi.NewMessage(chatID, cs))
		b.pushAssistant(chatID, cs)
		b.enterSmalltalk(chatID)
		return
	}

	if b.NLP != nil {
		nt := normalizeForClassify(text)
		if slug, conf, err := b.NLP.Classify(nt); err == nil && slug == "smalltalk" && conf >= classifyThreshold {
			b.pushUser(chatID, text)
			if mini, llm, err := b.NLP.ChatPlus(text, lang, b.getHistory(chatID)); err == nil {
				if mini != nil && strings.TrimSpace(*mini) != "" {
					out := sanitizeAnswer(*mini, lang)
					b.API.Send(tgbotapi.NewMessage(chatID, out))
					b.pushAssistant(chatID, out)
					b.enterSmalltalk(chatID)
					return
				}
				if strings.TrimSpace(llm) == "" {
					llm = "Помогу по WKATU: поступление, программы, гранты, общежитие. Что интересно?"
				}
				out := sanitizeAnswer(llm, lang)
				b.API.Send(tgbotapi.NewMessage(chatID, out))
				b.pushAssistant(chatID, out)
				b.enterSmalltalk(chatID)
				return
			}
		}

		b.pushUser(chatID, text)
		if mini, llm, err := b.NLP.ChatPlus(text, lang, b.getHistory(chatID)); err == nil {
			if mini != nil && strings.TrimSpace(*mini) != "" {
				out := sanitizeAnswer(*mini, lang)
				b.API.Send(tgbotapi.NewMessage(chatID, out))
				b.pushAssistant(chatID, out)
				return
			}
			if strings.TrimSpace(llm) == "" {
				llm = "Помогу по WKATU: поступление, программы, гранты, общежитие. Что интересно?"
			}
			out := sanitizeAnswer(llm, lang)
			b.API.Send(tgbotapi.NewMessage(chatID, out))
			b.pushAssistant(chatID, out)
			return
		}
	}

	if lang == "kz" {
		b.API.Send(tgbotapi.NewMessage(chatID, "Түсінбедім 🙂 WKATU бойынша көмектесе аламын: қабылдау, бағдарламалар, гранттар, жатақхана. Қайсысы қызықты?"))
	} else {
		b.API.Send(tgbotapi.NewMessage(chatID, "Понял не всё 🙂 Могу помочь по WKATU: поступление, программы, гранты, общежитие. Что именно интересно?"))
	}
}

func (b *Bot) sendFromAPI(chatID int64, slug, lang string) {
	c, err := b.APICl.Get(slug, lang)
	if err != nil {
		b.API.Send(tgbotapi.NewMessage(chatID, "Данные скоро обновим."))
		return
	}
	b.API.Send(tgbotapi.NewMessage(chatID, c.Title+"\n\n"+c.Body))
}
