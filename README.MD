# Instructions to run

## Without SSL

First it's necessary to install the rabbitmq in order to configure one server, the easiest way is with docker with the following command

`docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3.11-management`

The port 5672 is used for connections with the message brokers

The port 15672 is used for accesing the UI.

This command will set an user for the rabbitmq, make sure to use this user in the code.
`docker exec rabbitmq rabbitmqctl add_user username secret`

It'll be necessary to give this user administrator rights
`docker exec rabbitmq rabbitmqctl set_user_tags username administrator`

It's safer to delete the guest user that comes with the instalation

`docker exec rabbitmq rabbitmqctl delete_user guest`

Let's create the vhost for our project

`docker exec rabbitmq rabbitmqctl add_vhost customers`

And of course we have to give our user the permission for this vhost

`docker exec rabbitmq rabbitmqctl set_permissions -p customers username ".*" ".*" ".*"`

Now that some things are set, it's necessary to install the go package

`go get github.com/rabbitmq/amqp091-go`

Let's create the exchanges used in this project

`docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_events type=fanout -u percy -p secret durable=true`

`docker exec rabbitmq rabbitmqadmin declare exchange --vhost=customers name=customer_callbacks type=direct -u percy -p secret durable=true`

Now it's time to set the topics
``
`docker exec rabbitmq rabbitmqctl set_topic_permissions -p customers percy customer_callbacks ".*" ".*"`

Now it's just to run the producer and consumer
`go run cmd/producer/main.go`
`go run cmd/consumer/main.go`

## With tls

First we have to generate SSL keys if you have already you can skip this part

`git clone https://github.com/rabbitmq/tls-gen tls-gen
cd tls-gen/basic
make PASSWORD=
make verify`

Giving the permission to the ssl folder

`sudo chmod 644 tls-gen/basic/result/*`

Running the rabbitmq docker with the configuration

`docker run -d --name rabbitmq -v "$(pwd)"/rabbitmq_definitions.json:/etc/rabbitmq/rabbitmq_definitions.json:ro -v "$(pwd)"/rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro -v "$(pwd)"/tls-gen/basic/result:/certs -p 5671:5671 -p 15672:15672 rabbitmq:3.11-management`


Now it's just to run the producer and consumer
`go run cmd/producer/main.go`
`go run cmd/consumer/main.go`