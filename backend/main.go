// Servicio de Registro y Postulación al Directorio del BIESS (Frameworks de Programación Web).
//
// Arquitectura en capas con Inyección de Dependencias:
//   main (compone todo) -> handlers (HTTP) -> services (negocio) -> repositories (PostgreSQL / Redis)
// Cada capa conoce solo las INTERFACES de la capa inferior.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"biess-backend/config"
	"biess-backend/handlers"
	"biess-backend/repositories"
	"biess-backend/routes"
	"biess-backend/services"
)

func main() {
	config.LoadEnv()
	ctx := context.Background()

	// 1) Infraestructura: PostgreSQL y Redis son obligatorios.
	db, err := config.ConnectDB(ctx)
	if err != nil {
		log.Fatalf("Error de base de datos: %v", err)
	}
	defer db.Close()

	rdb, err := config.ConnectRedis(ctx)
	if err != nil {
		log.Fatalf("Error de Redis: %v", err)
	}
	defer rdb.Close()

	// 2) Repositories (reciben la conexión).
	catalogoRepo := repositories.NewCatalogoRepository(db)

	// 3) Services (reciben repositories por interfaz).
	catalogoSvc := services.NewCatalogoService(catalogoRepo)

	// 4) Handlers (reciben services por interfaz).
	deps := routes.Dependencias{
		Health:   handlers.NewHealthHandler(db, repositories.NewRedisPinger(rdb)),
		Catalogo: handlers.NewCatalogoHandler(catalogoSvc),
	}

	// 5) Servidor web.
	app := fiber.New(fiber.Config{AppName: "BIESS Postulación API"})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: origenesPermitidos(),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))
	routes.Registrar(app, deps)

	port := config.GetEnv("PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("El servidor se detuvo: %v", err)
		}
	}()
	log.Printf("API escuchando en http://localhost:%s", port)

	// Apagado ordenado con Ctrl+C o SIGTERM (Docker).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Apagando el servidor...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("Error al apagar: %v", err)
	}
}

// origenesPermitidos devuelve los orígenes CORS: el frontend de Vite por defecto más los de CORS_ORIGINS.
func origenesPermitidos() string {
	origenes := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	if extra := config.GetEnv("CORS_ORIGINS", ""); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origenes = append(origenes, o)
			}
		}
	}
	return strings.Join(origenes, ",")
}
