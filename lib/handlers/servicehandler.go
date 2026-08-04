package handlers

import "net/http"

const version = "0.1.0"

func Pong(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("PONG"))
}