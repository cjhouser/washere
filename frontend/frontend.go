package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/cjhouser/washere/models"
	"google.golang.org/grpc"
)

type signaturePageData struct {
	PageTitle          string
	SignatureResponses []models.SignatureResponse
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
		var opts []grpc.DialOption
		conn, err := grpc.Dial("localhost:8081", opts...)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		client := models.NewSignatureClient(conn)
		signatureRequest := &models.SignatureRequest{}
		signatureResponses := []models.SignatureResponse{}
		stream, err := client.Get(context.Background(), signatureRequest)
		if err != nil {
			log.Println(err)
			return
		}
		for {
			signatureResponse, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
				return
			}
			signatureResponses = append(signatureResponses, signatureResponse)
		}
		template := template.Must(template.ParseFiles("index.html"))
		data := signaturePageData{
			PageTitle:          "washere",
			SignatureResponses: signatureResponses,
		}
		err = template.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
		log.Println(http.StatusOK, "GET", "/")
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
		log.Println(http.StatusNotFound)
	}
}
