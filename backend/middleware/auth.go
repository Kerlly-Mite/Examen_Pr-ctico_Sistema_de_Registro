// Package middleware contiene filtros que se ejecutan antes de los handlers.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"biess-backend/services"
)

// Claves con las que se guarda la sesión en c.Locals(...) para que los handlers la lean.
const (
	LocalPostulanteID = "postulante_id" // int64
	LocalCedula       = "cedula"        // string
)

// RequiereAuth protege rutas: exige el encabezado "Authorization: Bearer <jwt>" con un token válido y vigente.
// Recibe la interfaz GestorTokens (Inyección de Dependencias): no conoce el secreto ni el algoritmo.
func RequiereAuth(tokens services.GestorTokens) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cabecera := c.Get("Authorization")
		if !strings.HasPrefix(cabecera, "Bearer ") {
			return noAutorizado(c)
		}
		claims, err := tokens.Validar(strings.TrimSpace(strings.TrimPrefix(cabecera, "Bearer ")))
		if err != nil {
			return noAutorizado(c)
		}
		c.Locals(LocalPostulanteID, claims.PostulanteID)
		c.Locals(LocalCedula, claims.Cedula)
		return c.Next()
	}
}

func noAutorizado(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"error":   "no_autorizado",
		"message": "Su sesión no es válida o expiró. Inicie sesión nuevamente.",
	})
}
