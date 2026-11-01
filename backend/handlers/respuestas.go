// Package handlers es la capa HTTP: traduce peticiones a llamadas de services y errores a respuestas JSON.
// No contiene reglas de negocio ni SQL.
package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"biess-backend/models"
)

// responderError convierte cualquier error en una respuesta JSON uniforme.
// Los AppError muestran su mensaje; los errores inesperados se registran en el log y se ocultan al cliente.
func responderError(c *fiber.Ctx, err error) error {
	if ae, ok := models.EsAppError(err); ok {
		if ae.Err != nil {
			log.Printf("[%s] %s %s -> %v", ae.Code, c.Method(), c.Path(), ae.Err)
		}
		return c.Status(ae.Status).JSON(fiber.Map{"error": ae.Code, "message": ae.Message})
	}
	log.Printf("[error_interno] %s %s -> %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "error_interno",
		"message": "Ocurrió un error interno, intente nuevamente",
	})
}
