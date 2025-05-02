package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type GRPCServer struct {
	Host              string        `env:"HOST"`
	Port              string        `env:"PORT"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT_S"`
	MaxConnectionIdle time.Duration `env:"MAX_CONN_IDLE_S"`
}

// конфиг будем читать из .env
type Config struct {
	GRPCServer
}

func NewConfig(logger *zap.Logger) *Config {
	var cfg Config
	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		var (
			Host              = "0.0.0.0"
			Port              = "50051"
			ShutdownTimeout   = 10 * time.Second
			MaxConnectionIdle = 5 * time.Minute
		)

		logger.Error("error loading config, loaded default values",
			zap.Error(err),
			zap.String("Host", Host),
			zap.String("Port", Port),
			zap.Duration("ShutdownTimeout", ShutdownTimeout),
			zap.Duration("MaxConnectionIdle", MaxConnectionIdle),
		)
		return &Config{
			GRPCServer: GRPCServer{
				Host:              Host,
				Port:              Port,
				ShutdownTimeout:   ShutdownTimeout,
				MaxConnectionIdle: MaxConnectionIdle,
			},
		}
	}

	logger.Info("loaded config",
		zap.Error(err),
		zap.String("Host", cfg.GRPCServer.Host),
		zap.String("Port", cfg.GRPCServer.Port),
		zap.Duration("ShutdownTimeout", cfg.GRPCServer.ShutdownTimeout),
		zap.Duration("MaxConnectionIdle", cfg.GRPCServer.MaxConnectionIdle),
	)

	return &cfg
}
