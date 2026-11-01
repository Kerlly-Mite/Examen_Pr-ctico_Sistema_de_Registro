package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"biess-backend/models"
)

// catalogoCache es un DECORADOR: implementa la misma interfaz CatalogoRepository, envuelve al repositorio
// de PostgreSQL y aplica el patrón Cache-Aside con Redis:
//   1) busca en Redis  ->  2) si no está (MISS), consulta PostgreSQL  ->  3) guarda en Redis con TTL.
// El service no se entera: recibe una interfaz y no cambia ni una línea.
type catalogoCache struct {
	origen CatalogoRepository
	client *redis.Client
	ttl    time.Duration
}

// NewCatalogoCache envuelve al repositorio real con caché. Si Redis falla, se responde desde PostgreSQL.
func NewCatalogoCache(origen CatalogoRepository, client *redis.Client, ttl time.Duration) CatalogoRepository {
	return &catalogoCache{origen: origen, client: client, ttl: ttl}
}

// obtener implementa Cache-Aside para cualquier lista (genéricos). Las listas vacías NO se cachean,
// para que ids inexistentes no llenen Redis de llaves basura.
func obtener[T any](ctx context.Context, c *catalogoCache, clave string, cargar func(context.Context) ([]T, error)) ([]T, error) {
	raw, err := c.client.Get(ctx, clave).Bytes()
	if err == nil {
		var datos []T
		if json.Unmarshal(raw, &datos) == nil {
			log.Printf("[cache] HIT  %s", clave)
			return datos, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("[cache] Redis no disponible (%v): se consulta PostgreSQL", err)
	}

	datos, err := cargar(ctx)
	if err != nil {
		return nil, err
	}
	if len(datos) > 0 {
		if payload, errJSON := json.Marshal(datos); errJSON == nil {
			if errSet := c.client.Set(ctx, clave, payload, c.ttl).Err(); errSet != nil {
				log.Printf("[cache] no se pudo guardar %s: %v", clave, errSet)
			}
		}
	}
	log.Printf("[cache] MISS %s (TTL %s)", clave, c.ttl)
	return datos, nil
}

func (c *catalogoCache) ListarPaises(ctx context.Context) ([]models.Pais, error) {
	return obtener(ctx, c, "catalogo:paises", c.origen.ListarPaises)
}

func (c *catalogoCache) ListarProvincias(ctx context.Context, paisID int) ([]models.Provincia, error) {
	clave := fmt.Sprintf("catalogo:provincias:%d", paisID)
	return obtener(ctx, c, clave, func(ctx context.Context) ([]models.Provincia, error) {
		return c.origen.ListarProvincias(ctx, paisID)
	})
}

func (c *catalogoCache) ListarCantones(ctx context.Context, provinciaID int) ([]models.Canton, error) {
	clave := fmt.Sprintf("catalogo:cantones:%d", provinciaID)
	return obtener(ctx, c, clave, func(ctx context.Context) ([]models.Canton, error) {
		return c.origen.ListarCantones(ctx, provinciaID)
	})
}

// Las verificaciones de existencia solo se usan cuando una lista llega vacía: pasan directo a PostgreSQL.
func (c *catalogoCache) ExistePais(ctx context.Context, paisID int) (bool, error) {
	return c.origen.ExistePais(ctx, paisID)
}

func (c *catalogoCache) ExisteProvincia(ctx context.Context, provinciaID int) (bool, error) {
	return c.origen.ExisteProvincia(ctx, provinciaID)
}
