package main

import (
	"awesome-portal/backend/model"
	"awesome-portal/backend/server"

	"github.com/joho/godotenv"
)

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

func main() {
	_ = godotenv.Load()

	model.InitLog()

	model.Log.Info("Connecting to database")
	model.Connect()

	model.Log.Info("Launching indexing")
	// indexer.Index(AW_ROOT, 0)

	model.Log.Info("Starting web server")
	server.StartServer()
}
