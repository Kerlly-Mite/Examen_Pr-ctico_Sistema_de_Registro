package handlers

import (
	"github.com/gofiber/fiber/v2"

	"biess-backend/middleware"
	"biess-backend/models"
	"biess-backend/services"
)

// PostulacionHandler atiende las rutas /api/postulacion/* (EF-12). Todas requieren sesión.
type PostulacionHandler struct {
	svc services.PostulacionService
}

// NewPostulacionHandler recibe el service como interfaz (Inyección de Dependencias).
func NewPostulacionHandler(svc services.PostulacionService) *PostulacionHandler {
	return &PostulacionHandler{svc: svc}
}

// Obtener responde GET /api/postulacion: la postulación de quien tiene la sesión.
func (h *PostulacionHandler) Obtener(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	p, err := h.svc.Obtener(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(p)
}

// Guardar responde PUT /api/postulacion: guarda el formulario completo en una transacción.
func (h *PostulacionHandler) Guardar(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	var req models.PostulacionReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	p, err := h.svc.Guardar(c.UserContext(), id, req)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(p)
}

// Opciones responde GET /api/postulacion/opciones: cargos y preguntas de inhabilidades.
func (h *PostulacionHandler) Opciones(c *fiber.Ctx) error {
	op, err := h.svc.Opciones(c.UserContext())
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(op)
}
