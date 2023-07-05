package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
)

type signature struct {
	Id   uint64
	Text string
}

type server struct {
	databaseConnection *pgx.Conn
	context            context.Context
}

func (s server) getSignatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			log.Println("E: failed to covert query to integer", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		resp, err := s.databaseConnection.Query(s.context, fmt.Sprintf("SELECT * FROM signatures ORDER BY id LIMIT 10 OFFSET %d;", page*10))
		if err != nil {
			log.Println("E: failed to select from database", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		signatures := []signature{}
		for resp.Next() {
			var id uint64
			var text string
			if resp.Scan(&id, &text) != nil {
				log.Println("E: scan failure", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			signatures = append(signatures, signature{id, text})
		}
		data, err := json.Marshal(signatures)
		if err != nil {
			log.Println("E: marshal failure", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			log.Println("E: write failure", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		log.Println("I: sending data:", signatures)
	default:
		log.Println("E: received non-GET request")
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
