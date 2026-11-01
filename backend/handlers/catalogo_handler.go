package handlers

import (
	"github.com/gofiber/fiber/v2"

	"biess-backend/models"
	"biess-backend/services"
)

// CatalogoHandler atiende las rutas /api/catalogos/*.
type CatalogoHandler struct {
	svc services.CatalogoService
}

// NewCatalogoHandler recibe el service como interfaz (Inyección de Dependencias).
func NewCatalogoHandler(svc services.CatalogoService) *CatalogoHandler {
	return &CatalogoHandler{svc: svc}
}

// Paises responde GET /api/catalogos/paises
func (h *CatalogoHandler) Paises(c *fiber.Ctx) error {
	data, err := h.svc.Paises(c.UserContext())
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(fiber.Map{"data": data})
}

// Provincias responde GET /api/catalogos/paises/:id/provincias
func (h *CatalogoHandler) Provincias(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return responderError(c, models.ErrSolicitudInvalida("El identificador del país debe ser numérico"))
	}
	data, err := h.svc.Provincias(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(fiber.Map{"data": data})
}

// Cantones responde GET /api/catalogos/provincias/:id/cantones
func (h *CatalogoHandler) Cantones(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return responderError(c, models.ErrSolicitudInvalida("El identificador de la provincia debe ser numérico"))
	}
	data, err := h.svc.Cantones(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(fiber.Map{"data": data})
}
