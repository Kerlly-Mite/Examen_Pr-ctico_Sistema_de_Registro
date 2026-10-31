package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv carga el archivo .env. Busca primero en la carpeta actual (backend/) y luego en la raíz
// del repositorio (../.env), que es donde vive el .env compartido con docker-compose.
func LoadEnv() {
	for _, ruta := range []string{".env", "../.env"} {
		if err := godotenv.Load(ruta); err == nil {
			log.Printf("Variables de entorno cargadas desde %s", ruta)
			return
		}
	}
	log.Println("No se encontró archivo .env: se usan las variables del sistema y los valores por defecto")
}

// GetEnv devuelve el valor de una variable de entorno o un valor por defecto si está vacía.
func GetEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// GetEnvInt interpreta una variable de entorno como entero.
func GetEnvInt(key string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil {
		return n
	}
	return def
}
