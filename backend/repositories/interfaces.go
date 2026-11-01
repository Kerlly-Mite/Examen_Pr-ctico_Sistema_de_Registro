// Package repositories es la capa de acceso a datos. Los services dependen de estas INTERFACES,
// nunca de la implementación concreta (así se pueden cambiar o simular en pruebas).
package repositories

import (
	"context"

	"biess-backend/models"
)

// CatalogoRepository lee los catálogos de división política.
type CatalogoRepository interface {
	ListarPaises(ctx context.Context) ([]models.Pais, error)
	ListarProvincias(ctx context.Context, paisID int) ([]models.Provincia, error)
	ListarCantones(ctx context.Context, provinciaID int) ([]models.Canton, error)
	ExistePais(ctx context.Context, paisID int) (bool, error)
	ExisteProvincia(ctx context.Context, provinciaID int) (bool, error)
}

// EstadoIntentos describe el resultado de registrar un intento fallido.
type EstadoIntentos struct {
	Bloqueado       bool // true si con este fallo se alcanzó el máximo y se activó el bloqueo
	Restantes       int  // intentos que quedan antes del bloqueo
	SegundosBloqueo int  // duración del bloqueo (solo si Bloqueado)
}

// LimitadorIntentos controla los intentos fallidos consecutivos. Implementado sobre Redis (con TTL).
// "contexto" separa los contadores: "registro", "login", "recuperacion". "id" es la cédula.
type LimitadorIntentos interface {
	// Verificar devuelve los segundos de bloqueo restantes (0 si no está bloqueado).
	Verificar(ctx context.Context, contexto, id string) (int, error)
	RegistrarFallo(ctx context.Context, contexto, id string) (EstadoIntentos, error)
	Reiniciar(ctx context.Context, contexto, id string) error
}

// Pinger lo implementan PostgreSQL y Redis (a través de un adaptador) para el endpoint de salud.
type Pinger interface {
	Ping(ctx context.Context) error
}
