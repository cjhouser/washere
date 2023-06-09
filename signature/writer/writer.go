package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/nsqio/go-nsq"
)

type newSignatureHandler struct {
	context            context.Context
	databaseConnection *pgx.Conn
}

// HandleMessage implements the Handler interface.
func (h *newSignatureHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}

	resp, err := h.databaseConnection.Query(h.context, fmt.Sprintf("INSERT INTO signatures (signature) VALUES ('%s');", m.Body))
	if err != nil {
		log.Println("failed to insert into database", err)
	}

	for resp.Next() {
		err = resp.Scan()
		if err != nil {
			log.Println("scan failure", err)
		}
	}

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return err
}

func main() {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("unable to connect", err)
	}
	defer conn.Close(context.Background())

	resp, err := conn.Query(context.Background(), "CREATE TABLE IF NOT EXISTS signatures (id BIGSERIAL PRIMARY KEY, signature TEXT NOT NULL);")
	if err != nil {
		log.Fatalln("error creating table", err)
	}

	for resp.Next() {
		err = resp.Scan()
		if err != nil {
			log.Println("row failure when creating table", err)
		}
	}

	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("new-signatures", "database", config)
	if err != nil {
		log.Fatalln("failed to create new consumer", err)
	}
	consumer.AddHandler(&newSignatureHandler{context.Background(), conn})

	err = consumer.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_URL"))
	if err != nil {
		log.Fatalln("failed to connect to nsqlookupd", err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	consumer.Stop()
}
