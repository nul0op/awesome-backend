package main

import (
	"awesome-portal/backend/indexer"
	"awesome-portal/backend/model"
	"awesome-portal/backend/server"
	"awesome-portal/backend/util"

	"github.com/joho/godotenv"
)

const GH_API_URL = "https://api.github.com"
const AW_ROOT = "https://github.com/sindresorhus/awesome"

func main() {
	_ = godotenv.Load()

	util.InitLog()

	util.Log.Info("Connecting to database")
	model.Connect()

	util.Log.Info("Launching indexing")
	// indexer.Index(AW_ROOT, 0, 0)
	indexer.Index("https://github.com/drmonkeyninja/awesome-textpattern#readme", 0, 0)

	util.Log.Info("Starting web server")
	server.StartServer()
}
