package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MHSarmadi/Umbra/Server/database"
	"github.com/MHSarmadi/Umbra/Server/logger"
	"github.com/MHSarmadi/Umbra/Server/web"
)

func main() {
	if err := logger.Init("Logs"); err != nil {
		panic(err)
	}
	defer logger.Close()

	s, err := database.NewBadgerStore("./data")
	if err != nil {
		panic(err)
	}
	defer s.Close()

	mainCtx, cancelMain := context.WithCancel(context.Background())
	defer cancelMain()

	go s.StartExpiryJanitor(mainCtx, 1*time.Minute)

	srv := web.NewServer(mainCtx, "localhost:8888", s)
	logger.Infof("server starting on %s", "localhost:8888")

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-quit:
		logger.Infof("shutdown signal received")
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("server failed: %v", err)
		}
		cancelMain()
		return
	}
	cancelMain()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.ShutDown(ctx); err != nil {
		logger.Errorf("shutdown error: %v", err)
		return
	}
	logger.Infof("server stopped gracefully")
}
