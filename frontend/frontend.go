package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/cjhouser/washere/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type signaturePageData struct {
	PageTitle  string
	Signatures []string
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
		conn, err := grpc.Dial("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		client := models.NewSignatureClient(conn)
		signatureRequest := &models.SignatureRequest{}
		signatures := []string{}
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
			signatures = append(signatures, signatureResponse.GetSignature())
		}
		template := template.Must(template.ParseFiles("index.html"))
		data := signaturePageData{
			PageTitle:  "washere",
			Signatures: signatures,
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
