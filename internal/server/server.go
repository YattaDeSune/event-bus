package server

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/YattaDeSune/event-bus/internal/config"
	"github.com/YattaDeSune/event-bus/internal/proto"
	"github.com/YattaDeSune/event-bus/internal/service"
	"github.com/YattaDeSune/event-bus/pkg/subpub"
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
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.log.Info("starting gRPC server", zap.String("addr", addr))

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.log.Fatal("failed to serve", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	s.log.Info("shutting down gRPC server...")

	// контекст с таймаутом для graceful shutdown
	ctx, busCancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer busCancel()

	grpcDone := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(grpcDone)
	}()

	select {
	case <-grpcDone:
		s.log.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.grpcServer.Stop()
		s.log.Warn("gRPC server stopped by timeout")
	}

	// закрываем шину
	if err := s.bus.Close(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.log.Warn("event bus closed by timeout")
		} else {
			s.log.Error("failed to close event bus", zap.Error(err))
		}
	}

	return nil
}
