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

	ea := coordinator.NewEventAggregator()
	wc = coordinator.NewWebappConsumer(ea)
	ql := coordinator.NewQueueListener(ea)
	go ql.ListenForNewSource()
	dc = coordinator.NewDatabaseConsumer(ea)

	// var a string
	// fmt.Scanln(&a)

	// wg.Wait()
	<-forever
}
