package coordinator

import (
	"bytes"
	"distributed/dto"
	"distributed/qutils"
	"encoding/gob"
	"github.com/streadway/amqp"
	"fmt"
)

type WebappConsumer struct {
//	er      EventRaiser
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources []string
}

func NewWebappConsumer() *WebappConsumer {
	wc := WebappConsumer{
	//	er: er,
	}


	wc.conn, wc.ch = qutils.GetChannel(url)
	// ? for what
	//qutils.GetQueue(qutils.PersistReadingsQueue, wc.ch, false)

	// publish data to WebappReadingsExchange for web app to consume
	qutils.CreateExchange(wc.ch, qutils.WebappReadingsExchange, "fanout")
	// publish sensor name(data source) to web apps
	qutils.CreateExchange(wc.ch, qutils.WebappSourceExchange, "fanout")

	go wc.ListenForDiscoveryRequests()

	// wc.er.AddListener("DataSourceDiscovered",
	// 	func(eventData interface{}) {
	// 		wc.SubscribeToDataEvent(eventData.(string))
	// 	})

	return &wc
}

func (wc *WebappConsumer) ListenForDiscoveryRequests() {
	q := qutils.GetQueue(qutils.WebappDiscoveryQueue, wc.ch, false)
	msgs, _ := wc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)
	fmt.Println("ListenForDiscoveryRequests called")
	for range msgs {
		fmt.Println("discovery request received from web app!")
		for _, src := range wc.sources {

			wc.SendMessageSource(src)
		}
	}
}

func (wc *WebappConsumer) SendMessageSource(src string) {
	err := wc.ch.Publish(
		qutils.WebappSourceExchange, //exchange string,
		"",    //key string,
		false, //mandatory bool,
		false, //immediate bool,
		amqp.Publishing{Body: []byte(src)}) //msg amqp.Publishing)
		qutils.FailOnError(err, "publish source failed")

		fmt.Println("published source: ", src)

}

// publish data to WebappReadingsExchange, fanout
func (wc *WebappConsumer) PublishData(eventData interface{})  {
	ed := eventData.(EventData)
	sm := dto.SensorMessage{
		Name:      ed.Name,
		Value:     ed.Value,
		Timestamp: ed.Timestamp,
	}
	buf := new(bytes.Buffer)
	buf.Reset()

	enc := gob.NewEncoder(buf)
	enc.Encode(sm)

	msg := amqp.Publishing{
		Body: buf.Bytes(),
	}

	wc.ch.Publish(
		qutils.WebappReadingsExchange, //exchange string,
		"", //routing key string,
		false, //mandatory bool,
		false, //immediate bool,
		msg)   //msg amqp.Publishing)
}
//
//
// func (wc *WebappConsumer) SubscribeToDataEvent(eventName string) {
// 	for _, v := range wc.sources {
// 		if v == eventName {
// 			return
// 		}
// 	}
//
// 	wc.sources = append(wc.sources, eventName)
//
// 	wc.SendMessageSource(eventName)
//
// 	wc.er.AddListener("MessageReceived_"+eventName,
// 		func(eventData interface{}) {
// 			ed := eventData.(EventData)
// 			sm := dto.SensorMessage{
// 				Name:      ed.Name,
// 				Value:     ed.Value,
// 				Timestamp: ed.Timestamp,
// 			}
//
// 			buf := new(bytes.Buffer)
//
// 			enc := gob.NewEncoder(buf)
// 			enc.Encode(sm)
//
// 			msg := amqp.Publishing{
// 				Body: buf.Bytes(),
// 			}
//
// 			wc.ch.Publish(
// 				qutils.WebappReadingsExchange, //exchange string,
// 				"",    //key string,
// 				false, //mandatory bool,
// 				false, //immediate bool,
// 				msg)   //msg amqp.Publishing)
// 		})
// }
