package main

import (
	"log"

	"github.com/luqus/templater/storage"
)

func main() {

	sqliteStorage, err := storage.NewSqliteStorage()
	if err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":3000", sqliteStorage)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
