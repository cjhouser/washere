package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
)

type signature struct {
	id   uint64
	text string
}

type server struct {
	databaseConnection *pgx.Conn
	context            context.Context
}

func (s server) getSignatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		page := r.URL.Query().Get("page")
		resp, err := s.databaseConnection.Query(s.context, fmt.Sprintf("SELECT * FROM signatures ORDER BY id WHERE id=%s LIMIT 10;", page))
		if err != nil {
			log.Println("E: failed to select from database", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal server error"}`))
			return
		}
		signatures := []signature{}
		for resp.Next() {
			var id uint64
			var text string
			err = resp.Scan(&id, &text)
			if err != nil {
				log.Println("E: scan failure", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"internal server error"}`))
				return
			}
			signatures = append(signatures, signature{id, text})
		}
		data, err := json.Marshal(signatures)
		if err != nil {
			log.Println("E: marshal failure", err)
		}
		w.Write(data)
	default:
		log.Println("E: received non-GET request")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"message":"method not allowed"}`))
		return
	}
}

func main() {
	listenSocket := os.Getenv("LISTEN_SOCKET")
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("E: unable to connect to database", err)
	}
	defer conn.Close(context.Background())
	log.Println("I: database connection established")
	serverInstance := server{
		conn,
		context.Background(),
	}
	http.HandleFunc("/signatures", serverInstance.getSignatures)
	log.Println("I: listening on", listenSocket)
	http.ListenAndServe(listenSocket, nil)
}
