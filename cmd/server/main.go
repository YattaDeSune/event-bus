package main

import (
	"github.com/YattaDeSune/event-bus/internal/config"
	"github.com/YattaDeSune/event-bus/internal/logger"
	"github.com/YattaDeSune/event-bus/internal/server"
	"go.uber.org/zap"
)

func main() {
	zapLogger := logger.NewLogger()
	cfg := config.NewConfig(zapLogger)
	srv := server.NewServer(cfg, zapLogger)

	if err := srv.Run(); err != nil {
		zapLogger.Fatal("failed to run server", zap.Error(err))
	}
}
