package services

import (
	"context"
	"log"
)

// Notificador envía correos. En el documento funcional el correo de recuperación sale automáticamente;
// aquí se SIMULA escribiendo el enlace en la consola del backend (no hay servidor SMTP en el examen).
type Notificador interface {
	EnviarRecuperacion(ctx context.Context, correo, enlace string) error
}

type notificadorConsola struct{}

// NewNotificadorConsola devuelve el notificador simulado. Para uso real se cambiaría por uno SMTP
// que implemente la misma interfaz, sin tocar el AuthService.
func NewNotificadorConsola() Notificador { return notificadorConsola{} }

func (notificadorConsola) EnviarRecuperacion(_ context.Context, correo, enlace string) error {
	log.Printf("[CORREO SIMULADO] Para: %s | Restablezca su contraseña en: %s", correo, enlace)
	return nil
}
