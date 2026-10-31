package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"biess-backend/repositories"
)

// HealthHandler informa si la API y sus dependencias (PostgreSQL y Redis) están disponibles.
type HealthHandler struct {
	db    repositories.Pinger
	redis repositories.Pinger
}

func NewHealthHandler(db, redis repositories.Pinger) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Check responde GET /api/health. Devuelve 200 si todo responde y 503 si alguna dependencia falla.
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.UserContext(), 3*time.Second)
	defer cancel()

	estado := func(err error) string {
		if err != nil {
			return "caido"
		}
		return "ok"
	}
	dbErr := h.db.Ping(ctx)
	redisErr := h.redis.Ping(ctx)

	codigo := fiber.StatusOK
	general := "ok"
	if dbErr != nil || redisErr != nil {
		codigo = fiber.StatusServiceUnavailable
		general = "degradado"
	}
	return c.Status(codigo).JSON(fiber.Map{
		"estado":     general,
		"postgresql": estado(dbErr),
		"redis":      estado(redisErr),
	})
}
