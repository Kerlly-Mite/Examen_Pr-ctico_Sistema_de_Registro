// Package routes registra todas las rutas de la API en un solo lugar.
package routes

import (
	"github.com/gofiber/fiber/v2"

	"biess-backend/handlers"
)

// Dependencias agrupa los handlers ya construidos (con sus services y repositories inyectados).
// En las fases siguientes se agregan aquí Auth y Postulacion.
type Dependencias struct {
	Health   *handlers.HealthHandler
	Catalogo *handlers.CatalogoHandler
}

// Registrar enlaza cada ruta con su handler.
func Registrar(app *fiber.App, d Dependencias) {
	api := app.Group("/api")

	api.Get("/health", d.Health.Check)

	// Catálogos geográficos (públicos: se necesitan en el formulario). La caché Redis llega en la Fase 3.
	cat := api.Group("/catalogos")
	cat.Get("/paises", d.Catalogo.Paises)
	cat.Get("/paises/:id/provincias", d.Catalogo.Provincias)
	cat.Get("/provincias/:id/cantones", d.Catalogo.Cantones)
}
