package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Mensaje es un correo listo para enviar: cuerpo en HTML y texto plano, y adjuntos opcionales.
type Mensaje struct {
	Para     string
	Asunto   string
	Texto    string
	HTML     string
	Adjuntos []Adjunto
}

// Adjunto es un archivo que viaja con el correo (p. ej. el formulario de postulación en PDF).
type Adjunto struct {
	Nombre string
	Tipo   string // tipo MIME, p. ej. "application/pdf"
	Datos  []byte
}

// EnviadorCorreo envía un Mensaje. Hay dos implementaciones: SMTP (real) y consola (sin configuración).
type EnviadorCorreo interface {
	Enviar(ctx context.Context, m Mensaje) error
}

// ConfigSMTP son los datos del servidor de correo (variables SMTP_* del .env).
type ConfigSMTP struct {
	Host      string // p. ej. smtp.gmail.com
	Puerto    int    // 587 (STARTTLS) o 465 (TLS directo)
	Usuario   string
	Clave     string // en Gmail: contraseña de aplicación, nunca la contraseña normal
	Remitente string // dirección "From"; si está vacía se usa Usuario
	Nombre    string // nombre visible del remitente
}

// NewEnviadorCorreo elige la implementación: si hay SMTP_HOST configurado envía correos REALES por SMTP;
// si no, los escribe en la consola para que el proyecto funcione en cualquier computadora.
func NewEnviadorCorreo(cfg ConfigSMTP) EnviadorCorreo {
	if strings.TrimSpace(cfg.Host) == "" {
		log.Println("[correo] SMTP_HOST no está configurado: los correos se mostrarán en la consola")
		return enviadorConsola{}
	}
	if cfg.Remitente == "" {
		cfg.Remitente = cfg.Usuario
	}
	if cfg.Puerto == 0 {
		cfg.Puerto = 587
	}
	log.Printf("[correo] Envío real por SMTP: %s:%d como %s", cfg.Host, cfg.Puerto, cfg.Remitente)
	return &enviadorSMTP{cfg: cfg}
}

// ---------------------------------------------------------------------------------------------
// SMTP real
// ---------------------------------------------------------------------------------------------

type enviadorSMTP struct {
	cfg ConfigSMTP
}

const timeoutSMTP = 20 * time.Second

func (e *enviadorSMTP) Enviar(ctx context.Context, m Mensaje) error {
	datos, err := construirMIME(e.cfg.Remitente, e.cfg.Nombre, m)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeoutSMTP)
	defer cancel()
	direccion := net.JoinHostPort(e.cfg.Host, fmt.Sprint(e.cfg.Puerto))
	tlsCfg := &tls.Config{ServerName: e.cfg.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	dialer := &net.Dialer{}
	if e.cfg.Puerto == 465 { // TLS directo (SMTPS)
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: tlsCfg}).DialContext(ctx, "tcp", direccion)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", direccion)
	}
	if err != nil {
		return fmt.Errorf("conectar a %s: %w", direccion, err)
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	c, err := smtp.NewClient(conn, e.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("iniciar SMTP: %w", err)
	}
	defer c.Close()

	// En el puerto 587 la conexión empieza sin cifrar y se cifra con STARTTLS antes de enviar la clave.
	if ok, _ := c.Extension("STARTTLS"); ok && e.cfg.Puerto != 465 {
		if err := c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("STARTTLS: %w", err)
		}
	}
	if e.cfg.Usuario != "" {
		if err := c.Auth(smtp.PlainAuth("", e.cfg.Usuario, e.cfg.Clave, e.cfg.Host)); err != nil {
			return fmt.Errorf("autenticación SMTP (revise SMTP_USER y SMTP_PASSWORD): %w", err)
		}
	}
	if err := c.Mail(e.cfg.Remitente); err != nil {
		return fmt.Errorf("remitente: %w", err)
	}
	if err := c.Rcpt(m.Para); err != nil {
		return fmt.Errorf("destinatario %s: %w", m.Para, err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := w.Write(datos); err != nil {
		return fmt.Errorf("escribir mensaje: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("cerrar mensaje: %w", err)
	}
	return c.Quit()
}

// construirMIME arma el correo en formato MIME:
//
//	multipart/mixed
//	├── multipart/alternative (texto plano + HTML: el cliente de correo muestra el que soporte)
//	└── adjuntos en base64
func construirMIME(remitente, nombre string, m Mensaje) ([]byte, error) {
	if strings.ContainsAny(m.Para, "\r\n") || strings.ContainsAny(m.Asunto, "\r\n") {
		return nil, errors.New("encabezado de correo inválido") // evita inyección de encabezados
	}
	limiteMixto, limiteAlt := limiteMIME(), limiteMIME()
	var b bytes.Buffer

	from := remitente
	if nombre != "" {
		from = mime.QEncoding.Encode("utf-8", nombre) + " <" + remitente + ">"
	}
	dominio := remitente[strings.LastIndex(remitente, "@")+1:]
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", m.Para)
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", m.Asunto))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintf(&b, "Message-ID: <%s@%s>\r\n", limiteMIME(), dominio)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", limiteMixto)

	fmt.Fprintf(&b, "--%s\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n", limiteMixto, limiteAlt)
	parteBase64(&b, limiteAlt, "text/plain; charset=utf-8", "", []byte(m.Texto))
	parteBase64(&b, limiteAlt, "text/html; charset=utf-8", "", []byte(m.HTML))
	fmt.Fprintf(&b, "--%s--\r\n", limiteAlt)

	for _, a := range m.Adjuntos {
		disposicion := fmt.Sprintf("attachment; filename=%q", a.Nombre)
		parteBase64(&b, limiteMixto, a.Tipo, disposicion, a.Datos)
	}
	fmt.Fprintf(&b, "--%s--\r\n", limiteMixto)
	return b.Bytes(), nil
}

func parteBase64(b *bytes.Buffer, limite, tipo, disposicion string, datos []byte) {
	fmt.Fprintf(b, "--%s\r\nContent-Type: %s\r\nContent-Transfer-Encoding: base64\r\n", limite, tipo)
	if disposicion != "" {
		fmt.Fprintf(b, "Content-Disposition: %s\r\n", disposicion)
	}
	b.WriteString("\r\n")
	codificado := base64.StdEncoding.EncodeToString(datos)
	for len(codificado) > 76 { // las líneas de un correo no deben superar 78 caracteres
		b.WriteString(codificado[:76] + "\r\n")
		codificado = codificado[76:]
	}
	b.WriteString(codificado + "\r\n")
}

func limiteMIME() string {
	var r [12]byte
	_, _ = rand.Read(r[:])
	return "biess-" + hex.EncodeToString(r[:])
}

// ---------------------------------------------------------------------------------------------
// Consola (sin SMTP configurado)
// ---------------------------------------------------------------------------------------------

type enviadorConsola struct{}

func (enviadorConsola) Enviar(_ context.Context, m Mensaje) error {
	adjuntos := make([]string, len(m.Adjuntos))
	for i, a := range m.Adjuntos {
		adjuntos[i] = fmt.Sprintf("%s (%d KB)", a.Nombre, len(a.Datos)/1024)
	}
	log.Printf("[CORREO EN CONSOLA: configure SMTP_HOST para enviarlo de verdad] Para: %s | Asunto: %s\n%s\nAdjuntos: %s",
		m.Para, m.Asunto, m.Texto, strings.Join(adjuntos, ", "))
	return nil
}
