package config

import "os"

type Config struct {
	Token       string
	APIBase     string
	NLPBase     string
	DatabaseURL string
}

func FromEnv() Config {
	return Config{
		Token:   os.Getenv("TELEGRAM_TOKEN"),
		APIBase: os.Getenv("CONTENT_API_URL"),
		NLPBase: os.Getenv("NLP_API_URL"),
		//DatabaseURL: os.Getenv("DATABASE_URL"),
	}
}
