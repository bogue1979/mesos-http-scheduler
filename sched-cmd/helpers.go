package main

import "log"

func debugLog(s string) {
	if *debug {
		log.Println(s)
	}
}
