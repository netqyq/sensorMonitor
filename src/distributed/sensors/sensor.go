package main

import (
	"bytes"
	"distributed/dto"
	"distributed/qutils"
	"encoding/gob"
	"flag"
	"github.com/streadway/amqp"
	"log"
	"math/rand"
	"strconv"
	"time"
	"fmt"
)

var url = "amqp://guest:guest@localhost:5672"

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var value float64
var nom float64
var err error

func main() {
	flag.Parse()
	value = r.Float64()*(*max-*min) + *min
	nom = (*max-*min)/2 + *min

	conn, ch := qutils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	// 1. publish sensor name, to inform coordinators
	qutils.CreateExchange(ch, qutils.SensorOnline, "fanout")
	publishQueueName(ch)

	// 2. SensorDiscoveryExchange Setup and listen for new
	//    discovery request
	qutils.CreateExchange(ch, qutils.SensorDiscovery, "fanout")
	discoveryQueue := qutils.GetQueue("", ch, true)
	//fmt.Println("discoveryQueue Name: " + discoveryQueue.Name)
	err = ch.QueueBind(
		discoveryQueue.Name, //name string,
		"",                  //key string,
		qutils.SensorDiscovery, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)
  qutils.FailOnError(err, "Failed to bind discovery Q")
	go listenForDiscoverRequests(discoveryQueue.Name, ch)


	// 3. dataQueue Setup and start send messages
	fmt.Println("dataQueue Name: " + *name)

	// fanout dataQueue init
	//qutils.CreateExchange(ch, *name, "fanout")

	// direct dataQueue init
	q, err := ch.QueueDeclare(
		*name, // name
		false,         // durable
		true,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	qutils.FailOnError(err, "Failed to declare a queue: "+*name)


	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")

	signal := time.Tick(dur)

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	for range signal {
		calcValue()
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}

		buf.Reset()
		enc = gob.NewEncoder(buf)
		enc.Encode(reading)

		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		// direct
		// ch.Publish(
		// 	"",  //exchange string,
		// 	dataQueue.Name,	//key string,
		// 	false,          //mandatory bool,
		// 	false,          //immediate bool,
		// 	msg)            //msg amqp.Publishing)

		//fanout publish
		// ch.Publish(
		// 	*name,  //exchange string,
		// 	"",	//key string,
		// 	false,          //mandatory bool,
		// 	false,          //immediate bool,
		// 	msg)            //msg amqp.Publishing)

		// direct publish
		ch.Publish(
				"",  //exchange string,
				q.Name,	//key string,
				false,          //mandatory bool,
				false,          //immediate bool,
				msg)            //msg amqp.Publishing)


		log.Printf(" %s, Value: %v\n", *name, value)
	}
}

func listenForDiscoverRequests(name string, ch *amqp.Channel) {

	msgs, _ := ch.Consume(
		name,  //queue string,
		"",    //consumer string,
		true,  //autoAck bool,
		false, //exclusive bool,
		false, //noLocal bool,
		false, //noWait bool,
		nil)   //args amqp.Table)
	for range msgs {
		log.Println("received discovery request")
		publishQueueName(ch)
	}
}

func publishQueueName(ch *amqp.Channel) {
	msg := amqp.Publishing{Body: []byte(*name)}
	ch.Publish(
		qutils.SensorOnline, //exchange string,
		"",    //key string,
		false,        //mandatory bool,
		false,        //immediate bool,
		msg)          //msg amqp.Publishing)
}

func calcValue() {
	var maxStep, minStep float64

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min)
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += r.Float64()*(maxStep-minStep) + minStep
}
