package repositories

import "errors"

// Errores "centinela" de la capa de datos. Los services los reconocen con errors.Is y los traducen a mensajes de negocio.
var (
	ErrNoEncontrado    = errors.New("registro no encontrado")
	ErrCorreoDuplicado = errors.New("el correo ya está registrado")
	ErrCedulaDuplicada = errors.New("la cédula ya está registrada")
)
