package utils

import "log"

func Try(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Error:%v\n", err)
		}
	}()
	f()
}
