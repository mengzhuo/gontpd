//+build debug

package gontpd

import "log"

const debug = true

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetPrefix("[DEBG] ")
}
