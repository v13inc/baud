package utils

import (
	"log"
	"net/http"
)

func Error(w http.ResponseWriter, r *http.Request, message string, err error) {
	log.Println("ERROR: ", message)
	log.Println(err)
	http.Error(w, message, http.StatusInternalServerError)
}

func MethodNotSupported(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.ErrNotSupported.Error(), http.StatusMethodNotAllowed)
}
