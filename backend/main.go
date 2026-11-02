// Servicio de Registro y Postulación al Directorio del BIESS (Frameworks de Programación Web).
//
// Arquitectura en capas con Inyección de Dependencias:
//
//	main (compone todo) -> handlers (HTTP) -> services (negocio) -> repositories (PostgreSQL / Redis)
//
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
	"biess-backend/middleware"
	"biess-backend/repositories"
	"biess-backend/routes"
	"biess-backend/services"
)

func main() {
	config.LoadEnv()
	ctx := context.Background()

	// Seguridad: el secreto del JWT es obligatorio (mínimo 32 caracteres).
	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET no está definido o es muy corto (mínimo 32 caracteres). Agrégalo a tu archivo .env")
	}

	// Reglas de EF-11: 3 intentos fallidos consecutivos -> bloqueo de 10 minutos (configurables).
	maxIntentos := config.GetEnvInt("MAX_INTENTOS", 3)
	if maxIntentos < 1 {
		maxIntentos = 3
	}
	bloqueoMin := config.GetEnvInt("BLOQUEO_MINUTOS", 10)
	if bloqueoMin < 1 {
		bloqueoMin = 10
	}

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

	// 2) Repositories (reciben la conexión). El de catálogos se envuelve con la caché Redis (decorador).
	catalogoTTL := time.Duration(config.GetEnvInt("CATALOGO_TTL_HORAS", 24)) * time.Hour
	catalogoRepo := repositories.NewCatalogoCache(repositories.NewCatalogoRepository(db), rdb, catalogoTTL)
	registroCivilRepo := repositories.NewRegistroCivilRepository(db)
	postulanteRepo := repositories.NewPostulanteRepository(db)
	limitador := repositories.NewLimitadorIntentos(rdb, maxIntentos, time.Duration(bloqueoMin)*time.Minute)
	tokenStore := repositories.NewTokenStore(rdb)
	postulacionRepo := repositories.NewPostulacionRepository(db)
	historialRepo := repositories.NewHistorialLaboralRepository(db)

	// 3) Services (reciben repositories por interfaz).
	gestorTokens := services.NewGestorTokens(jwtSecret, time.Duration(config.GetEnvInt("JWT_HORAS", 8))*time.Hour)
	catalogoSvc := services.NewCatalogoService(catalogoRepo)
	authSvc := services.NewAuthService(
		registroCivilRepo,
		postulanteRepo,
		limitador,
		tokenStore,
		gestorTokens,
		services.NewNotificadorConsola(),
		config.GetEnv("FRONTEND_URL", "http://localhost:5180"),
	)
	postulacionSvc := services.NewPostulacionService(
		postulacionRepo,
		postulanteRepo,
		registroCivilRepo,
		historialRepo,
		catalogoRepo, // con caché Redis: se usa para ubicar a Ecuador en el catálogo de países
		services.NewNotificadorPostulacionConsola(),
	)

	// 4) Handlers (reciben services por interfaz).
	deps := routes.Dependencias{
		Health:       handlers.NewHealthHandler(db, repositories.NewRedisPinger(rdb)),
		Catalogo:     handlers.NewCatalogoHandler(catalogoSvc),
		Auth:         handlers.NewAuthHandler(authSvc),
		Postulacion:  handlers.NewPostulacionHandler(postulacionSvc),
		RequiereAuth: middleware.RequiereAuth(gestorTokens),
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

	port := config.GetEnv("PORT", "3001")
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

// origenesPermitidos devuelve los orígenes CORS: el frontend de Vite (puerto 5180) más los de CORS_ORIGINS.
func origenesPermitidos() string {
	origenes := []string{"http://localhost:5180", "http://127.0.0.1:5180"}
	if extra := config.GetEnv("CORS_ORIGINS", ""); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			if o = strings.TrimSpace(o); o != "" {
				origenes = append(origenes, o)
			}
		}
	}
	return strings.Join(origenes, ",")
}
