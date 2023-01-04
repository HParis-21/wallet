package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	if NewStart() {
		CreateNewAccount()
	}
	r.HandleFunc("/api/wallet/{address}/balance", getBalance).Methods("GET")
	r.HandleFunc("/api/transactions", getLast).Methods("GET")
	r.HandleFunc("/api/send", postSend).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
}
