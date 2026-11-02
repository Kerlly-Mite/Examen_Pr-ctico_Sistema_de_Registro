package services

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"biess-backend/models"
)

// datosFormulario son los textos que el PDF necesita y que no están en la postulación (nombres de catálogos).
type datosFormulario struct {
	Pais, Provincia, Ciudad string
	Cargo                   string
	Preguntas               []models.PreguntaInhabilidad
}

// textoConsentimiento es el texto de 1.10.1 del requerimiento funcional.
var textoConsentimiento = []string{
	"1. Autorizo de manera libre y voluntaria, emitiendo consentimiento explícito al Instituto Ecuatoriano de Seguridad " +
		"Social, para el uso de mis datos personales dentro del concurso de méritos y oposición para la selección de los " +
		"Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en " +
		"Representación de los Afiliados, Jubilados y Empleadores, de conformidad con lo prescrito en el artículo 7 de la " +
		"Ley Orgánica de Protección de Datos Personales.",
	"2. Declaro que la información del presente formulario y documentación adjunta corresponde a la verdad y autorizo al " +
		"Instituto Ecuatoriano de Seguridad Social para que se compruebe ante cualquier organismo público o privado su veracidad.",
	"3. Acepto expresamente cumplir con las normas aplicables al concurso de méritos y oposición para la selección de los " +
		"Miembros Principales y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en " +
		"Representación de los Afiliados, Jubilados y Empleadores; así como las resoluciones y disposiciones impartidas por el " +
		"Instituto Ecuatoriano de Seguridad Social. En caso de incumplimiento de requisitos o encontrarme inmerso/a en las " +
		"prohibiciones o inhabilidades acepto mi descalificación del proceso.",
	"4. Acepto que mi nombre como postulante al concurso de méritos y oposición para la selección de los Miembros Principales " +
		"y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en Representación de los " +
		"Afiliados, Jubilados y Empleadores; sea publicado en la página web del IESS - BIESS.",
	"5. El presente formulario, debe ser impreso y con firma manuscrita, entregado junto con los requisitos solicitados, en " +
		"el lugar, fecha y horario establecido por el Instituto Ecuatoriano de Seguridad Social.",
}

// etiquetas traduce los códigos guardados en la base a los textos de las listas del formulario.
var etiquetas = map[string]string{
	"MASCULINO": "Masculino", "FEMENINO": "Femenino",
	"SOLTERO": "Soltero/a", "CASADO": "Casado/a", "UNION_DE_HECHO": "Unión de hecho", "DIVORCIADO": "Divorciado/a", "VIUDO": "Viudo/a",
	"CONVIVIENTE": "Conviviente", "CONYUGE": "Cónyuge", "CUNADO": "Cuñado/a", "HERMANO": "Hermano/a", "HIJO_POLITICO": "Hijo político",
	"HIJO": "Hijo/a", "MADRE": "Madre", "NIETO": "Nieto/a", "NUERA": "Nuera", "PADRE": "Padre", "PRIMO_HERMANO": "Primo hermano/a",
	"SOBRINO": "Sobrino/a", "SUEGRO": "Suegro/a", "TIO": "Tío/a", "YERNO": "Yerno",
	"TERCER_NIVEL": "Tercer Nivel de Grado", "CUARTO_ESPECIALIDAD": "Cuarto Nivel - Especialidad Académica",
	"CUARTO_MAESTRIA": "Cuarto Nivel - Maestría Académica", "CUARTO_PHD": "Cuarto Nivel PhD - Doctorado",
	"NACIONAL": "Nacional", "EXTRANJERA": "Extranjera",
	"FINANZAS": "Finanzas", "MERCADO_CAPITALES": "Mercado de Capitales", "ADMINISTRACION": "Administración de Empresas",
	"ECONOMIA": "Economía", "DERECHO": "Derecho", "AFINES": "Materias afines",
	"GENERAL": "General", "DIRECCION": "Dirección",
	"PUBLICACION": "Publicación", "TRABAJO_INVESTIGACION": "Trabajo de investigación",
	"CURSO": "Curso", "SEMINARIO": "Seminario", "TALLER": "Taller",
}

