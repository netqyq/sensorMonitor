package main

import (
	"distributed/coordinator"
	//"fmt"
	//"sync"
)

var dc *coordinator.DatabaseConsumer
var wc *coordinator.WebappConsumer

func main() {
	// wait, like a daemon
	// var wg sync.WaitGroup
	// wg.Add(1)

	forever := make(chan bool)

	dc = coordinator.NewDatabaseConsumer()
	//wc = coordinator.NewWebappConsumer()
	ql := coordinator.NewQueueListener(dc)
	go ql.ListenForNewSource()


	// wg.Wait()
	<-forever
}
