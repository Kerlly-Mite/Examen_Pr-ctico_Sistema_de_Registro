package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"biess-backend/models"
)

// catalogoPG es la implementación de CatalogoRepository sobre PostgreSQL.
type catalogoPG struct {
	db *pgxpool.Pool
}

// NewCatalogoRepository recibe el pool por parámetro (Inyección de Dependencias).
func NewCatalogoRepository(db *pgxpool.Pool) CatalogoRepository {
	return &catalogoPG{db: db}
}

func (r *catalogoPG) ListarPaises(ctx context.Context) ([]models.Pais, error) {
	rows, err := r.db.Query(ctx, `SELECT id, nombre FROM paises ORDER BY nombre`)
	if err != nil {
		return nil, fmt.Errorf("listar países: %w", err)
	}
	defer rows.Close()

	out := []models.Pais{} // slice vacío (no nil) para que el JSON sea [] y no null
	for rows.Next() {
		var p models.Pais
		if err := rows.Scan(&p.ID, &p.Nombre); err != nil {
			return nil, fmt.Errorf("leer país: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *catalogoPG) ListarProvincias(ctx context.Context, paisID int) ([]models.Provincia, error) {
	rows, err := r.db.Query(ctx, `SELECT id, pais_id, nombre FROM provincias WHERE pais_id = $1 ORDER BY nombre`, paisID)
	if err != nil {
		return nil, fmt.Errorf("listar provincias: %w", err)
	}
	defer rows.Close()

	out := []models.Provincia{}
	for rows.Next() {
		var p models.Provincia
		if err := rows.Scan(&p.ID, &p.PaisID, &p.Nombre); err != nil {
			return nil, fmt.Errorf("leer provincia: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *catalogoPG) ListarCantones(ctx context.Context, provinciaID int) ([]models.Canton, error) {
	rows, err := r.db.Query(ctx, `SELECT id, provincia_id, nombre FROM cantones WHERE provincia_id = $1 ORDER BY nombre`, provinciaID)
	if err != nil {
		return nil, fmt.Errorf("listar cantones: %w", err)
	}
	defer rows.Close()

	out := []models.Canton{}
	for rows.Next() {
		var c models.Canton
		if err := rows.Scan(&c.ID, &c.ProvinciaID, &c.Nombre); err != nil {
			return nil, fmt.Errorf("leer cantón: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *catalogoPG) ExistePais(ctx context.Context, paisID int) (bool, error) {
	var existe bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM paises WHERE id = $1)`, paisID).Scan(&existe)
	return existe, err
}

func (r *catalogoPG) ExisteProvincia(ctx context.Context, provinciaID int) (bool, error) {
	var existe bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM provincias WHERE id = $1)`, provinciaID).Scan(&existe)
	return existe, err
}
