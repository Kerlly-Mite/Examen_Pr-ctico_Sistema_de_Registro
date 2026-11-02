package services

import (
	"context"
	"fmt"
	"html"

	"biess-backend/models"
)

// NotificadorCorreo arma los correos del sistema y los entrega con un EnviadorCorreo (SMTP real o consola).
// Implementa las dos interfaces de notificación:
//   - Notificador (EF-11): enlace de recuperación de contraseña.
//   - NotificadorPostulacion (EF-12, 1.10.3): confirmación del envío con el formulario en PDF adjunto.
type NotificadorCorreo struct {
	enviador EnviadorCorreo
}

// NewNotificadorCorreo recibe el enviador como interfaz (Inyección de Dependencias).
func NewNotificadorCorreo(enviador EnviadorCorreo) *NotificadorCorreo {
	return &NotificadorCorreo{enviador: enviador}
}

var (
	_ Notificador            = (*NotificadorCorreo)(nil)
	_ NotificadorPostulacion = (*NotificadorCorreo)(nil)
)

func (n *NotificadorCorreo) EnviarRecuperacion(ctx context.Context, correo, enlace string) error {
	return n.enviador.Enviar(ctx, Mensaje{
		Para:   correo,
		Asunto: "Recuperación de contraseña - Postulación Directorio BIESS",
		Texto: "Recibimos una solicitud para restablecer su contraseña.\n\n" +
			"Abra este enlace (vence en 30 minutos): " + enlace + "\n\n" +
			"Si usted no la solicitó, ignore este correo.",
		HTML: plantillaCorreo("Recuperación de contraseña",
			`<p>Recibimos una solicitud para restablecer su contraseña.</p>
			 <p><a href="`+html.EscapeString(enlace)+`" style="display:inline-block;background:#1d4ed8;color:#ffffff;
			    padding:10px 18px;border-radius:6px;text-decoration:none">Restablecer contraseña</a></p>
			 <p style="color:#555555">El enlace vence en 30 minutos. Si usted no la solicitó, ignore este correo.</p>`),
	})
}

func (n *NotificadorCorreo) EnviarConfirmacion(ctx context.Context, p *models.Postulacion, formularioPDF []byte) error {
	nombre := p.Nombres + " " + p.Apellidos
	return n.enviador.Enviar(ctx, Mensaje{
		Para:   p.Correo1,
		Asunto: "Postulación " + p.Codigo + " enviada - Directorio BIESS",
		Texto: fmt.Sprintf("Estimado/a %s:\n\nSu postulación %s fue enviada correctamente.\n"+
			"Adjuntamos el formulario de postulación con la información registrada. Imprímalo, fírmelo y entréguelo "+
			"junto con los requisitos en el lugar, fecha y horario que establezca el IESS.", nombre, p.Codigo),
		HTML: plantillaCorreo("Postulación enviada",
			`<p>Estimado/a <b>`+html.EscapeString(nombre)+`</b>:</p>
			 <p>Su postulación <b>`+html.EscapeString(p.Codigo)+`</b> fue enviada correctamente.</p>
			 <p>Adjuntamos el <b>formulario de postulación</b> con la información registrada. Imprímalo, fírmelo y
			    entréguelo junto con los requisitos en el lugar, fecha y horario que establezca el IESS.</p>`),
		Adjuntos: []Adjunto{{Nombre: "Formulario-" + p.Codigo + ".pdf", Tipo: "application/pdf", Datos: formularioPDF}},
	})
}

func plantillaCorreo(titulo, cuerpo string) string {
	return `<!doctype html><html><body style="font-family:Arial,sans-serif;color:#1f2937;max-width:560px">
<h2 style="color:#1d4ed8">` + html.EscapeString(titulo) + `</h2>` + cuerpo + `
<hr style="border:none;border-top:1px solid #e5e7eb"><p style="font-size:12px;color:#6b7280">
Concurso Público de Méritos y Oposición - Directorio del BIESS. Este es un correo automático, no responda a esta dirección.</p>
</body></html>`
}
