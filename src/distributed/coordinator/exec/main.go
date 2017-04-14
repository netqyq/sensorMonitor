package main

import (
	"distributed/coordinator"
	//"fmt"
	"sync"
)

var dc *coordinator.DatabaseConsumer
var wc *coordinator.WebappConsumer

func main() {
	// wait, like a daemon
	var wg sync.WaitGroup
	wg.Add(1)

	ea := coordinator.NewEventAggregator()
	dc = coordinator.NewDatabaseConsumer(ea)
	wc = coordinator.NewWebappConsumer(ea)
	ql := coordinator.NewQueueListener(ea)
	go ql.ListenForNewSource()

	// var a string
	// fmt.Scanln(&a)

	wg.Wait()
}
