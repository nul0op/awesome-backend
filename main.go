package main

import (
	"awesome-portal/backend/server"

	"github.com/joho/godotenv"

	db "awesome-portal/backend/database"
)

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

func main() {
	_ = godotenv.Load()

	db.Connect()

	// indexer.Index(AW_ROOT, 0)
	// fmt.Print(json)
	server.StartServer()
}
