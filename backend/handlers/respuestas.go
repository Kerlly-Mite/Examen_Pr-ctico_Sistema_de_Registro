// Package handlers es la capa HTTP: traduce peticiones a llamadas de services y errores a respuestas JSON.
// No contiene reglas de negocio ni SQL.
package handlers

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"

	"biess-backend/models"
)

// responderError convierte cualquier error en una respuesta JSON uniforme.
// Los AppError muestran su mensaje (y sus campos extra, p. ej. intentos_restantes);
// los errores inesperados se registran en el log y se ocultan al cliente.
func responderError(c *fiber.Ctx, err error) error {
	if ae, ok := models.EsAppError(err); ok {
		if ae.Err != nil {
			log.Printf("[%s] %s %s -> %v", ae.Code, c.Method(), c.Path(), ae.Err)
		}
		cuerpo := fiber.Map{"error": ae.Code, "message": ae.Message}
		for k, v := range ae.Extra {
			cuerpo[k] = v
		}
		return c.Status(ae.Status).JSON(cuerpo)
	}
	log.Printf("[error_interno] %s %s -> %v", c.Method(), c.Path(), err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error":   "error_interno",
		"message": "Ocurrió un error interno, intente nuevamente",
	})
}

// erroresFiber traduce los errores que genera el propio Fiber (antes de llegar a un handler) al mismo
// formato {error, message} del resto de la API.
var erroresFiber = map[int]*models.AppError{
	fiber.StatusNotFound:         models.NuevoError(fiber.StatusNotFound, "ruta_no_encontrada", "La ruta solicitada no existe"),
	fiber.StatusMethodNotAllowed: models.NuevoError(fiber.StatusMethodNotAllowed, "metodo_no_permitido", "El método HTTP no está permitido en esta ruta"),
}

// ManejadorErrores es el ErrorHandler global de Fiber: rutas inexistentes, métodos no permitidos,
// pánicos recuperados y cualquier error que un handler devuelva sin usar responderError responden en JSON.
func ManejadorErrores(c *fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		if ae, ok := erroresFiber[fe.Code]; ok {
			return responderError(c, ae)
		}
		if fe.Code < fiber.StatusInternalServerError {
			return responderError(c, models.NuevoError(fe.Code, "solicitud_invalida", fe.Message))
		}
	}
	return responderError(c, err)
}
