package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"biess-backend/models"
)

type historialLaboralPG struct {
	db *pgxpool.Pool
}

// NewHistorialLaboralRepository recibe el pool por parámetro (Inyección de Dependencias).
// En producción se reemplazaría por un cliente del servicio de Historial Laboral del IESS con la misma interfaz.
func NewHistorialLaboralRepository(db *pgxpool.Pool) HistorialLaboralRepository {
	return &historialLaboralPG{db: db}
}

func (r *historialLaboralPG) BuscarPorCedula(ctx context.Context, cedula string) (*models.HistorialLaboral, error) {
	var h models.HistorialLaboral
	err := r.db.QueryRow(ctx,
		`SELECT es_afiliado, es_jubilado, es_empleador FROM historial_laboral_simulado WHERE cedula = $1`, cedula).
		Scan(&h.EsAfiliado, &h.EsJubilado, &h.EsEmpleador)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar historial laboral: %w", err)
	}
	return &h, nil
}
