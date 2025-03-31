package main

import (
	"artio-relay/pkg/relay"
	"artio-relay/pkg/server"
	"artio-relay/pkg/storage"
	"log"
)

/*
main the main running function
*/
func main() {

	log.Printf("Initializing storage...")
	store := storage.Storage{}
	err := store.Init()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[OK] Initializing storage")

	log.Printf("Initializing relay...")
	r := *relay.NewRelay(store)
	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("[OK] Initializing relay")

	log.Println("Initializing server...")

	s := server.NewServer(&r)
	_ = s.Start()

	log.Println("[OK] Initializing server")
}
