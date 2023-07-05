package main

import (
	"log"
	"net/http"
	"os"

	"github.com/nsqio/go-nsq"
)

type server struct {
	producer *nsq.Producer
}

func (s server) handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if r.ParseForm() != nil {
			log.Println("E: failed to parse form")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		signature := r.PostForm.Get("signature")
		if s.producer.Publish("new-signatures", []byte(signature)) != nil {
			log.Println("E: failed to publish signature")
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "created"}`))
		log.Println("I: created signature:", signature)
	default:
		log.Println("E: received non-GET request")
		http.Error(w, "internal server error", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	listenSocket := os.Getenv("LISTEN_SOCKET")
	nsqdSocket := os.Getenv("NSQD_SOCKET")
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqdSocket, config)
	if err != nil {
		log.Fatal("E: failed creating nsq producer", err)
	}
	serverInstance := server{
		producer,
	}
	http.HandleFunc("/signatures", serverInstance.handler)
	log.Println("I: listening on", listenSocket)
	http.ListenAndServe(listenSocket, nil)
}
