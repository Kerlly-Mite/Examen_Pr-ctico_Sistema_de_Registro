package models

import (
	"errors"
	"net/http"
)

// AppError es un error de negocio con código HTTP y un código estable para que el frontend decida qué mostrar.
// Las capas inferiores (repositories) devuelven errores "crudos"; los services los traducen a AppError.
type AppError struct {
	Status  int    // código HTTP
	Code    string // identificador estable, p. ej. "no_encontrado"
	Message string // mensaje para el usuario final (en español)
	Err     error  // causa original (no se expone al cliente)
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

// Errores de negocio frecuentes.
func ErrSolicitudInvalida(msg string) *AppError {
	return NuevoError(http.StatusBadRequest, "solicitud_invalida", msg)
}

func ErrNoEncontrado(msg string) *AppError {
	return NuevoError(http.StatusNotFound, "no_encontrado", msg)
}

// ErrInterno envuelve un error inesperado sin exponer sus detalles al cliente.
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
