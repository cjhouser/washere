package main

import (
	"log"
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/nsqio/go-nsq"
)

type server struct {
	producer *nsq.Producer
}

func (s server) handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			log.Println("E: failed to parse form", err)
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
		if _, err := w.Write([]byte(`{"message": "created"}`)); err != nil {
			log.Println("E: failed to send response", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		log.Println("I: created signature:", signature)
	default:
		log.Println("E: received non-GET request")
		http.Error(w, "internal server error", http.StatusMethodNotAllowed)
		return
	}
}

func main() {
	listenSocket := os.Getenv("LISTEN_SOCKET")
	newRelicLicense := os.Getenv("NEW_RELIC_LICENSE")
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("signature"),
		newrelic.ConfigLicense(newRelicLicense),
		newrelic.ConfigAppLogForwardingEnabled(false),
	)
	if err != nil {
		log.Fatalln("F: failed to register New Relic agent", err)
	}
	nsqdSocket := os.Getenv("NSQD_SOCKET")
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(nsqdSocket, config)
	if err != nil {
		log.Fatalln("F: failed creating nsq producer", err)
	}
	serverInstance := server{
		producer,
	}
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/signatures/create", serverInstance.handler))
	log.Println("I: listening on", listenSocket)
	if err := http.ListenAndServe(listenSocket, nil); err != nil {
		log.Fatalln("F: listen and serve failure", err)
	}
}
