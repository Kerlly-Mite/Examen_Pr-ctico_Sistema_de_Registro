// Package services contiene la lógica de negocio. No conoce HTTP (Fiber) ni SQL: solo recibe
// datos, aplica reglas y se apoya en las interfaces de repositories.
package services

import (
	"context"
	"fmt"

	"biess-backend/models"
	"biess-backend/repositories"
)

// CatalogoService expone los casos de uso de consulta de catálogos geográficos.
type CatalogoService interface {
	Paises(ctx context.Context) ([]models.Pais, error)
	Provincias(ctx context.Context, paisID int) ([]models.Provincia, error)
	Cantones(ctx context.Context, provinciaID int) ([]models.Canton, error)
}

type catalogoService struct {
	repo repositories.CatalogoRepository
}

// NewCatalogoService recibe el repositorio como interfaz (Inyección de Dependencias).
// En main.go se le inyecta el repositorio CON caché Redis (decorador); este archivo no sabe que existe.
func NewCatalogoService(repo repositories.CatalogoRepository) CatalogoService {
	return &catalogoService{repo: repo}
}

func (s *catalogoService) Paises(ctx context.Context) ([]models.Pais, error) {
	paises, err := s.repo.ListarPaises(ctx)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return paises, nil
}

func (s *catalogoService) Provincias(ctx context.Context, paisID int) ([]models.Provincia, error) {
	if paisID <= 0 {
		return nil, models.ErrSolicitudInvalida("El identificador del país no es válido")
	}
	provincias, err := s.repo.ListarProvincias(ctx, paisID)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	// Una lista vacía es válida (países sin provincias cargadas: se ingresan manualmente),
	// salvo que el país no exista. Solo en ese caso se consulta la existencia.
	if len(provincias) == 0 {
		existe, err := s.repo.ExistePais(ctx, paisID)
		if err != nil {
			return nil, models.ErrInterno(fmt.Errorf("verificar país: %w", err))
		}
		if !existe {
			return nil, models.ErrNoEncontrado("El país indicado no existe")
		}
	}
	return provincias, nil
}

func (s *catalogoService) Cantones(ctx context.Context, provinciaID int) ([]models.Canton, error) {
	if provinciaID <= 0 {
		return nil, models.ErrSolicitudInvalida("El identificador de la provincia no es válido")
	}
	cantones, err := s.repo.ListarCantones(ctx, provinciaID)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	if len(cantones) == 0 {
		existe, err := s.repo.ExisteProvincia(ctx, provinciaID)
		if err != nil {
			return nil, models.ErrInterno(fmt.Errorf("verificar provincia: %w", err))
		}
		if !existe {
			return nil, models.ErrNoEncontrado("La provincia indicada no existe")
		}
	}
	return cantones, nil
}
