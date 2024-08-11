package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AndreAppolariFilho/eventdrivenrabbit/internal"
	"github.com/rabbitmq/amqp091-go"
)

const CONNECT_TLS bool = false

func main() {
	var conn *amqp091.Connection
	var err error
	if CONNECT_TLS {
		conn, err = internal.ConnectRabbitMQTLS("andre", "123456", "localhost:5672", "customers",
			"ca_certificate.pem",
			"client_certificate.pem",
			"client_key.pem",
		)

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
	var consumeConn *amqp091.Connection
	if CONNECT_TLS {
		consumeConn, err = internal.ConnectRabbitMQTLS("andre", "123456", "localhost:5672", "customers",
			"ca_certificate.pem",
			"client_certificate.pem",
			"client_key.pem")
		if err != nil {
			panic(err)
		}
	} else {
		consumeConn, err = internal.ConnectRabbitMQ("andre", "123456", "localhost:5672", "customers")
		if err != nil {
			panic(err)
		}
	}
	defer consumeConn.Close()

	client, err := internal.NewRabbitMQClient(conn)

	if err != nil {
		panic(err)
	}

	defer client.Close()

	consumeClient, err := internal.NewRabbitMQClient(consumeConn)
	if err != nil {
		panic(err)
	}
	defer consumeClient.Close()
	queue, err := consumeClient.CreateQueue("", true, true)
	if err != nil {
		panic(err)
	}
	if err := consumeClient.CreateBinding(queue.Name, queue.Name, "customer_callbacks"); err != nil {
		panic(err)
	}

	messageBus, err := consumeClient.Consume(queue.Name, "customer-api", true)
	if err != nil {
		panic(err)
	}
	go func() {
		for message := range messageBus {
			log.Println("Message Callback %s\n", message.CorrelationId)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		if err := client.Send(ctx, "customer_events", "customers.created.us", amqp091.Publishing{
			ContentType:   "text/plain",
			ReplyTo:       queue.Name,
			CorrelationId: fmt.Sprintf("customer_created_%d", i),
			DeliveryMode:  amqp091.Persistent,
			Body:          []byte(`A cool message between services`),
		}); err != nil {
			panic(err)
		}

	}
	time.Sleep(10 * time.Second)

	log.Println(client)
	var blocking chan struct{}
	<-blocking
}
