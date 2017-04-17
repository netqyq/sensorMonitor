package controller

import (
	"bytes"
	"distributed/dto"
	"distributed/qutils"
	"distributed/web/model"
	"encoding/gob"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"net/http"
	"sync"
	"fmt"
)

const url = "amqp://guest:guest@localhost:5672"

type websocketController struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	sockets  []*websocket.Conn
	mutex    sync.Mutex
	upgrader websocket.Upgrader
}

func newWebsocketController() *websocketController {
	wsc := new(websocketController)

	wsc.conn, wsc.ch = qutils.GetChannel(url)

	wsc.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	go wsc.listenForSources()
	go wsc.listenForMessages()

	return wsc
}

func (wsc *websocketController) discoverSensorName()  {
	q := qutils.GetQueue(qutils.WebappDiscoveryQueue, wsc.ch, false)
	wsc.ch.Publish("", q.Name, false, false, amqp.Publishing{})
	fmt.Println("discovery meassge sent.")
}

func (wsc *websocketController) handleMessage(w http.ResponseWriter, r *http.Request) {
	socket, _ := wsc.upgrader.Upgrade(w, r, nil)
	wsc.addSocket(socket)
	go wsc.listenForDiscoveryRequests(socket)
}

func (wsc *websocketController) addSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	wsc.sockets = append(wsc.sockets, socket)
	wsc.mutex.Unlock()
}

func (wsc *websocketController) removeSocket(socket *websocket.Conn) {
	wsc.mutex.Lock()
	socket.Close()

	for i := range wsc.sockets {
		if wsc.sockets[i] == socket {
			wsc.sockets = append(wsc.sockets[:i], wsc.sockets[i+1:]...)

		}
	}
	wsc.mutex.Unlock()
}

func (wsc *websocketController) listenForSources() {
	fmt.Println("listen for sources called")
	qutils.CreateExchange(wsc.ch, qutils.WebappSourceExchange, "fanout")
	q := qutils.GetQueue("", wsc.ch, true)
	err := wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		qutils.WebappSourceExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)
	qutils.FailOnError(err, "bind failed")

	msgs, err := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)
  qutils.FailOnError(err, "consume failed")
	fmt.Println("after consume")
	// send discovery meassage to coordinator
	// dirct, only one coordinator response is ok.
	// must call this after Consume is succeed
	wsc.discoverSensorName()
	go wsc.handleSources(msgs)

}

func (wsc *websocketController) handleSources(msgs <-chan amqp.Delivery)  {
	for msg := range msgs {
		fmt.Println("msg: ", string(msg.Body))
		sensor, err := model.GetSensorByName(string(msg.Body))
		fmt.Printf("Received source: %v\n", string(msg.Body))
		if err != nil {
			println(err.Error())
		}
		wsc.sendMessage(message{
			Type: "source",
			Data: sensor,
		})
	}
}


func (wsc *websocketController) listenForMessages() {
	qutils.CreateExchange(wsc.ch, qutils.WebappReadingsExchange, "fanout")
	q := qutils.GetQueue("", wsc.ch, true)

	wsc.ch.QueueBind(
		q.Name, //name string,
		"",     //key string,
		qutils.WebappReadingsExchange, //exchange string,
		false, //noWait bool,
		nil)   //args amqp.Table)

	msgs, _ := wsc.ch.Consume(
		q.Name, //queue string,
		"",     //consumer string,
		true,   //autoAck bool,
		false,  //exclusive bool,
		false,  //noLocal bool,
		false,  //noWait bool,
		nil)    //args amqp.Table)

	for msg := range msgs {
		buf := bytes.NewBuffer(msg.Body)
		dec := gob.NewDecoder(buf)
		sm := dto.SensorMessage{}
		err := dec.Decode(&sm)
		fmt.Printf("Received message: %v\n", sm)

		if err != nil {
			println(err.Error())
		}

		wsc.sendMessage(message{
			Type: "reading",
			Data: sm,
		})

	}
}

func (wsc *websocketController) sendMessage(msg message) {
	socketsToRemove := []*websocket.Conn{}

	for _, socket := range wsc.sockets {
		err := socket.WriteJSON(msg)

		if err != nil {
			socketsToRemove = append(socketsToRemove, socket)
		}
	}

	for _, socket := range socketsToRemove {
		wsc.removeSocket(socket)
	}
}

// listen for discovery request from web client(JavaScript)
func (wsc *websocketController) listenForDiscoveryRequests(socket *websocket.Conn) {
	for {
		msg := message{}
		err := socket.ReadJSON(&msg)

		if err != nil {
			wsc.removeSocket(socket)
			break
		}

		if msg.Type == "discover" {
			wsc.ch.Publish(
				"", //exchange string,
				qutils.WebappDiscoveryQueue, //key string,
				false,             //mandatory bool,
				false,             //immediate bool,
				amqp.Publishing{}) //msg amqp.Publishing)
		}
	}
}

type message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
