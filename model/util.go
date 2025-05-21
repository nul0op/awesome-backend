package model

import (
	"os"

	"github.com/charmbracelet/log"
)

var Log *log.Logger

func InitLog() {
	Log = log.New(os.Stderr)
	Log.SetLevel(log.DebugLevel)
}
