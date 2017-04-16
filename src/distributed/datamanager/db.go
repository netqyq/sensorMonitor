package datamanager

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres",
		"postgres://distributed:distributed@localhost/distributed?sslmode=disable")
	if err != nil {
		log.Println("can not connect to database!")
		panic(err.Error())
	}
}
