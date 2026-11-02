package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"biess-backend/models"
)

// servidorSMTPPrueba es un servidor SMTP mínimo en memoria: habla el protocolo de verdad (EHLO, AUTH, MAIL,
// RCPT, DATA) y guarda el mensaje recibido. Permite probar el envío real sin depender de Gmail ni de internet.
func servidorSMTPPrueba(t *testing.T) (puerto int, recibido <-chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	canal := make(chan string, 1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		responder := func(s string) { io.WriteString(conn, s+"\r\n") }
		responder("220 prueba ESMTP")
		var datos strings.Builder
		for {
			linea, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(linea))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				responder("250-prueba\r\n250 AUTH PLAIN")
			case strings.HasPrefix(cmd, "AUTH"):
				responder("235 autenticado")
			case strings.HasPrefix(cmd, "MAIL"), strings.HasPrefix(cmd, "RCPT"):
				responder("250 ok")
			case cmd == "DATA":
				responder("354 envie el mensaje")
				for {
					l, err := r.ReadString('\n')
					if err != nil || l == ".\r\n" {
						break
					}
					datos.WriteString(l)
				}
				canal <- datos.String()
				responder("250 recibido")
			case cmd == "QUIT":
				responder("221 adios")
				return
			default:
				responder("250 ok")
			}
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, canal
}

func TestCorreoSMTPConAdjunto(t *testing.T) {
	puerto, recibido := servidorSMTPPrueba(t)
	enviador := NewEnviadorCorreo(ConfigSMTP{
		Host: "localhost", Puerto: puerto, Usuario: "biess@correo.com", Clave: "clave", Nombre: "Postulación BIESS",
	})
	p := &models.Postulacion{Codigo: "BIESS-2026-000009", Nombres: "LUCIA", Apellidos: "MORA", Correo1: "lucia@correo.com"}
	pdf := []byte("%PDF-1.4 contenido %%EOF")

	if err := NewNotificadorCorreo(enviador).EnviarConfirmacion(context.Background(), p, pdf); err != nil {
		t.Fatalf("el envío por SMTP falló: %v", err)
	}

	var crudo string
	select {
	case crudo = <-recibido:
	case <-time.After(5 * time.Second):
		t.Fatal("el servidor SMTP no recibió el mensaje")
	}

	// El mensaje recibido debe ser un correo MIME válido, con el asunto en UTF-8 y el PDF adjunto intacto.
	msg, err := mail.ReadMessage(strings.NewReader(crudo))
	if err != nil {
		t.Fatalf("mensaje MIME inválido: %v", err)
	}
	asunto, _ := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	if msg.Header.Get("To") != "lucia@correo.com" || !strings.Contains(asunto, "BIESS-2026-000009") {
		t.Errorf("encabezados incorrectos: To=%q Subject=%q", msg.Header.Get("To"), asunto)
	}
	_, params, _ := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	partes := multipart.NewReader(msg.Body, params["boundary"])
	var adjunto []byte
	for {
		parte, err := partes.NextPart()
		if err != nil {
			break
		}
		if parte.FileName() == "Formulario-BIESS-2026-000009.pdf" {
			adjunto, _ = io.ReadAll(base64.NewDecoder(base64.StdEncoding, parte))
		}
	}
	if !bytes.Equal(adjunto, pdf) {
		t.Errorf("el PDF adjunto no llegó intacto (%d bytes)", len(adjunto))
	}
}

func TestCorreoRechazaInyeccionDeEncabezados(t *testing.T) {
	if _, err := construirMIME("a@b.com", "", Mensaje{Para: "x@y.com\r\nBcc: espia@z.com", Asunto: "hola"}); err == nil {
		t.Error("un destinatario con salto de línea debe rechazarse")
	}
}

func TestEsContenidoPDF(t *testing.T) {
	casos := map[string]bool{
		"%PDF-1.7\n...contenido...\n%%EOF\n": true,
		"MZ\x90\x00 ejecutable renombrado":   false, // un .exe con extensión .pdf
		"%PDF-1.7 sin final":                 false,
		"":                                   false,
	}
	for contenido, esperado := range casos {
		if esContenidoPDF([]byte(contenido)) != esperado {
			t.Errorf("esContenidoPDF(%q) debería ser %v", contenido, esperado)
		}
	}
}

func TestFormularioPDFSeGenera(t *testing.T) {
	no := false
	p := &models.Postulacion{
		Codigo: "BIESS-2026-000001", Estado: "ENVIADA", Cedula: "0104567896", Nombres: "JOSÉ LUIS", Apellidos: "RAMÍREZ",
		Nacionalidad: "ECUATORIANA", EstadoCivil: "SOLTERO", Correo1: "jose@correo.com", FechaNacimiento: fecha("1990-01-25"),
	}
	p.CargoCodigo, p.RespaldadoOrganizacion = "AFILIADOS", &no
	p.Experiencia = []models.Experiencia{{Tipo: "DIRECCION", Cargo: "Gerente", Institucion: "Cooperativa",
		FechaInicio: fecha("2015-01-01"), FechaFin: fecha("2019-12-31"), Actividades: "FINANZAS",
		DescripcionFunciones: "Dirección financiera", DocumentoRespaldo: strings.Repeat("a", 32) + "__certificado.pdf"}}
	pdf, err := generarFormularioPDF(p, datosFormulario{Pais: "Ecuador", Provincia: "Azuay", Ciudad: "Cuenca",
		Preguntas: []models.PreguntaInhabilidad{{Codigo: "1.9.1", Texto: "¿Es Ecuatoriano/a?", RespuestaHabilitante: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if !esContenidoPDF(pdf) || len(pdf) < 1500 {
		t.Errorf("el formulario generado no es un PDF válido (%d bytes)", len(pdf))
	}
}
