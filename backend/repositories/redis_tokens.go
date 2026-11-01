package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type tokensRedis struct {
	client *redis.Client
}

// NewTokenStore crea el almacén de tokens temporales sobre Redis. Cada token expira solo (TTL).
func NewTokenStore(client *redis.Client) TokenStore {
	return &tokensRedis{client: client}
}

func claveToken(prefijo, token string) string { return fmt.Sprintf("token:%s:%s", prefijo, token) }

func (t *tokensRedis) Guardar(ctx context.Context, prefijo, token, valor string, ttl time.Duration) error {
	if err := t.client.Set(ctx, claveToken(prefijo, token), valor, ttl).Err(); err != nil {
		return fmt.Errorf("guardar token: %w", err)
	}
	return nil
}

func (t *tokensRedis) Obtener(ctx context.Context, prefijo, token string) (string, error) {
	valor, err := t.client.Get(ctx, claveToken(prefijo, token)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNoEncontrado
	}
	if err != nil {
		return "", fmt.Errorf("leer token: %w", err)
	}
	return valor, nil
}

func (t *tokensRedis) Eliminar(ctx context.Context, prefijo, token string) error {
	if err := t.client.Del(ctx, claveToken(prefijo, token)).Err(); err != nil {
		return fmt.Errorf("eliminar token: %w", err)
	}
	return nil
}
