package gontpd

import "log"

func init() {
	if log.Flags() != 0 {
		log.SetPrefix("[GoNTPD] ")
	}
}
