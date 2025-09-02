package main

import (
	"context"
	"log"

	"github.com/skinkvi/money_managment/internal/config"
	"github.com/skinkvi/money_managment/internal/storage"
	"github.com/skinkvi/money_managment/pkg/logger"
)

func main() {
	cfg, err := config.MustLoadConfig("../../config/dev.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log, err := logger.New(&cfg.Logger)
	if err != nil {
		return
	}

	ctx := context.Background()

	log.Info(ctx, "config load", logger.Field{
		Key:   "cfg",
		Value: cfg,
	})

	db, err := storage.Connect(ctx, cfg.DataBase, log)
	if err != nil {
		return
	}

	// TODO: run migrations
	// TODO: init Redis cache
	// TODO: setup Gin router
	// TODO: start HTTP server with graceful shutdown

}
