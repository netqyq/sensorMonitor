package main

import (
	"distributed/web/controller"
	"net/http"
	"fmt"
)

func main() {
	controller.Initialize()
	err := http.ListenAndServe(":4000", nil)
	fmt.Println(err, "start server failed")
}
