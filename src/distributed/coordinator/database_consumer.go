package coordinator

import (
	"bytes"
	"distributed/dto"
	"distributed/qutils"
	"encoding/gob"
	"github.com/streadway/amqp"
	"time"
	"fmt"
)

const maxRate = 5 * time.Second

type DatabaseConsumer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   *amqp.Queue
//	sources []string
	sources map[string]<-chan amqp.Delivery
}

func NewDatabaseConsumer() *DatabaseConsumer {
	dc := DatabaseConsumer{

	}

	dc.conn, dc.ch = qutils.GetChannel(url)
	dc.queue = qutils.GetQueue(qutils.PersistReadingsQueue,
															dc.ch, false)
	return &dc
}

func (dc *DatabaseConsumer) GetSources()  {
	// fmt.Println("sources: %v", dc.sources)
	time.Sleep(1000 * time.Millisecond)
	fmt.Println("sources: %v", dc.sources)
	return
}

// custom the source queue which DatabaseConsumer want to consume
func (dc *DatabaseConsumer) CustomSources(sources map[string]<-chan amqp.Delivery) map[string]<-chan amqp.Delivery {
	return sources
}

// publish data to PersistReadingsQueue
func (dc *DatabaseConsumer) PublishData(eventData interface{})  {
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

	dc.ch.Publish(
		"", //exchange string,
		qutils.PersistReadingsQueue, //key string,
		false, //mandatory bool,
		false, //immediate bool,
		msg)   //msg amqp.Publishing)
}
