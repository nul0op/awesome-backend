package main

import (
	"awesome-portal/backend/server"

	"github.com/joho/godotenv"

	model "awesome-portal/backend/db"
)

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

func main() {
	_ = godotenv.Load()

	model.Connect()

	// indexer.Index(AW_ROOT, 0)
	// fmt.Print(json)
	server.Start_server()
}
