//go:build redis || allstreams

package stream

import (
	"context"

	nc "github.com/geniusrabbit/notificationcenter/v2"
	"github.com/geniusrabbit/notificationcenter/v2/redis"
)

func init() {
	subscriberConnectors["redis"] = connectRedisSubscriber
	subscriberConnectors["rediss"] = connectRedisSubscriber
	publisherConnectors["redis"] = connectRedisPublisher
	publisherConnectors["rediss"] = connectRedisPublisher
}

func connectRedisSubscriber(_ context.Context, url string) (nc.Subscriber, error) {
	return redis.NewSubscriber(redis.WithRedisURL(url))
}

func connectRedisPublisher(_ context.Context, url string) (nc.Publisher, error) {
	return redis.NewPublisher(redis.WithRedisURL(url))
}
