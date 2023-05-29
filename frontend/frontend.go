package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/cjhouser/washere/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type signaturePageData struct {
	PageTitle  string
	Signatures []models.Signature
}

func main() {
	http.HandleFunc("/", serveHome)
	log.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		serveTemplate(w)
		log.Println(http.StatusOK, "GET", "/")
	} else {
		w.WriteHeader(http.StatusNotFound)
		serveTemplate(w)
		log.Println(http.StatusNotFound)
	}
}

func serveTemplate(w http.ResponseWriter) {
	template := template.Must(template.ParseFiles("index.html"))
	data := signaturePageData{
		PageTitle: "washere",
		Signatures: []models.Signature{
			{Text: "charles was here", CreatedAt: timestamppb.Now()},
			{Text: "i was not here", CreatedAt: timestamppb.Now()},
		},
	}
	err := template.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}
