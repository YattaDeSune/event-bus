package service

import (
	"context"

	"github.com/YattaDeSune/event-bus/internal/proto"
	"github.com/YattaDeSune/event-bus/pkg/subpub"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PubSubService struct {
	proto.PubSubServer
	bus subpub.SubPub
	log *zap.Logger
}

func NewPubSubService(bus subpub.SubPub, logger *zap.Logger) *PubSubService {
	return &PubSubService{
		bus: bus,
		log: logger,
	}
}

func (s *PubSubService) Subscribe(req *proto.SubscribeRequest, stream proto.PubSub_SubscribeServer) error {
	key := req.GetKey()
	if key == "" {
		return status.Error(codes.InvalidArgument, "empty key")
	}

	sub, err := s.bus.Subscribe(key, func(msg interface{}) {
		if event, ok := msg.(*proto.Event); ok {
			if err := stream.Send(event); err != nil {
				s.log.Error("failed to send event", zap.Error(err))
			}
		}
	})
	if err != nil {
		return status.Error(codes.Internal, "failed to subscribe")
	}
	s.log.Info("subscribtion added", zap.String("key", key))

	defer sub.Unsubscribe() // если стрим падает - отписываемся
	defer s.log.Info("subscribtion ended", zap.String("key", key))

	<-stream.Context().Done() // ждем пока клиент упадет, потом перестаем слушать
	return nil
}

func (s *PubSubService) Publish(ctx context.Context, req *proto.PublishRequest) (*emptypb.Empty, error) {
	key := req.GetKey()
	data := req.GetData()
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "empty key")
	}

	event := &proto.Event{Data: data}
	if err := s.bus.Publish(key, event); err != nil {
		s.log.Error("failed to publish event", zap.Error(err))
	}
	defer s.log.Info("event publicated", zap.String("key", key))

	return &emptypb.Empty{}, nil
}
