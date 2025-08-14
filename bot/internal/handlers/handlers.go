package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"sync"
	"telegramBot/bot/internal/client"
	"telegramBot/bot/internal/keybord"
	"telegramBot/bot/internal/nlpclient"
)

type Bot struct {
	API     *tgbotapi.BotAPI
	Store   sync.Map
	History sync.Map
	APICl   *contentclient.Client
	NLP     *nlpclient.Client
}

func isMenuSelection(text string) bool {
	switch text {
	case "🎓 Образовательные программы",
		"📑 Документы",
		"🎁 Гранты",
		"🏠 Общежитие",
		"🎓 Білім беру бағдарламалары",
		"📑 Құжаттар",
		"🎁 Гранттар",
		"🏠 Жатақхана":
		return true
	default:
		return false
	}
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
func (b *Bot) HandleMessage(upd tgbotapi.Update) {
	if upd.Message == nil {
		return
	}
	chatID := upd.Message.Chat.ID
	text := upd.Message.Text

	switch text {
	case "/start":
		b.Store.Delete(chatID)
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
		b.History.Delete(chatID)
		msg := tgbotapi.NewMessage(chatID, "Выберите раздел:")
		msg.ReplyMarkup = keyboard.Menu(keyboard.MenuRU)
		b.API.Send(msg)
		return
	case "🇰🇿 Қазақша":
		b.Store.Store(chatID, "kz")
		b.History.Delete(chatID)
		msg := tgbotapi.NewMessage(chatID, "Бөлімді таңдаңыз:")
		msg.ReplyMarkup = keyboard.Menu(keyboard.MenuKZ)
		b.API.Send(msg)
		return
	}

	lang := b.langOf(chatID)

	if isMenuSelection(text) {
		switch text {
		// RU
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
		// KZ
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
	}

	if b.NLP != nil {
		if slug, conf, err := b.NLP.Classify(text); err == nil && conf >= 0.65 {
			if slug == "smalltalk" {
				b.pushUser(chatID, text)
				if ans, err := b.NLP.Chat(text, lang, b.getHistory(chatID)); err == nil && ans != "" {
					b.API.Send(tgbotapi.NewMessage(chatID, ans))
					b.pushAssistant(chatID, ans)
					return
				}
			} else if slug != "unknown" {
				b.sendFromAPI(chatID, slug, lang)
				return
			}
		}

		b.pushUser(chatID, text)
		if ans, err := b.NLP.Chat(text, lang, b.getHistory(chatID)); err == nil && ans != "" {
			b.API.Send(tgbotapi.NewMessage(chatID, ans))
			b.pushAssistant(chatID, ans)
			fmt.Println("NLP chat error:", err)
			return
		}
	}

	if lang == "kz" {
		b.API.Send(tgbotapi.NewMessage(chatID, "Сұрағыңды түсінбедім. Мәзірден таңда немесе қысқаша жаз."))
	} else {
		b.API.Send(tgbotapi.NewMessage(chatID, "Не понял. Выберите пункт меню или сформулируйте короче."))
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
