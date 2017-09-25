package gontpd

import (
	"log"
	"os"
)

const (
	NoSyncClock = 1 + iota
)

var (
	DebugFlag uint8
	Debug     *log.Logger
	Info      *log.Logger
	Warn      *log.Logger
	Error     *log.Logger
)

func init() {
	Debug = log.New(os.Stderr, "[DEBG]", log.LstdFlags|log.LstdFlags)
	Info = log.New(os.Stderr, "[INFO]", log.LstdFlags|log.LstdFlags)
	Warn = log.New(os.Stderr, "[WARN]", log.LstdFlags|log.LstdFlags)
	Error = log.New(os.Stderr, "[EROR]", log.LstdFlags|log.LstdFlags)
}
