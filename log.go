package gontpd

import (
	"log"
	"os"
)

var (
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func init() {
	Info = log.New(os.Stderr, "[INFO] ", log.LstdFlags|log.LstdFlags)
	Warn = log.New(os.Stderr, "[WARN] ", log.LstdFlags|log.LstdFlags)
	Error = log.New(os.Stderr, "[EROR] ", log.LstdFlags|log.LstdFlags)
}
