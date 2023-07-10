package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	listenSocket := os.Getenv("LISTEN_SOCKET")
	static := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static", static))
	log.Println("I: listening on", listenSocket)
	if err := http.ListenAndServe(listenSocket, nil); err != nil {
		log.Fatalln("F: listend and server failure", err)
	}
}
