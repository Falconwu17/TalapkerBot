package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegramBot/bot/internal/client"
	"telegramBot/bot/internal/config"
	"telegramBot/bot/internal/handlers"
	"telegramBot/bot/internal/nlpclient"
)

func main() {
	cfg := config.FromEnv()
	if cfg.Token == "" || cfg.APIBase == "" || cfg.NLPBase == "" {
		log.Fatal("TELEGRAM_TOKEN, CONTENT_API_URL или NLP_API_URL пустые")
	}

	b, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}

	h := &handlers.Bot{
		API:   b,
		APICl: contentclient.New(cfg.APIBase),
		NLP:   nlpclient.New(cfg.NLPBase),
	}

	b.Request(tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "start", Description: "Запустить бота"},
		tgbotapi.BotCommand{Command: "help", Description: "Помощь"},
	))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)

	log.Printf("Bot @%s started", b.Self.UserName)
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down...")
			return
		case upd := <-updates:
			h.HandleMessage(upd)
		}
	}
}
