package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/MHSarmadi/Umbra/Server/web"
)

func main() {
	s, err := database.NewBadgerStore("./data")
	if err != nil {
		panic(err)
	}
	defer s.Close()

	srv := web.NewServer("localhost:8888", s)

	go func() {
		if err := srv.Run(); err != nil {
			log.Fatalf("server could not start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.ShutDown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
