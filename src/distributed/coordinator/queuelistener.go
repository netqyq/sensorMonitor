package coordinator

import (
	"bytes"
	"distributed/dto"
	"distributed/qutils"
	"encoding/gob"
	"fmt"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type QueueListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery		// soure of sensor's channels
	//sources []string
	ea      *EventAggregator
}

func NewQueueListener(ea *EventAggregator) *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
		ea:      ea,
	}
	ql.conn, ql.ch = qutils.GetChannel(url)
	return &ql
}

func (ql *QueueListener) DiscoverSensors() {
	qutils.CreateExchange(ql.ch, qutils.SensorDiscovery, "fanout")
	ql.ch.Publish(
		qutils.SensorDiscovery, //exchange string,
		"",                //key string,
		false,             //mandatory bool,
		false,             //immediate bool,
		amqp.Publishing{}) //msg amqp.Publishing)
	fmt.Println("Discovery Message Sent.")
}

func (ql *QueueListener) ListenForNewSource() {
	// 1. consume SensorOnline queue
	// 2. send discovery message to SensorDiscovery
	// 3. add discovered sensorName to
	// 4. consume the sensorName channel
	// 5. publish to inform the data consumers, down stream reader

	qutils.CreateExchange(ql.ch, qutils.SensorOnline,"fanout")
  onlineQueue := qutils.GetQueue("", ql.ch, true)
	ql.ch.QueueBind(onlineQueue.Name, // queue name
		                "",     // routing key
		                qutils.SensorOnline, // exchange
		                false,
		                nil)
	msgs, err := ql.ch.Consume(
											onlineQueue.Name, //queue string,
											"",     //consumer string,
											true,   //autoAck bool,
											false,  //exclusive bool,
											false,  //noLocal bool,
											false,  //noWait bool,
											nil)    //args amqp.Table)
	qutils.FailOnError(err, "Failed to Consume" + onlineQueue.Name)

	ql.DiscoverSensors()

	fmt.Println("listening for new sources")
	for msg := range msgs {
		// sensor name or its queue name
		fmt.Println("msg"+string(msg.Body))
		queueName := string(msg.Body)

		// new sensor discovered
		if ql.sources[queueName] == nil {
			// send message
			//ql.ea.PublishEvent("DataSourceDiscovered", queueName)
			//
			//ql.ea.DiscoverNewSource(queueName)

			// inform new sensor is comming

			// go func(){
			// 	ql.ea.sensors <- queueName
			// }()

			// sourceChan is source channel of one sensor's queue
			fmt.Println("before consume new source")
			qutils.CreateExchange(ql.ch, queueName, "fanout")
			dataQueue := qutils.GetQueue("", ql.ch, true)
			ql.ch.QueueBind(dataQueue.Name, // queue name
												"",     // routing key
												queueName, // exchange
												false,
												nil)
			sourceChan, err := ql.ch.Consume(
				dataQueue.Name, //queue string,
				"",               //consumer string,
				false,             //autoAck bool,
				false,            //exclusive bool,
				false,            //noLocal bool,
				false,            //noWait bool,
				nil)              //args amqp.Table)
			fmt.Println("sourceChan is: ", sourceChan)
			qutils.FailOnError(err, "consume new source failed")

			ql.sources[queueName] = sourceChan
			fmt.Println("new source received: "+ queueName)
			//go ql.TrigEvent(sourceChan)
			go ql.SendToPersistQueue(sourceChan)
			fmt.Println("after go")
		} else {
			// no new sensor
			fmt.Println("old source: "+ queueName)
		}
	}
}

// receive meassages from sensor,
func (ql *QueueListener) SendToPersistQueue(msgs <-chan amqp.Delivery)  {
	// receive meassage from one sensor
	fmt.Println("SendToPersistQueue Called")
	fmt.Println(msgs)
	for msg := range msgs {
		fmt.Println("in for")
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dto.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Received message: %v\n", sd)
		msg.Ack(false)
		// ed := EventData{
		// 	Name:      sd.Name,
		// 	Timestamp: sd.Timestamp,
		// 	Value:     sd.Value,
		// }
		// trigger the event

	}
}

// receive the message from sensor queue and trigger an event.
func (ql *QueueListener) TrigEvent(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dto.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Received message: %v\n", sd)

		ed := EventData{
			Name:      sd.Name,
			Timestamp: sd.Timestamp,
			Value:     sd.Value,
		}
		// trigger the event
		//fmt.Println(msg.RoutingKey)
		ql.ea.PublishEvent("MessageReceived_"+msg.RoutingKey, ed)
	}
}
