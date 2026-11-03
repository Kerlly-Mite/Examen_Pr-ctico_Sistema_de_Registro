package handlers

import (
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"

	"biess-backend/middleware"
	"biess-backend/models"
	"biess-backend/services"
)

// PostulacionHandler atiende las rutas /api/postulacion/* (EF-12). Todas requieren sesión.
type PostulacionHandler struct {
	svc services.PostulacionService
}

// NewPostulacionHandler recibe el service como interfaz (Inyección de Dependencias).
func NewPostulacionHandler(svc services.PostulacionService) *PostulacionHandler {
	return &PostulacionHandler{svc: svc}
}

// Obtener responde GET /api/postulacion: la postulación de quien tiene la sesión.
func (h *PostulacionHandler) Obtener(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	p, err := h.svc.Obtener(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(p)
}

// Guardar responde PUT /api/postulacion: guarda el formulario completo en una transacción.
func (h *PostulacionHandler) Guardar(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	var req models.PostulacionReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	p, err := h.svc.Guardar(c.UserContext(), id, req)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(p)
}

// VerificarCargo responde GET /api/postulacion/verificar-cargo/:codigo (1.4.x): muestra al instante
// "Postulante afiliado al IESS" o el mensaje de rechazo, antes de guardar.
func (h *PostulacionHandler) VerificarCargo(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	v, err := h.svc.VerificarCargo(c.UserContext(), id, c.Params("codigo"))
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(v)
}

// Enviar responde POST /api/postulacion/enviar (1.10): valida que esté completa y la marca como ENVIADA.
func (h *PostulacionHandler) Enviar(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	var req models.EnviarReq
	if err := c.BodyParser(&req); err != nil {
		return responderError(c, errCuerpoInvalido)
	}
	p, err := h.svc.Enviar(c.UserContext(), id, req)
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(p)
}

// SubirDocumento responde POST /api/postulacion/documentos (multipart/form-data):
// campo "archivo" (el PDF) y campo "tipo" ("respaldo" u "organizacion").
func (h *PostulacionHandler) SubirDocumento(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	archivo, err := c.FormFile("archivo")
	if err != nil {
		return responderError(c, models.ErrSolicitudInvalida("Adjunte el PDF en el campo \"archivo\""))
	}
	f, err := archivo.Open()
	if err != nil {
		return responderError(c, models.ErrInterno(err))
	}
	defer f.Close()
	// Se lee como máximo 1 byte más que el límite mayor: si llega a ese tamaño, el service lo rechaza.
	datos, err := io.ReadAll(io.LimitReader(f, 4<<20+1))
	if err != nil {
		return responderError(c, models.ErrInterno(err))
	}
	doc, err := h.svc.SubirDocumento(c.UserContext(), id, c.FormValue("tipo"), archivo.Filename, datos)
	if err != nil {
		return responderError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(doc)
}

// Documento responde GET /api/postulacion/documentos/:id: descarga un PDF propio.
func (h *PostulacionHandler) Documento(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	docID := c.Params("id")
	datos, err := h.svc.Documento(c.UserContext(), id, docID)
	if err != nil {
		return responderError(c, err)
	}
	return enviarPDF(c, datos, nombreArchivo(docID))
}

// FormularioPDF responde GET /api/postulacion/formulario.pdf: el formulario para imprimir y firmar (1.10.2).
func (h *PostulacionHandler) FormularioPDF(c *fiber.Ctx) error {
	id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
	datos, codigo, err := h.svc.FormularioPDF(c.UserContext(), id)
	if err != nil {
		return responderError(c, err)
	}
	return enviarPDF(c, datos, "Formulario-"+codigo+".pdf")
}

func enviarPDF(c *fiber.Ctx, datos []byte, nombre string) error {
	c.Set(fiber.HeaderContentType, "application/pdf")
	c.Set(fiber.HeaderContentDisposition, `inline; filename="`+nombre+`"`)
	c.Set("X-Content-Type-Options", "nosniff")
	return c.Send(datos)
}

// nombreArchivo quita el prefijo aleatorio del id ("<aleatorio>__nombre.pdf" -> "nombre.pdf").
func nombreArchivo(id string) string {
	if i := strings.Index(id, "__"); i >= 0 {
		return id[i+2:]
	}
	return id
}

// Opciones responde GET /api/postulacion/opciones: cargos y preguntas de inhabilidades.
func (h *PostulacionHandler) Opciones(c *fiber.Ctx) error {
	op, err := h.svc.Opciones(c.UserContext())
	if err != nil {
		return responderError(c, err)
	}
	return c.JSON(op)
}
