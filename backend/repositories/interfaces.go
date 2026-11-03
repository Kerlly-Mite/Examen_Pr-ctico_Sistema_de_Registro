// Package repositories es la capa de acceso a datos. Los services dependen de estas INTERFACES,
// nunca de la implementación concreta (así se pueden cambiar o simular en pruebas).
package repositories

import (
	"context"
	"time"

	"biess-backend/models"
)

// CatalogoRepository lee los catálogos de división política.
type CatalogoRepository interface {
	ListarPaises(ctx context.Context) ([]models.Pais, error)
	ListarProvincias(ctx context.Context, paisID int) ([]models.Provincia, error)
	ListarCantones(ctx context.Context, provinciaID int) ([]models.Canton, error)
	ExistePais(ctx context.Context, paisID int) (bool, error)
	ExisteProvincia(ctx context.Context, provinciaID int) (bool, error)
}

// RegistroCivilRepository consulta la fuente de identidad (simulada) por cédula.
type RegistroCivilRepository interface {
	// BuscarPorCedula devuelve ErrNoEncontrado si la cédula no existe.
	BuscarPorCedula(ctx context.Context, cedula string) (*models.RegistroCivil, error)
}

// HistorialLaboralRepository consulta la calidad de afiliado, jubilado o empleador (fuente simulada del IESS).
type HistorialLaboralRepository interface {
	// BuscarPorCedula devuelve ErrNoEncontrado si la cédula no tiene historial.
	BuscarPorCedula(ctx context.Context, cedula string) (*models.HistorialLaboral, error)
}

// DocumentoStore guarda los PDF de respaldo de cada postulante. Implementado en disco (carpeta UPLOADS_DIR);
// se podría cambiar por un almacenamiento en la nube con la misma interfaz.
type DocumentoStore interface {
	// Guardar almacena el archivo y devuelve su identificador (el valor que se guarda en documento_respaldo).
	Guardar(ctx context.Context, postulanteID int64, nombreOriginal string, datos []byte) (string, error)
	// Leer devuelve ErrNoEncontrado si el documento no existe o no pertenece a ese postulante.
	Leer(ctx context.Context, postulanteID int64, id string) ([]byte, error)
	Existe(ctx context.Context, postulanteID int64, id string) (bool, error)
	// Listar devuelve los documentos guardados del postulante (para controlar el máximo de 50 MB).
	Listar(ctx context.Context, postulanteID int64) ([]DocumentoInfo, error)
}

// DocumentoInfo describe un documento guardado sin leer su contenido.
type DocumentoInfo struct {
	ID         string
	Tamano     int64
	Modificado time.Time
}

// PostulanteRepository gestiona las cuentas de acceso.
type PostulanteRepository interface {
	// Crear devuelve ErrCedulaDuplicada o ErrCorreoDuplicado si ya existen.
	Crear(ctx context.Context, cedula, correo, passwordHash string) (int64, error)
	// BuscarPorCedula devuelve ErrNoEncontrado si no hay cuenta.
	BuscarPorCedula(ctx context.Context, cedula string) (*models.Postulante, error)
	BuscarPerfilPorID(ctx context.Context, id int64) (*models.PerfilPostulante, error)
	ActualizarPassword(ctx context.Context, id int64, passwordHash string) error
}

// PostulacionRepository persiste el formulario EF-12 (cabecera postulaciones + tablas hijas).
type PostulacionRepository interface {
	// BuscarPorPostulante devuelve ErrNoEncontrado si la persona aún no guarda su postulación.
	BuscarPorPostulante(ctx context.Context, postulanteID int64) (*models.Postulacion, error)
	// Guardar inserta o actualiza la postulación y reemplaza TODAS sus tablas hijas en UNA transacción
	// (BEGIN ... COMMIT). Si cualquier inserción falla se hace ROLLBACK y no queda nada a medias.
	// Devuelve ErrPostulacionEnviada si ya fue enviada, ErrCodigoInexistente si un cargo o pregunta no existe
	// y ErrDatoInvalido si la base rechaza un dato (CHECK, FK, NOT NULL, longitud).
	Guardar(ctx context.Context, p *models.Postulacion) error
	// MarcarEnviada pasa la postulación de BORRADOR a ENVIADA solo si no cambió desde que se validó
	// (bloqueo optimista con actualizado_en). Devuelve ErrPostulacionEnviada si ya estaba enviada
	// y ErrPostulacionModificada si otro guardado la cambió mientras tanto.
	MarcarEnviada(ctx context.Context, postulanteID int64, actualizadoEn time.Time) error
	ListarCargos(ctx context.Context) ([]models.Cargo, error)
	ListarPreguntasInhabilidad(ctx context.Context) ([]models.PreguntaInhabilidad, error)
}

// EstadoIntentos describe el resultado de registrar un intento fallido.
type EstadoIntentos struct {
	Bloqueado       bool // true si con este fallo se alcanzó el máximo y se activó el bloqueo
	Restantes       int  // intentos que quedan antes del bloqueo
	SegundosBloqueo int  // duración del bloqueo (solo si Bloqueado)
}

// LimitadorIntentos controla los intentos fallidos consecutivos. Implementado sobre Redis (con TTL).
// "contexto" separa los contadores: "registro", "login", "recuperacion". "id" es la cédula.
type LimitadorIntentos interface {
	// Verificar devuelve los segundos de bloqueo restantes (0 si no está bloqueado).
	Verificar(ctx context.Context, contexto, id string) (int, error)
	RegistrarFallo(ctx context.Context, contexto, id string) (EstadoIntentos, error)
	Reiniciar(ctx context.Context, contexto, id string) error
}

// TokenStore guarda tokens temporales de un solo propósito (verificación de registro, recuperación). Implementado en Redis.
type TokenStore interface {
	Guardar(ctx context.Context, prefijo, token, valor string, ttl time.Duration) error
	// Obtener devuelve ErrNoEncontrado si el token no existe o expiró.
	Obtener(ctx context.Context, prefijo, token string) (string, error)
	Eliminar(ctx context.Context, prefijo, token string) error
}

// Pinger lo implementan PostgreSQL y Redis (a través de un adaptador) para el endpoint de salud.
type Pinger interface {
	Ping(ctx context.Context) error
}
