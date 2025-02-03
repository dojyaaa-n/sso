package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/utils/logger"
	"syscall"
)

func main() {
	//Config loading
	cfg := config.MustLoadConfig()

	//Logger setup
	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting application", slog.String("env", cfg.Env))
	log.Debug("Debug messages are enabled")

	//gRPC server startup
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)
	go func() {
		application.GRPCServer.MustRunServer()
	}()

	//Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("Stopping application", slog.String("signal", sign.String()))

	application.GRPCServer.Stop()

	log.Info("Application stopped")
}
