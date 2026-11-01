package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"biess-backend/models"
)

type postulantePG struct {
	db *pgxpool.Pool
}

// NewPostulanteRepository recibe el pool por parámetro (Inyección de Dependencias).
func NewPostulanteRepository(db *pgxpool.Pool) PostulanteRepository {
	return &postulantePG{db: db}
}

func (r *postulantePG) Crear(ctx context.Context, cedula, correo, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO postulantes (cedula, correo, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		cedula, correo, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		// 23505 = unique_violation. El nombre de la restricción indica qué columna se repitió.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "correo") {
				return 0, ErrCorreoDuplicado
			}
			return 0, ErrCedulaDuplicada
		}
		return 0, fmt.Errorf("crear postulante: %w", err)
	}
	return id, nil
}

func (r *postulantePG) BuscarPorCedula(ctx context.Context, cedula string) (*models.Postulante, error) {
	var p models.Postulante
	err := r.db.QueryRow(ctx,
		`SELECT id, cedula, correo, password_hash, creado_en FROM postulantes WHERE cedula = $1`, cedula).
		Scan(&p.ID, &p.Cedula, &p.Correo, &p.PasswordHash, &p.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar postulante por cédula: %w", err)
	}
	return &p, nil
}

func (r *postulantePG) BuscarPerfilPorID(ctx context.Context, id int64) (*models.PerfilPostulante, error) {
	var p models.PerfilPostulante
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.cedula, p.correo, rc.nombres, rc.apellidos
		FROM postulantes p
		JOIN registro_civil_simulado rc ON rc.cedula = p.cedula
		WHERE p.id = $1`, id).
		Scan(&p.ID, &p.Cedula, &p.Correo, &p.Nombres, &p.Apellidos)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("buscar perfil: %w", err)
	}
	return &p, nil
}

func (r *postulantePG) ActualizarPassword(ctx context.Context, id int64, passwordHash string) error {
	tag, err := r.db.Exec(ctx, `UPDATE postulantes SET password_hash = $1 WHERE id = $2`, passwordHash, id)
	if err != nil {
		return fmt.Errorf("actualizar contraseña: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNoEncontrado
	}
	return nil
}
