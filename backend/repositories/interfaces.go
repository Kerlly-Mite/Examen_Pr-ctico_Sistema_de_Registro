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

// Pinger lo implementan PostgreSQL y Redis (a través de un adaptador) para el endpoint de salud.
type Pinger interface {
	Ping(ctx context.Context) error
}
