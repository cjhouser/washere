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

func getSignatures(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			log.Println("E: failed to covert query to integer", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Println("E: failed to connect to database", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close(ctx)
		rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM signatures ORDER BY id LIMIT 10 OFFSET %d;", page*10))
		if err != nil {
			log.Println("E: failed to select from database", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		signatures := []signature{}
		for rows.Next() {
			var id uint64
			var text string
			if rows.Scan(&id, &text) != nil {
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
	log.Println("I: database connection established")
	http.HandleFunc("/signatures", getSignatures)
	log.Println("I: listening on", listenSocket)
	http.ListenAndServe(listenSocket, nil)
}
