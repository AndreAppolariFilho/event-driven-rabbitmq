package internal

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitClient struct {
	//Connection used by the client
	conn *amqp.Connection
	//Channel is used to process / Send messages
	ch *amqp.Channel
}

func ConnectRabbitMQTLS(username, password, host, vhost, caCert, clientCert, clientKey string) (*amqp.Connection, error) {
	ca, err := os.ReadFile(caCert)
	if err != nil {
		return nil, err
	}

	//Load key pair
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)

	if err != nil {
		return nil, err
	}

	// Add the RootCA to the cert pool

	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(ca)

	tlsConfg := &tls.Config{RootCAs: rootCAs, Certificates: []tls.Certificate{cert}}
	return amqp.DialTLS(fmt.Sprintf("amqps://%s:%s@%s/%s", username, password, host, vhost), tlsConfg)
}

func ConnectRabbitMQ(username, password, host, vhost string) (*amqp.Connection, error) {
	// Add the RootCA to the cert pool
	return amqp.Dial(fmt.Sprintf("amqps://%s:%s@%s/%s", username, password, host, vhost))
}

func NewRabbitMQClient(conn *amqp.Connection) (RabbitClient, error) {
	ch, err := conn.Channel()
	if err != nil {
		return RabbitClient{}, err
	}
	if err := ch.Confirm(false); err != nil {
		return RabbitClient{}, err
	}
	return RabbitClient{
		conn: conn,
		ch:   ch,
	}, nil
}

func (rc RabbitClient) Close() error {
	return rc.ch.Close()
}

func (rc RabbitClient) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
	g, err := rc.ch.QueueDeclare(queueName, durable, autodelete, false, false, nil)
	if err != nil {
		return amqp.Queue{}, nil
	}
	return g, err
}

func (rc RabbitClient) CreateBinding(name, binding, exchange string) error {
	//leaving nowait to false will make the channel returns an error if it fails to bind
	return rc.ch.QueueBind(name, binding, exchange, false, nil)
}

func (rc RabbitClient) Send(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	confirmation, err := rc.ch.PublishWithDeferredConfirmWithContext(
		ctx,
		exchange,
		routingKey,
		//Is used to determine if an error should be returned upon failure
		true,
		false,
		options,
	)
	if err != nil {
		return err
	}
	log.Println(confirmation.Wait())
	return nil
}

func (rc RabbitClient) Consume(queue, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	return rc.ch.Consume(queue, consumer, autoAck, false, false, false, nil)
}

// ApplyQos
// prefetch count - an integer on how many unacknowlodged messages the server can send
// prefetch size - an integer on how many bytes a queue can have
// global - Determines if the rule can be applied globally or not
func (rc RabbitClient) ApplyQos(count, size int, global bool) error {
	return rc.ch.Qos(count, size, global)
}
