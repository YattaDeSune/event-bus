package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/YattaDeSune/event-bus/internal/config"
	"github.com/YattaDeSune/event-bus/internal/service"
	"github.com/YattaDeSune/event-bus/internal/subpub"
	"github.com/YattaDeSune/event-bus/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Server struct {
	grpcServer *grpc.Server
	cfg        *config.Config
	log        *zap.Logger
	bus        subpub.SubPub
}

func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	bus := subpub.NewSubPub()

	// регистрируем сервер
	// если клиент неактивен - соединение дропается через MaxConnectionIdle секунд
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: cfg.MaxConnectionIdle,
		}),
	)
	pubSubService := service.NewPubSubService(bus, logger)
	proto.RegisterPubSubServer(grpcServer, pubSubService)

	return &Server{
		grpcServer: grpcServer,
		cfg:        cfg,
		log:        logger,
		bus:        bus,
	}
}

func (s *Server) Run() error {
	addr := net.JoinHostPort(s.cfg.GRPCServer.Host, s.cfg.GRPCServer.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Info("starting gRPC server", zap.String("addr", addr))

	// в отдельной горутине, чтобы не блокироваться
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// оставляем таймаут для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.GRPCServer.ShutdownTimeout)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// при получении сиграла начинаем graceful shutdown
	<-sigChan
	s.log.Info("shutting down gRPC server...")

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.log.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.grpcServer.Stop()
		s.log.Info("gRPC server stopped before timeout")
	}

	// Закрываем шину событий
	if err := s.bus.Close(ctx); err != nil {
		s.log.Error("failed to close event bus", zap.Error(err))
	}

	return nil
}
