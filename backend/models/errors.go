package models

import (
	"errors"
	"net/http"
)

// AppError es un error de negocio con código HTTP y un código estable para que el frontend decida qué mostrar.
// Las capas inferiores (repositories) devuelven errores "crudos"; los services los traducen a AppError.
type AppError struct {
	Status  int            // código HTTP
	Code    string         // identificador estable, p. ej. "no_encontrado"
	Message string         // mensaje para el usuario final (en español)
	Extra   map[string]any // campos adicionales que se envían al cliente (p. ej. intentos_restantes)
	Err     error          // causa original (no se expone al cliente)
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

// NuevoError construye un AppError sin causa interna.
func NuevoError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// ErrSolicitudInvalida: datos mal formados o incompletos (400).
func ErrSolicitudInvalida(msg string) *AppError {
	return NuevoError(http.StatusBadRequest, "solicitud_invalida", msg)
}

// ErrNoEncontrado: el recurso pedido no existe (404).
func ErrNoEncontrado(msg string) *AppError {
	return NuevoError(http.StatusNotFound, "no_encontrado", msg)
}

// ErrNoAutorizado: falta sesión o es inválida (401).
func ErrNoAutorizado(msg string) *AppError {
	return NuevoError(http.StatusUnauthorized, "no_autorizado", msg)
}

// ErrConflicto: el recurso ya existe (409).
func ErrConflicto(code, msg string) *AppError {
	return NuevoError(http.StatusConflict, code, msg)
}

// ErrIntentoFallido: credenciales o datos de identidad incorrectos (401). Informa cuántos intentos quedan.
func ErrIntentoFallido(code, msg string, restantes int) *AppError {
	return &AppError{
		Status:  http.StatusUnauthorized,
		Code:    code,
		Message: msg,
		Extra:   map[string]any{"intentos_restantes": restantes},
	}
}

// ErrBloqueado: se superó el máximo de intentos y hay un bloqueo temporal vigente (429).
func ErrBloqueado(msg string, segundos int) *AppError {
	return &AppError{
		Status:  http.StatusTooManyRequests,
		Code:    "cuenta_bloqueada",
		Message: msg,
		Extra:   map[string]any{"segundos_restantes": segundos},
	}
}

// ErrInterno envuelve un error inesperado sin exponer sus detalles al cliente (500).
func ErrInterno(err error) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Code: "error_interno", Message: "Ocurrió un error interno, intente nuevamente", Err: err}
}

// EsAppError indica si err (o alguno de sus envueltos) es un *AppError.
func EsAppError(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
