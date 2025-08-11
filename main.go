package main

import (
	"telegramBot/dbConn"
	"telegramBot/services"
)

func main() {
	dbConn.ConnectToDB()
	defer dbConn.Pool.Close()
	services.Bot()
}
