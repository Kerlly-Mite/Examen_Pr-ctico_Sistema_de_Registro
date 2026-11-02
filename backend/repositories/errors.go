package repositories

import (
	"errors"
	"fmt"
)

// Errores "centinela" de la capa de datos. Los services los reconocen con errors.Is y los traducen a mensajes de negocio.
var (
	ErrNoEncontrado    = errors.New("registro no encontrado")
	ErrCorreoDuplicado = errors.New("el correo ya está registrado")
	ErrCedulaDuplicada = errors.New("la cédula ya está registrada")

	ErrPostulacionEnviada    = errors.New("la postulación ya fue enviada y no se puede modificar")
	ErrPostulacionModificada = errors.New("la postulación cambió mientras se validaba el envío")
	ErrCodigoInexistente     = errors.New("código de catálogo inexistente")
	ErrDatoInvalido          = errors.New("la base de datos rechazó un dato")
)

// DatoInvalidoError detalla qué tabla y restricción rechazó el dato. Envuelve ErrDatoInvalido
// (errors.Is(err, ErrDatoInvalido) == true) y con errors.As el service sabe en qué sección ocurrió.
type DatoInvalidoError struct {
	Tabla       string // p. ej. "experiencia_profesional"
	Restriccion string // nombre del CHECK/FK o la columna afectada
	Causa       error
}

func (e *DatoInvalidoError) Error() string {
	return fmt.Sprintf("dato inválido en %s (%s): %v", e.Tabla, e.Restriccion, e.Causa)
}

func (e *DatoInvalidoError) Unwrap() []error { return []error{ErrDatoInvalido, e.Causa} }

// CodigoInexistenteError indica qué código de catálogo no existe (cargo o pregunta de inhabilidad).
type CodigoInexistenteError struct {
	Catalogo string // "cargos" | "inhabilidad_preguntas"
	Codigo   string
}

func (e *CodigoInexistenteError) Error() string {
	return fmt.Sprintf("el código %q no existe en %s", e.Codigo, e.Catalogo)
}

func (e *CodigoInexistenteError) Unwrap() error { return ErrCodigoInexistente }
