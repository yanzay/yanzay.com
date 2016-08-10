package main

import (
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":4200", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	response := `{"status": "awesome"}`
	w.Write([]byte(response))
}
