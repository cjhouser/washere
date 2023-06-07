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

	_, err := h.databaseConnection.Query(h.context, fmt.Sprintf("INSERT INTO signatures (signature) VALUES (%s);", m.Body))
	if err != nil {
		log.Println("failed to insert into database", err)
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

	_, err = conn.Query(context.Background(), "CREATE TABLE IF NOT EXISTS signatures VALUES (id INTEGER PRIMARY KEY AUTO INCREMENT, signature TEXT NOT NULL);")
	if err != nil {
		log.Fatalln("error creating table", err)
	}

	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("new-signatures", "database", config)
	if err != nil {
		log.Fatal(err)
	}
	consumer.AddHandler(&newSignatureHandler{context.Background(), conn})

	err = consumer.ConnectToNSQLookupd("localhost:4161")
	if err != nil {
		log.Fatal(err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	consumer.Stop()
}
