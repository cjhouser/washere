package main

import (
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/cjhouser/washere/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type server struct {
	readerConnection *grpc.ClientConn
	template         *template.Template
}

func (s server) home(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		client := models.NewSignatureClient(s.readerConnection)
		signatures := []string{}
		stream, err := client.Get(context.Background(), &models.SignatureRequest{})
		if err != nil {
			log.Println("ERROR:", "failed to initialize message stream", err)
			return
		}
		for signatureResponse, err := stream.Recv(); err != io.EOF; {
			if err != nil {
				log.Println("ERROR:", "error receiving from message stream", err)
				return
			} else {
				signatures = append(signatures, signatureResponse.GetSignature())
			}
		}
		data := struct {
			PageTitle  string
			Signatures []string
		}{
			PageTitle:  "washere",
			Signatures: signatures,
		}
		if s.template.Execute(w, data) != nil {
			log.Println("ERROR:", "template generation error", err)
			return
		}
		log.Println("INFO:", r.Method, r.RequestURI, http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "not found"}`))
		log.Println("ERROR:", r.Method, r.RequestURI, http.StatusNotFound)
	}
}

func main() {
	listenSocket := os.Getenv("LISTEN_SOCKET")
	readerSocket := os.Getenv("READER_SOCKET")
	// Open a reusable connection to the reader service
	readerConnection, err := grpc.Dial(readerSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("ERROR:", "grpc dailing error", err)
		return
	}
	defer readerConnection.Close()
	server := server{
		readerConnection,
		template.Must(template.ParseFiles("index.html")),
	}
	http.HandleFunc("/", server.home)
	log.Println("INFO: listening on", listenSocket)
	http.ListenAndServe(listenSocket, nil)
}
