package handlers

import (
	"github.com/gofiber/fiber/v2"

	"biess-backend/middleware"
	"biess-backend/models"
	"biess-backend/services"
)

// AuthHandler atiende las rutas /api/auth/* (EF-11).
type AuthHandler struct {
	svc services.AuthService
}

// NewAuthHandler recibe el service como interfaz (Inyección de Dependencias).
func NewAuthHandler(svc services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

var errCuerpoInvalido = models.ErrSolicitudInvalida("El formato de la solicitud no es válido")

// VerificarIdentidad responde POST /api/auth/verificar-identidad (paso 1 del registro).
func (h *AuthHandler) VerificarIdentidad(c *fiber.Ctx) error {
	var req models.VerificarIdentidadReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	resp, err := h.svc.VerificarIdentidad(c.UserContext(), req)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(resp)
}

// Registrar responde POST /api/auth/registro (paso 2 del registro).
func (h *AuthHandler) Registrar(c *fiber.Ctx) error {
	var req models.RegistroReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	if err := h.svc.Registrar(c.UserContext(), req); err != nil {
		return responderError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registro exitoso. Ya puede iniciar sesión con su cédula y contraseña.",
	})
}

// Login responde POST /api/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	resp, err := h.svc.Login(c.UserContext(), req)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(resp)
}

// Me responde GET /api/auth/me (ruta protegida): devuelve el perfil de quien tiene la sesión.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	perfil, err := h.svc.Perfil(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(perfil)
}

// SolicitarRecuperacion responde POST /api/auth/recuperar/solicitar.
func (h *AuthHandler) SolicitarRecuperacion(c *fiber.Ctx) error {
	var req models.VerificarIdentidadReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	if err := h.svc.SolicitarRecuperacion(c.UserContext(), req); err != nil {
		return responderError(c, err)
	}
	return c.JSON(fiber.Map{
		"message": "Si los datos son correctos, enviaremos un correo con las instrucciones para restablecer su contraseña.",
	})
}

// RestablecerPassword responde POST /api/auth/recuperar/restablecer.
func (h *AuthHandler) RestablecerPassword(c *fiber.Ctx) error {
	var req models.RestablecerReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	if err := h.svc.RestablecerPassword(c.UserContext(), req); err != nil {
		return responderError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Su contraseña fue actualizada. Ya puede iniciar sesión."})
}
