package slogger

import (
	"log"
	"net/http"
	"os"
)

func LoggerHandler(r *http.Request) {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
}
