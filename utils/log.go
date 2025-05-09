package utils

import "log"

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}
