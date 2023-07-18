package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type signature struct {
	Id   uint64
	Text string
}

type server struct {
	pool *pgxpool.Pool
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

		rows, err := s.pool.Query(r.Context(), fmt.Sprintf("SELECT * FROM signatures ORDER BY id LIMIT 10 OFFSET %d;", page*10))
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
			if err := rows.Scan(&id, &text); err != nil {
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
		if _, err := w.Write(data); err != nil {
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
	newRelicLicense := os.Getenv("NEW_RELIC_LICENSE")
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("signature-reader"),
		newrelic.ConfigLicense(newRelicLicense),
		newrelic.ConfigAppLogForwardingEnabled(false),
	)
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("F: failed to create database connection pool")
	}
	defer pool.Close()
	serverInstance := server{
		pool,
	}
	log.Println("I: database connection established")
	http.HandleFunc(newrelic.WrapHandleFunc(app, "/signatures", serverInstance.getSignatures))
	log.Println("I: listening on", listenSocket)
	if err := http.ListenAndServe(listenSocket, nil); err != nil {
		log.Fatalln("F: listen and server failure", err)
	}
}
