package keyboard

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var MenuRU = []string{"🎓 Образовательные программы", "📑 Документы", "🎁 Гранты", "🏠 Общежитие"}
var MenuKZ = []string{"🎓 Білім беру бағдарламалары", "📑 Құжаттар", "🎁 Гранттар", "🏠 Жатақхана"}

func LangKeyboard() tgbotapi.ReplyKeyboardMarkup {
	kb := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🇰🇿 Қазақша"),
			tgbotapi.NewKeyboardButton("🇷🇺 Русский"),
		),
	)
	kb.ResizeKeyboard = true
	return kb
}

func Menu(items []string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, l := range items {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(l)))
	}
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	return kb
}
