# event-bus

- log +
- cfg +
- все остальное)
    - сервис +
    - сервер +
    - мейн (норм шотдаун) +
- ридми -

// grpcurl -proto proto/subpub.proto -plaintext -d '{"key": "test"}' localhost:50051 pubsub.PubSub/Subscribe
// grpcurl -proto proto/subpub.proto -plaintext -d '{"key": "test", "data": "test message"}' localhost:50051 pubsub.PubSub/Publish