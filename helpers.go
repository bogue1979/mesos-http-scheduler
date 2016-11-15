package main

import (
	"io"
	"log"
	"net/http"
)

func debugLog(s string) {
	if *debug {
		log.Println(s)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<a href=\"/health\">Found</a>.")
}

func health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "healthy")
}
