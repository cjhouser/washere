package main

import (
	"log"
	"net/http"

	"github.com/cjhouser/washere/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	http.HandleFunc("/signature", createSignature)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createSignature(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// check for empty data. no empty data allowed
		signature := models.Signature{Text: "charles was here", CreatedAt: timestamppb.Now()}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "OK"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}
