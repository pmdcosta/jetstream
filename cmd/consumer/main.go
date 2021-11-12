package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/nats.go"
)

type configuration struct {
	NatsHost  string `envconfig:"NATS_HOST" default:"localhost:4222"`
	Topic     string `envconfig:"TOPIC" default:"ORDERS.new"`
	Consumer  string `envconfig:"CONSUMER" default:"new-consumer"`
	BatchSize int    `envconfig:"BATCH_SIZE" default:"10"`
}

func main() {
	// create application context.
	appCtx, cancel := context.WithCancel(context.Background())
	go done(appCtx, cancel)

	// load app configuration variables.
	var conf configuration
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal("failed to load ENV variables:", err.Error())
	}

	// start producing messages.
	log.Println("started consuming messages")
	if err := runConsumer(appCtx, conf.NatsHost, conf.Topic, conf.Consumer, conf.BatchSize); err != nil {
		log.Fatal("failed consuming messages:", err.Error())
	}
}

func done(ctx context.Context, cancel func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer log.Println("shutting down")
	select {
	case <-ctx.Done():
	case <-sigChan:
		cancel()
	}
}

func runConsumer(ctx context.Context, host, topic, consumer string, size int) error {
	// connect to NATS.
	nc, err := nats.Connect(host)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// create JetStream Context.
	js, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// create batch consumer.
	sub, err := js.PullSubscribe(topic, consumer)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}
	defer sub.Drain()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// fetch messages.
			msgs, err := sub.Fetch(size)
			if err != nil && err != nats.ErrTimeout {
				return fmt.Errorf("failed to consume batch: %w", err)
			}
			if err := processBatch(msgs); err != nil {
				return err
			}
		}
	}
}

func processBatch(msgs []*nats.Msg) error {
	for _, msg := range msgs {
		log.Println("received message: " + string(msg.Data))
		if err := msg.Ack(); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
	}
	return nil
}
