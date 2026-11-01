package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB crea el pool de conexiones a PostgreSQL y comprueba que responde.
// Si existe DATABASE_URL (despliegue en la nube) tiene prioridad sobre las variables DB_*.
func ConnectDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := GetEnv("DATABASE_URL", "")
	if dsn == "" {
		u := &url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(GetEnv("DB_USER", "biess_user"), GetEnv("DB_PASSWORD", "biess_pass")),
			Host:   GetEnv("DB_HOST", "localhost") + ":" + GetEnv("DB_PORT", "5433"),
			Path:   GetEnv("DB_NAME", "biess_postulacion"),
		}
		q := u.Query()
		q.Set("sslmode", GetEnv("DB_SSLMODE", "disable"))
		u.RawQuery = q.Encode()
		dsn = u.String()
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("configuración de PostgreSQL inválida: %w", err)
	}
	cfg.MaxConns = int32(GetEnvInt("DB_MAX_CONNS", 10))
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("no se pudo crear el pool de PostgreSQL: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("PostgreSQL no responde (¿está levantado Docker?): %w", err)
	}

	log.Printf("Conectado a PostgreSQL (%s:%s)", cfg.ConnConfig.Host, fmt.Sprint(cfg.ConnConfig.Port))
	return pool, nil
}
