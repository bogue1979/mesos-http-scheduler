package main

import (
	"io"
	"net/http"
)

func root(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "<a href=\"/health\">Found</a>.")
}

func health(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "healthy")
}
