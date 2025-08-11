package services

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"telegramBot/dbConn"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var menuRu = []string{"🎓 Образовательные программы", "📑 Документы", "🎁 Гранты", "🏠 Общежитие"}
var menuKz = []string{"🎓 Білім беру бағдарламалары", "📑 Құжаттар", "🎁 Гранттар", "🏠 Жатақхана"}

var lang = map[int64]string{}

func Bot() {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN не задан")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Бот @%s запущен", bot.Self.UserName)

	bot.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "Запустить бота"},
		tgbotapi.BotCommand{Command: "help", Description: "Помощь"},
	))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка...")
			return
		case upd := <-updates:
			if upd.Message == nil {
				continue
			}
			chatID := upd.Message.Chat.ID
			text := upd.Message.Text

			switch text {
			case "/start":
				lang[chatID] = ""
				msg := tgbotapi.NewMessage(chatID, "Тілді таңдаңыз / Выберите язык:")
				msg.ReplyMarkup = languageKeyboard()
				bot.Send(msg)
				continue
			case "/help":
				bot.Send(tgbotapi.NewMessage(chatID, "Выберите язык и раздел. Напишите вопрос — отвечу."))
				continue
			case "🇷🇺 Русский":
				lang[chatID] = "ru"
				msg := tgbotapi.NewMessage(chatID, "Выберите раздел:")
				msg.ReplyMarkup = menuKeyboard(menuRu)
				bot.Send(msg)
				continue
			case "🇰🇿 Қазақша":
				lang[chatID] = "kz"
				msg := tgbotapi.NewMessage(chatID, "Бөлімді таңдаңыз:")
				msg.ReplyMarkup = menuKeyboard(menuKz)
				bot.Send(msg)
				continue
			}

			lng := chosenLang(chatID)

			// RU
			switch text {
			case "🎓 Образовательные программы":
				sendFromDB(bot, chatID, "programs", lng)
				continue
			case "📑 Документы":
				sendFromDB(bot, chatID, "documents", lng)
				continue
			case "🎁 Гранты":
				sendFromDB(bot, chatID, "grants", lng)
				continue
			case "🏠 Общежитие":
				sendFromDB(bot, chatID, "dorm", lng)
				continue
			}
			// KZ
			switch text {
			case "🎓 Білім беру бағдарламалары":
				sendFromDB(bot, chatID, "programs", lng)
				continue
			case "📑 Құжаттар":
				sendFromDB(bot, chatID, "documents", lng)
				continue
			case "🎁 Гранттар":
				sendFromDB(bot, chatID, "grants", lng)
				continue
			case "🏠 Жатақхана":
				sendFromDB(bot, chatID, "dorm", lng)
				continue
			}

			if lng == "kz" {
				bot.Send(tgbotapi.NewMessage(chatID, "Сұрағыңды түсінбедім. Мәзірден таңда немесе қысқаша жаз."))
			} else {
				bot.Send(tgbotapi.NewMessage(chatID, "Не понял. Выберите пункт меню или сформулируйте короче."))
			}
		}
	}
}

func chosenLang(chatID int64) string {
	if l, ok := lang[chatID]; ok && (l == "ru" || l == "kz") {
		return l
	}
	return "ru"
}

func sendFromDB(bot *tgbotapi.BotAPI, chatID int64, slug, lng string) {
	title, body, err := dbConn.Fetch(slug, lng)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Данные скоро обновим."))
		return
	}
	bot.Send(tgbotapi.NewMessage(chatID, title+"\n\n"+body))
}

func languageKeyboard() tgbotapi.ReplyKeyboardMarkup {
	btns := [][]tgbotapi.KeyboardButton{
		{tgbotapi.NewKeyboardButton("🇰🇿 Қазақша"), tgbotapi.NewKeyboardButton("🇷🇺 Русский")},
	}
	kb := tgbotapi.NewReplyKeyboard(btns...)
	kb.ResizeKeyboard = true
	return kb
}

func menuKeyboard(items []string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, label := range items {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(label)))
	}
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	return kb
}
