package controller

import (
	"net/http"
	"fmt"
)

var ws = newWebsocketController()

func Initialize() {
	registerRoutes()
	registerFileServers()
}

func registerRoutes() {
	http.HandleFunc("/ws", ws.handleMessage)
}

func registerFileServers() {
	http.Handle("/", http.FileServer(http.Dir("./assets/")))


	http.Handle("/public/lib/",http.StripPrefix("/public/lib/",
			http.FileServer(http.Dir("node_modules"))))
	fmt.Println("after register")
}
