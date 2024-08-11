package main

import (
	"context"
	"log"
	"time"

	"github.com/AndreAppolariFilho/eventdrivenrabbit/internal"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

const CONNECT_TLS bool = false

func main() {
	var conn *amqp091.Connection
	var err error
	if CONNECT_TLS {
		conn, err = internal.ConnectRabbitMQTLS("andre", "123456", "localhost:5672", "customers", "C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\ca_certificate.pem",
			"C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\client_LAPTOP-L5PLHJG8_certificate.pem",
			"C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\client_LAPTOP-L5PLHJG8_key.pem")

		if err != nil {
			panic(err)
		}
	} else {
		conn, err = internal.ConnectRabbitMQ("andre", "123456", "localhost:5672", "customers")
		if err != nil {
			panic(err)
		}
	}
	defer conn.Close()

	var publishConn *amqp091.Connection
	if CONNECT_TLS {
		publishConn, err = internal.ConnectRabbitMQTLS("andre", "123456", "localhost:5672", "customers", "C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\ca_certificate.pem",
			"C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\client_LAPTOP-L5PLHJG8_certificate.pem",
			"C:\\Users\\andre\\Documents\\dev\\event-driven-rmq\\tls-gen\\basic\\result\\client_LAPTOP-L5PLHJG8_key.pem")

		if err != nil {
			panic(err)
		}
	} else {
		publishConn, err = internal.ConnectRabbitMQ("andre", "123456", "localhost:5672", "customers")
		if err != nil {
			panic(err)
		}
	}
	defer publishConn.Close()

	client, err := internal.NewRabbitMQClient(conn)

	if err != nil {
		panic(err)
	}

	defer client.Close()

	publishClient, err := internal.NewRabbitMQClient(publishConn)
	if err != nil {
		panic(err)
	}
	defer publishClient.Close()

	queue, err := client.CreateQueue("", true, true)

	if err != nil {
		panic(err)
	}

	if err := client.CreateBinding(queue.Name, "", "customer_events"); err != nil {
		panic(err)
	}
	messageBus, err := client.Consume(queue.Name, "email-service", false)

	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	var blocking chan struct{}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)

	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	//Apply a hard limit on the server
	if err := client.ApplyQos(10, 0, true); err != nil {
		panic(err)
	}
	//errgroup allows to set concurrent tasks
	g.SetLimit(10)

	go func() {
		for message := range messageBus {
			// Spawn a worker
			msg := message
			g.Go(func() error {
				log.Printf("New Message: %v", msg)
				time.Sleep(10 * time.Second)
				if err := msg.Ack(false); err != nil {
					log.Println("Ack message failed")
					return err
				}

				if err := publishClient.Send(ctx, "customer_callbacks", msg.ReplyTo, amqp091.Publishing{
					ContentType:   "text/plain",
					DeliveryMode:  amqp091.Persistent,
					Body:          []byte("RPC Complete"),
					CorrelationId: msg.CorrelationId,
				}); err != nil {
					panic(err)
				}
				log.Printf("Acknowledge message %s\n", msg.MessageId)
				return nil
			})
		}
	}()
	log.Println("Consuming, use CTRL+C to exit")
	<-blocking
}
