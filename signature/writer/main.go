package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/nsqio/go-nsq"
)

type handler struct {
	databasePool *pgxpool.Pool
	newRelic     *newrelic.Application
}

// HandleMessage implements the Handler interface.
func (h handler) HandleMessage(m *nsq.Message) error {
	newRelicTransaction := h.newRelic.StartTransaction("consume-and-write")
	defer newRelicTransaction.End()
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		// In this case, a message with an empty body is simply ignored/discarded.
		return nil
	}
	row := h.databasePool.QueryRow(context.TODO(), fmt.Sprintf("INSERT INTO signatures (signature) VALUES ('%s') RETURNING id;", m.Body))
	err := row.Scan()
	if err != nil {
		log.Println("E: scan failure", err)
	}
	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return err
}

func main() {
	newRelicLicense := os.Getenv("NEW_RELIC_LICENSE")
	newRelic, err := newrelic.NewApplication(
		newrelic.ConfigAppName("signature"),
		newrelic.ConfigLicense(newRelicLicense),
		newrelic.ConfigAppLogForwardingEnabled(false),
	)
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("F: failed to create connection pool", err)
	}
	defer pool.Close()

	row := pool.QueryRow(context.Background(), "CREATE TABLE IF NOT EXISTS signatures (id BIGSERIAL PRIMARY KEY, signature TEXT NOT NULL);")
	if err := row.Scan(); err != nil {
		if err != pgx.ErrNoRows {
			log.Fatalln("F: failed to create signatures table", err)
		}
	}

	consumer, err := nsq.NewConsumer("new-signatures", "database", nsq.NewConfig())
	if err != nil {
		log.Fatalln("F: failed to create new-signatures consumer", err)
	}
	defer consumer.Stop()

	consumer.AddHandler(handler{pool, newRelic})
	if err := consumer.ConnectToNSQLookupd(os.Getenv("NSQLOOKUPD_URL")); err != nil {
		log.Fatalln("F: failed to connect to nsqlookupd", err)
	}

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
