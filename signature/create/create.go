package main

import (
	"log"
	"net/http"

	"github.com/nsqio/go-nsq"
)

func main() {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("192.168.0.252:31000", config)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "internal server error"}`))
			}
			signature := r.PostForm.Get("signature")
			err = producer.Publish("new-signatures", []byte(signature))
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "internal server error"}`))
			} else {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"message": "OK"}`))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"message": "not found"}`))
		}

	})
	log.Println("signature create - listening on 8082")
	err = http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal(err)
	}
}