func etiqueta(codigo string) string {
	if e, ok := etiquetas[codigo]; ok {
		return e
	}
	return codigo
}

func siNo(b bool) string {
	if b {
		return "Sí"
	}
	return "No"
}

func fechaTexto(f *models.Fecha) string {
	if f == nil {
		return ""
	}
	return f.Format("02/01/2006")
}

// nombreDocumento muestra el nombre original del PDF (el id guardado es "<aleatorio>__<nombre>.pdf").
func nombreDocumento(id string) string {
	if i := strings.Index(id, "__"); i >= 0 {
		return id[i+2:]
	}
	return id
}

func duracionTexto(d *models.Duracion) string {
	if d == nil {
		return ""
	}
	return fmt.Sprintf("%d años, %d meses, %d días", d.Anios, d.Meses, d.Dias)
}

// formularioPDF envuelve fpdf con ayudas para escribir el formulario en español (codificación cp1252).
type formularioPDF struct {
	pdf *fpdf.Fpdf
	tr  func(string) string
}

const anchoUtil = 180.0 // A4 (210 mm) menos márgenes de 15 mm

func (f *formularioPDF) seccion(titulo string) {
	if f.pdf.GetY() > 255 {
		f.pdf.AddPage()
	}
	f.pdf.Ln(3)
	f.pdf.SetFont("Helvetica", "B", 10.5)
	f.pdf.SetFillColor(29, 78, 216)
	f.pdf.SetTextColor(255, 255, 255)
	f.pdf.CellFormat(anchoUtil, 7, f.tr(titulo), "", 1, "L", true, 0, "")
	f.pdf.SetTextColor(31, 41, 55)
	f.pdf.Ln(1)
}

func (f *formularioPDF) campo(etiqueta, valor string) {
	if strings.TrimSpace(valor) == "" {
		valor = "-"
	}
	f.pdf.SetFont("Helvetica", "B", 9)
	f.pdf.CellFormat(55, 5.5, f.tr(etiqueta+":"), "", 0, "L", false, 0, "")
	f.pdf.SetFont("Helvetica", "", 9)
	f.pdf.MultiCell(anchoUtil-55, 5.5, f.tr(valor), "", "L", false)
}

func (f *formularioPDF) subtitulo(texto string) {
	f.pdf.Ln(1)
	f.pdf.SetFont("Helvetica", "BI", 9.5)
	f.pdf.CellFormat(anchoUtil, 6, f.tr(texto), "B", 1, "L", false, 0, "")
}

func (f *formularioPDF) parrafo(texto string, tam float64) {
	f.pdf.SetFont("Helvetica", "", tam)
	f.pdf.MultiCell(anchoUtil, 4.6, f.tr(texto), "", "J", false)
	f.pdf.Ln(1)
}

func (f *formularioPDF) ninguno() {
	f.pdf.SetFont("Helvetica", "I", 9)
	f.pdf.CellFormat(anchoUtil, 5.5, f.tr("No registra información."), "", 1, "L", false, 0, "")
}

