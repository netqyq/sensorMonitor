package qutils

import (
	_ "fmt"
	"github.com/streadway/amqp"
	"log"
)

const PersistReadingsQueue = "PersistReading"
const WebappSourceExchange = "WebappSources"
const WebappReadingsExchange = "WebappReading"
const WebappDiscoveryQueue = "WebappDiscovery"

// receive sensor name, topic
const SensorOnline = "SensorOnline"

// for discovery message, topic
const SensorDiscovery = "SensorDiscovery"

func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	FailOnError(err, "Failed to establish connection to message broker")
	ch, err := conn.Channel()
	FailOnError(err, "Failed to get channel for connection")

	return conn, ch
}

func CreateExchange(ch *amqp.Channel, name, kind string)  {
	err := ch.ExchangeDeclare(name,
											kind, //kind
											false, //durable,
											false, //autoDelete,
											false, //internal,
											false, //noWait,
											nil) //args)
	FailOnError(err, "Failed to declare Exchange")
}

func GetQueue(name string, ch *amqp.Channel, autoDelete bool) *amqp.Queue {
	q, err := ch.QueueDeclare(
		name,       //name string,
		false,      //durable bool,
		autoDelete, //autoDelete bool,
		false,      //exclusive bool,
		false,      //noWait bool,
		nil)        //args amqp.Table)

	FailOnError(err, "Failed to declare queue")

	return &q
}

func Consume(ch *amqp.Channel, Name, kind string ) <-chan amqp.Delivery  {
	CreateExchange(ch, Name,"fanout")
	onlineQueue := GetQueue(Name, ch, true)
	ch.QueueBind(onlineQueue.Name, // queue name
										"",     // routing key
										Name, // exchange
										false,
										nil)
	msgs, err := ch.Consume(
		onlineQueue.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)
	FailOnError(err, "Consume Failed")
	return msgs
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
