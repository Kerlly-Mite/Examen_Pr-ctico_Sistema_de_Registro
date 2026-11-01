package repositories

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// redisPinger adapta *redis.Client (cuyo Ping devuelve *StatusCmd) a la interfaz Pinger.
type redisPinger struct {
	client *redis.Client
}

// NewRedisPinger devuelve un Pinger para el cliente de Redis.
func NewRedisPinger(client *redis.Client) Pinger {
	return &redisPinger{client: client}
}

func (p *redisPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx).Err()
}
