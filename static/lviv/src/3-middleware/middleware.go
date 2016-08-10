package main

import (
	"log"
	"net/http"
	"time"
)

func logger(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler(w, r)
		duration := float64(time.Since(start)) / float64(time.Millisecond)
		log.Printf("[%.2fms] %s %s", duration, r.Method, r.RequestURI)
	}
}
