package util

import (
	"crypto/sha1"
	"encoding/hex"
	"os"

	"github.com/charmbracelet/log"
)

var Log *log.Logger

func InitLog() {
	Log = log.New(os.Stderr)
	Log.SetLevel(log.DebugLevel)
}

func GetSHA1(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
