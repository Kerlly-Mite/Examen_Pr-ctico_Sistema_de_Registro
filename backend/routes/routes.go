// Package routes registra todas las rutas de la API en un solo lugar.
package routes

import (
	"github.com/gofiber/fiber/v2"

	"biess-backend/handlers"
)

// Dependencias agrupa los handlers ya construidos (con sus services y repositories inyectados).
// En la Fase 4-5 se agrega aquí el handler de postulación.
type Dependencias struct {
	Health   *handlers.HealthHandler
	Catalogo *handlers.CatalogoHandler
	Auth     *handlers.AuthHandler

	// RequiereAuth es el middleware JWT. Las rutas que lo usan solo responden con sesión válida.
	// Ejemplo (Fase 4-5):  api.Group("/postulacion", d.RequiereAuth)
	RequiereAuth fiber.Handler
}

// Registrar enlaza cada ruta con su handler.
func Registrar(app *fiber.App, d Dependencias) {
	api := app.Group("/api")

	api.Get("/health", d.Health.Check)

	// Catálogos geográficos (públicos: se necesitan en el formulario). Con caché Redis (Cache-Aside).
	cat := api.Group("/catalogos")
	cat.Get("/paises", d.Catalogo.Paises)
	cat.Get("/paises/:id/provincias", d.Catalogo.Provincias)
	cat.Get("/provincias/:id/cantones", d.Catalogo.Cantones)

	// Autenticación y registro (EF-11): públicas, salvo /me.
	auth := api.Group("/auth")
	auth.Post("/verificar-identidad", d.Auth.VerificarIdentidad)
	auth.Post("/registro", d.Auth.Registrar)
	auth.Post("/login", d.Auth.Login)
	auth.Post("/recuperar/solicitar", d.Auth.SolicitarRecuperacion)
	auth.Post("/recuperar/restablecer", d.Auth.RestablecerPassword)
	auth.Get("/me", d.RequiereAuth, d.Auth.Me)
}
