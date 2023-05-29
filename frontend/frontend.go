package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/cjhouser/washere/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type signaturePageData struct {
	pageTitle  string
	signatures []models.Signature
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
		template := template.Must(template.ParseFiles("index.html"))
		data := signaturePageData{
			pageTitle: "washere",
			signatures: []models.Signature{
				{Text: "charles was here", CreatedAt: timestamppb.Now()},
			},
		}
		template.Execute(w, data)
		log.Println(http.StatusOK)
		http.ServeFile(w, r, "index.html")
	} else {
		log.Println(http.StatusNotFound)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
	}
}