// generarFormularioPDF arma el formulario de postulación completo (1.1 a 1.10) listo para imprimir y firmar.
func generarFormularioPDF(p *models.Postulacion, d datosFormulario) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 18)
	pdf.SetTitle("Formulario de postulación "+p.Codigo, true)
	pdf.AliasNbPages("")
	f := &formularioPDF{pdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}
	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(107, 114, 128)
		pdf.CellFormat(0, 5, f.tr(fmt.Sprintf("Formulario %s  -  Página %d de {nb}", p.Codigo, pdf.PageNo())), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	// Encabezado
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(anchoUtil, 6, f.tr("INSTITUTO ECUATORIANO DE SEGURIDAD SOCIAL"), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(anchoUtil, 8, f.tr("FORMULARIO DE POSTULACIÓN"), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.MultiCell(anchoUtil, 4.6, f.tr("Concurso Público de Méritos y Oposición para la Designación de los Miembros Principales "+
		"y Alternos del Directorio del BIESS, en representación de los Afiliados, Jubilados y Empleadores"), "", "C", false)
	pdf.Ln(2)
	enviada := "Pendiente de envío"
	if p.EnviadaEn != nil {
		enviada = p.EnviadaEn.Format("02/01/2006 15:04")
	}
	f.campo("Código de formulario", p.Codigo)
	f.campo("Estado", p.Estado)
	f.campo("Fecha de envío", enviada)

	f.seccion("1.1 Información personal")
	f.campo("Cédula", p.Cedula)
	f.campo("Nombres", p.Nombres)
	f.campo("Apellidos", p.Apellidos)
	f.campo("Género", etiqueta(p.Genero))
	f.campo("Fecha de nacimiento", fechaTexto(p.FechaNacimiento))
	f.campo("Nacionalidad", p.Nacionalidad)
	f.campo("Estado civil", etiqueta(p.EstadoCivil))
	if p.ParejaNombres != "" {
		f.campo("Nombres y apellidos de la pareja", p.ParejaNombres)
		f.campo("Cédula de la pareja", p.ParejaCedula)
	}

	f.seccion("1.2 Información domiciliaria y de contacto")
	f.campo("País", d.Pais)
	f.campo("Provincia / Estado", d.Provincia)
	f.campo("Ciudad", d.Ciudad)
	f.campo("Teléfono celular", p.TelefonoCelular)
	f.campo("Correo electrónico 1", p.Correo1)
	f.campo("Correo electrónico 2", p.Correo2)
	f.campo("Calle principal", p.CallePrincipal)
	f.campo("Número", p.Numero)
	f.campo("Calle secundaria", p.CalleSecundaria)
	f.campo("Sector", p.Sector)
	f.campo("Referencia", p.Referencia)
	f.campo("Teléfono de domicilio", p.TelefonoDomicilio)

	f.seccion("1.3 Información adicional")
	f.campo("Familiares en el IESS o BIESS", siNo(p.TieneFamiliaresIESS))
	for i, fam := range p.Familiares {
		f.subtitulo(fmt.Sprintf("Familiar %d", i+1))
		f.campo("Nombres y apellidos", fam.NombresApellidos)
		f.campo("Parentesco", etiqueta(fam.Parentesco))
		f.campo("Institución", fam.Institucion)
		f.campo("Área", fam.Area)
		f.campo("Cargo", fam.Cargo)
	}

	f.seccion("1.4 Cargo al que postula")
	f.campo("Cargo", d.Cargo)
	if p.VerificacionCargo != nil {
		f.campo("Historial Laboral IESS", p.VerificacionCargo.Estado)
	}
	if p.RespaldadoOrganizacion != nil {
		f.campo("Respaldado por organización", siNo(*p.RespaldadoOrganizacion))
	}
	if p.OrganizacionNombre != "" {
		f.campo("Nombre de la organización", p.OrganizacionNombre)
		f.campo("Documento de respaldo", nombreDocumento(p.OrganizacionDocumento))
	}

	f.seccion("1.5 Formación académica")
	if len(p.Formacion) == 0 {
		f.ninguno()
	}
	for i, x := range p.Formacion {
		f.subtitulo(fmt.Sprintf("Formación %d", i+1))
		f.campo("Nivel de instrucción", etiqueta(x.NivelInstruccion))
		f.campo("Institución educativa", x.Institucion)
		f.campo("Tipo", etiqueta(x.Tipo))
		f.campo("Título obtenido", x.Titulo)
		f.campo("Área de estudios", etiqueta(x.AreaEstudios))
		f.campo("Registro SENESCYT", x.RegistroSenescyt+"  ("+fechaTexto(x.FechaRegistroSenescyt)+")")
		f.campo("Documento de respaldo", nombreDocumento(x.DocumentoRespaldo))
	}

	f.seccion("1.6 Experiencia profesional")
	if len(p.Experiencia) == 0 {
		f.ninguno()
	}
	for i, x := range p.Experiencia {
		f.subtitulo(fmt.Sprintf("Experiencia %d", i+1))
		f.campo("Tipo de experiencia", etiqueta(x.Tipo))
		f.campo("Cargo o puesto", x.Cargo)
		f.campo("Institución o empresa", x.Institucion)
		f.campo("Periodo", fechaTexto(x.FechaInicio)+" al "+fechaTexto(x.FechaFin))
		f.campo("Tiempo", duracionTexto(x.Duracion))
		f.campo("Actividades relacionadas a", etiqueta(x.Actividades))
		f.campo("Descripción de funciones", x.DescripcionFunciones)
		f.campo("Documento de respaldo", nombreDocumento(x.DocumentoRespaldo))
	}
	if len(p.Experiencia) > 0 {
		f.pdf.Ln(1)
		f.campo("Experiencia total", duracionTexto(p.ExperienciaTotal))
	}

	f.seccion("1.7 Publicaciones e investigaciones")
	if len(p.Publicaciones) == 0 {
		f.ninguno()
	}
	for i, x := range p.Publicaciones {
		f.subtitulo(fmt.Sprintf("Publicación %d", i+1))
		f.campo("Título", x.Titulo)
		f.campo("Tipo", etiqueta(x.Tipo))
		f.campo("Medio de publicación", x.MedioPublicacion)
		f.campo("Fecha de publicación", fechaTexto(x.FechaPublicacion))
		f.campo("Relación con su formación", siNo(x.RelacionadaFormacion))
		f.campo("Documento de respaldo", nombreDocumento(x.DocumentoRespaldo))
	}

	f.seccion("1.8 Capacitación")
	if len(p.Capacitaciones) == 0 {
		f.ninguno()
	}
	for i, x := range p.Capacitaciones {
		f.subtitulo(fmt.Sprintf("Evento %d", i+1))
		f.campo("Tipo de evento", etiqueta(x.TipoEvento))
		f.campo("Nombre del evento", x.NombreEvento)
		f.campo("Institución", x.Institucion)
		f.campo("Periodo", fechaTexto(x.FechaInicio)+" al "+fechaTexto(x.FechaFin))
		f.campo("Número de horas", fmt.Sprint(x.Horas))
		f.campo("Documento de respaldo", nombreDocumento(x.DocumentoRespaldo))
	}

	f.seccion("1.9 Inhabilidades")
	respuestas := map[string]models.RespuestaInhabilidad{}
	for _, r := range p.Inhabilidades {
		respuestas[r.Codigo] = r
	}
	for _, q := range d.Preguntas {
		r, ok := respuestas[q.Codigo]
		resp := "Sin responder"
		if ok {
			resp = siNo(r.Respuesta)
		}
		pdf.SetFont("Helvetica", "", 8.5)
		y := pdf.GetY()
		pdf.MultiCell(anchoUtil-22, 4.4, f.tr(q.Codigo+"  "+q.Texto), "", "L", false)
		yFin := pdf.GetY()
		if yFin > y { // misma página: la respuesta va a la derecha de la pregunta
			pdf.SetXY(15+anchoUtil-22, y)
			pdf.SetFont("Helvetica", "B", 8.5)
			pdf.CellFormat(22, 4.4, f.tr(resp), "", 0, "R", false, 0, "")
			pdf.SetXY(15, yFin)
		}
		if ok && r.Codigo == codigoPreguntaRemocion && r.Respuesta {
			f.campo("   Institución", r.RemocionInstitucion)
			f.campo("   Cargo", r.RemocionCargo)
			f.campo("   Fecha de finalización", fechaTexto(r.RemocionFechaFin))
			f.campo("   Motivo de finalización", r.RemocionMotivo)
		}
		pdf.Ln(0.8)
	}

	f.seccion("1.10 Autorización y declaración")
	for _, t := range textoConsentimiento {
		f.parrafo(t, 8.5)
	}

	// 1.10.2 Firma del postulante con nombres, apellidos y número de cédula.
	if pdf.GetY() > 240 {
		pdf.AddPage()
	}
	pdf.Ln(16)
	pdf.SetX(60)
	pdf.CellFormat(90, 0, "", "T", 1, "C", false, 0, "")
	pdf.Ln(1.5)
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(anchoUtil, 5, f.tr("Firma del postulante"), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(anchoUtil, 5, f.tr(p.Nombres+" "+p.Apellidos), "", 1, "C", false, 0, "")
	pdf.CellFormat(anchoUtil, 5, f.tr("C.I. "+p.Cedula), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generar PDF del formulario: %w", err)
	}
	return buf.Bytes(), nil
}
