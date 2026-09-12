package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := ConnectMongo(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB at %s: %v", cfg.MongoURI, err)
	}
	log.Printf("connected to MongoDB (db=%s)", cfg.MongoDB)

	store := NewMongoStore(client, cfg.MongoDB)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	if err := SeedIfEmpty(seedCtx, store); err != nil {
		log.Fatalf("failed to seed data: %v", err)
	}

	api := NewAPI(store)
	router := NewRouter(api, cfg.FrontendOrigin)

	addr := ":" + cfg.Port
	log.Printf("media-sequencer backend listening on %s (frontend origin: %s)", addr, cfg.FrontendOrigin)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
