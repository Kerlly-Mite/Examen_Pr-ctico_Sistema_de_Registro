package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"biess-backend/models"
)

type registroCivilPG struct {
	db *pgxpool.Pool
}

// NewRegistroCivilRepository recibe el pool por parámetro (Inyección de Dependencias).
func NewRegistroCivilRepository(db *pgxpool.Pool) RegistroCivilRepository {
	return &registroCivilPG{db: db}
}

func (r *registroCivilPG) BuscarPorCedula(ctx context.Context, cedula string) (*models.RegistroCivil, error) {
	var rc models.RegistroCivil
	err := r.db.QueryRow(ctx, `
		SELECT cedula, nombres, apellidos, genero, fecha_nacimiento, nacionalidad,
		       estado_civil, fecha_expedicion, codigo_dactilar
		FROM registro_civil_simulado
		WHERE cedula = $1`, cedula).
		Scan(&rc.Cedula, &rc.Nombres, &rc.Apellidos, &rc.Genero, &rc.FechaNacimiento, &rc.Nacionalidad,
			&rc.EstadoCivil, &rc.FechaExpedicion, &rc.CodigoDactilar)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar en registro civil: %w", err)
	}
	return &rc, nil
}
