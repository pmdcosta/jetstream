package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/nats.go"
)

type configuration struct {
	NatsHost string `envconfig:"NATS_HOST" default:"localhost:4222"`
	Topic    string `envconfig:"TOPIC" default:"ORDERS.new"`
}

func main() {
	// create application context.
	appCtx, cancel := context.WithCancel(context.Background())
	go done(appCtx, cancel)

	// load app configuration variables.
	var conf configuration
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal("failed to load ENV variables: ", err.Error())
	}

	// start producing messages.
	log.Println("started producing messages")
	if err := runProducer(appCtx, conf.NatsHost, conf.Topic); err != nil {
		log.Fatal("failed producing messages: ", err.Error())
	}
}

func done(ctx context.Context, cancel func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer log.Println("started shutting down")
	select {
	case <-ctx.Done():
	case <-sigChan:
		cancel()
	}
}

func runProducer(ctx context.Context, host, topic string) error {
	// connect to NATS.
	nc, err := nats.Connect(host)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// create JetStream Context.
	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// publish sync message.
	_, err = js.Publish(topic, []byte("Started producing messages"))
	if err != nil {
		return fmt.Errorf("failed to publish sync message: %w", err)
	}

	// publish async messages.
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			// wait for all messages to be acked before closing.
			select {
			case <-js.PublishAsyncComplete():
				log.Println("all messages have been published")
			}
			return nil
		case <-ticker.C:
			if _, err = js.PublishAsync(topic, []byte(time.Now().Format(time.Stamp))); err != nil {
				return fmt.Errorf("failed to publish async message: %w", err)
			}
			log.Println("messaged produced")
		}
	}
}
