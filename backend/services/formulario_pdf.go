package services

import (
	"bytes"
	"fmt"
	"strings"
	"time"

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

// Diseño institucional del formulario: medidas y colores.
const (
	anchoPagina  = 210.0 // A4
	altoPagina   = 297.0
	margen       = 15.0
	anchoUtil    = anchoPagina - 2*margen
	margenInf    = 22.0 // espacio reservado para el pie de página
	altoLineaVal = 4.2
)

type colorRGB struct{ r, g, b int }

var (
	colAzulOscuro = colorRGB{14, 47, 94}    // barras de sección y títulos
	colAzulClaro  = colorRGB{232, 237, 245} // subtítulos y encabezados de tabla
	colFilaPar    = colorRGB{246, 248, 251} // filas alternas de tablas
	colBorde      = colorRGB{185, 193, 206}
	colEtiqueta   = colorRGB{100, 110, 125}
	colTexto      = colorRGB{25, 32, 45}
	colBlanco     = colorRGB{255, 255, 255}
	colRojo       = colorRGB{185, 28, 28}
	// Franja superior con los colores de la bandera del Ecuador.
	colAmarillo = colorRGB{255, 209, 0}
	colAzul     = colorRGB{0, 56, 147}
	colRojoEC   = colorRGB{206, 17, 38}
)

// formularioPDF envuelve fpdf con ayudas para escribir el formulario en español (codificación cp1252).
type formularioPDF struct {
	pdf *fpdf.Fpdf
	tr  func(string) string
}

func (f *formularioPDF) relleno(c colorRGB) { f.pdf.SetFillColor(c.r, c.g, c.b) }
func (f *formularioPDF) texto(c colorRGB)   { f.pdf.SetTextColor(c.r, c.g, c.b) }
func (f *formularioPDF) trazo(c colorRGB)   { f.pdf.SetDrawColor(c.r, c.g, c.b) }

// partir traduce el texto a cp1252 y lo divide en líneas que caben en el ancho dado con la fuente actual.
// (fpdf.SplitText espera UTF-8 y falla con caracteres fuera de cp1252, por eso se mide ya traducido).
func (f *formularioPDF) partir(texto string, ancho float64) []string {
	var lineas []string
	actual := ""
	for _, palabra := range strings.Fields(f.tr(texto)) {
		for f.pdf.GetStringWidth(palabra) > ancho { // palabra más larga que la celda: se corta
			n := len(palabra) - 1
			for n > 1 && f.pdf.GetStringWidth(palabra[:n]) > ancho {
				n--
			}
			if actual != "" {
				lineas, actual = append(lineas, actual), ""
			}
			lineas, palabra = append(lineas, palabra[:n]), palabra[n:]
		}
		switch {
		case actual == "":
			actual = palabra
		case f.pdf.GetStringWidth(actual+" "+palabra) <= ancho:
			actual += " " + palabra
		default:
			lineas, actual = append(lineas, actual), palabra
		}
	}
	if actual != "" || len(lineas) == 0 {
		lineas = append(lineas, actual)
	}
	return lineas
}

// espacio salta de página si lo que viene (alto en mm) no cabe en la actual.
func (f *formularioPDF) espacio(alto float64) {
	if f.pdf.GetY()+alto > altoPagina-margenInf {
		f.pdf.AddPage()
	}
}

func (f *formularioPDF) franjaNacional() {
	f.relleno(colAmarillo)
	f.pdf.Rect(0, 0, anchoPagina/2, 2.5, "F")
	f.relleno(colAzul)
	f.pdf.Rect(anchoPagina/2, 0, anchoPagina/4, 2.5, "F")
	f.relleno(colRojoEC)
	f.pdf.Rect(anchoPagina*3/4, 0, anchoPagina/4, 2.5, "F")
}

// encabezado completo en la primera página y uno compacto en las siguientes.
func (f *formularioPDF) encabezado(codigo string) {
	pdf := f.pdf
	f.franjaNacional()
	if pdf.PageNo() > 1 {
		pdf.SetXY(margen, 7)
		pdf.SetFont("Helvetica", "B", 8)
		f.texto(colAzulOscuro)
		pdf.CellFormat(anchoUtil/2, 5, f.tr("INSTITUTO ECUATORIANO DE SEGURIDAD SOCIAL"), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		f.texto(colEtiqueta)
		pdf.CellFormat(anchoUtil/2, 5, f.tr("Formulario de postulación N.º "+codigo), "", 1, "R", false, 0, "")
		f.trazo(colAzulOscuro)
		pdf.SetLineWidth(0.4)
		pdf.Line(margen, 13, anchoPagina-margen, 13)
		pdf.SetLineWidth(0.2)
		pdf.SetY(17)
		return
	}

	pdf.SetXY(margen, 9)
	pdf.SetFont("Helvetica", "", 7.5)
	f.texto(colEtiqueta)
	pdf.CellFormat(120, 4, f.tr("REPÚBLICA DEL ECUADOR"), "", 2, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 12.5)
	f.texto(colAzulOscuro)
	pdf.CellFormat(120, 6.5, f.tr("INSTITUTO ECUATORIANO DE SEGURIDAD SOCIAL"), "", 2, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	f.texto(colTexto)
	pdf.CellFormat(120, 4.5, f.tr("Concurso Público de Méritos y Oposición - Directorio del BIESS"), "", 0, "L", false, 0, "")

	// Recuadro con el número de formulario.
	const wCaja = 48.0
	x := anchoPagina - margen - wCaja
	f.trazo(colAzulOscuro)
	pdf.SetLineWidth(0.4)
	pdf.Rect(x, 9, wCaja, 15, "D")
	f.relleno(colAzulOscuro)
	pdf.Rect(x, 9, wCaja, 5, "F")
	pdf.SetXY(x, 9)
	pdf.SetFont("Helvetica", "B", 7)
	f.texto(colBlanco)
	pdf.CellFormat(wCaja, 5, f.tr("FORMULARIO N.º"), "", 2, "C", false, 0, "")
	pdf.SetFont("Helvetica", "B", 11)
	f.texto(colAzulOscuro)
	pdf.CellFormat(wCaja, 10, f.tr(codigo), "", 0, "C", false, 0, "")

	pdf.SetLineWidth(0.6)
	pdf.Line(margen, 28, anchoPagina-margen, 28)
	f.trazo(colAmarillo)
	pdf.SetLineWidth(0.5)
	pdf.Line(margen, 29, anchoPagina-margen, 29)
	pdf.SetLineWidth(0.2)
	pdf.SetY(33)
}

func (f *formularioPDF) pie(codigo, generado string) {
	pdf := f.pdf
	pdf.SetY(-16)
	f.trazo(colBorde)
	pdf.SetLineWidth(0.2)
	pdf.Line(margen, pdf.GetY(), anchoPagina-margen, pdf.GetY())
	pdf.Ln(1.5)
	pdf.SetFont("Helvetica", "", 7)
	f.texto(colEtiqueta)
	pdf.CellFormat(anchoUtil*0.75, 3.8, f.tr("Formulario de postulación N.º "+codigo+"  ·  Generado electrónicamente el "+generado),
		"", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 7)
	pdf.CellFormat(anchoUtil*0.25, 3.8, f.tr(fmt.Sprintf("Página %d de {nb}", pdf.PageNo())), "", 1, "R", false, 0, "")
	pdf.SetFont("Helvetica", "I", 6.5)
	pdf.CellFormat(anchoUtil, 3.5, f.tr("Documento válido únicamente con la firma manuscrita del postulante."), "", 0, "L", false, 0, "")
}

// seccion dibuja la barra azul con el número y el nombre de la sección.
func (f *formularioPDF) seccion(numero, titulo string) {
	pdf := f.pdf
	pdf.Ln(4)
	f.espacio(25) // evita que el título quede solo al final de la página
	y := pdf.GetY()
	f.relleno(colAzulOscuro)
	pdf.Rect(margen, y, anchoUtil, 7, "F")
	f.relleno(colAmarillo)
	pdf.Rect(margen, y, 1.5, 7, "F")
	pdf.SetXY(margen+4, y)
	pdf.SetFont("Helvetica", "B", 9.5)
	f.texto(colBlanco)
	pdf.CellFormat(anchoUtil-4, 7, f.tr(numero+"   "+strings.ToUpper(titulo)), "", 1, "L", false, 0, "")
	pdf.Ln(1.5)
}

// fila dibuja celdas con borde (etiqueta pequeña arriba y el dato debajo) que se reparten el ancho por igual.
func (f *formularioPDF) fila(campos ...[2]string) {
	pdf := f.pdf
	ancho := anchoUtil / float64(len(campos))
	pdf.SetFont("Helvetica", "", 9)
	lineas := make([][]string, len(campos))
	alto := 0.0
	for i, c := range campos {
		valor := strings.TrimSpace(c[1])
		if valor == "" {
			valor = "—"
		}
		lineas[i] = f.partir(valor, ancho-4)
		if h := 6.6 + float64(len(lineas[i]))*altoLineaVal; h > alto {
			alto = h
		}
	}
	f.espacio(alto)
	y := pdf.GetY()
	f.trazo(colBorde)
	for i, c := range campos {
		x := margen + float64(i)*ancho
		pdf.Rect(x, y, ancho, alto, "D")
		pdf.SetXY(x+2, y+1.2)
		pdf.SetFont("Helvetica", "B", 6.3)
		f.texto(colEtiqueta)
		pdf.CellFormat(ancho-4, 3.5, f.tr(strings.ToUpper(c[0])), "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		f.texto(colTexto)
		for j, l := range lineas[i] {
			pdf.SetXY(x+2, y+5+float64(j)*altoLineaVal)
			pdf.CellFormat(ancho-4, altoLineaVal, l, "", 0, "L", false, 0, "")
		}
	}
	pdf.SetXY(margen, y+alto)
}

// subtitulo encabeza cada registro de una lista (Formación 1, Experiencia 2, ...).
func (f *formularioPDF) subtitulo(texto string) {
	f.pdf.Ln(1.5)
	f.espacio(20)
	f.relleno(colAzulClaro)
	f.trazo(colBorde)
	f.pdf.SetFont("Helvetica", "B", 8.5)
	f.texto(colAzulOscuro)
	f.pdf.CellFormat(anchoUtil, 6, f.tr("  "+texto), "1", 1, "L", true, 0, "")
}

func (f *formularioPDF) ninguno() {
	f.trazo(colBorde)
	f.pdf.SetFont("Helvetica", "I", 8.5)
	f.texto(colEtiqueta)
	f.pdf.CellFormat(anchoUtil, 7, f.tr("  No registra información en esta sección."), "1", 1, "L", false, 0, "")
}

// tablaInhabilidades muestra las preguntas 1.9.x con su respuesta en una tabla con filas alternas.
func (f *formularioPDF) tablaInhabilidades(p *models.Postulacion, preguntas []models.PreguntaInhabilidad) {
	pdf := f.pdf
	const wNum, wResp, altoLinea = 14.0, 26.0, 3.9
	wPreg := anchoUtil - wNum - wResp
	cabecera := func() {
		f.relleno(colAzulClaro)
		f.trazo(colBorde)
		pdf.SetFont("Helvetica", "B", 7.5)
		f.texto(colAzulOscuro)
		pdf.CellFormat(wNum, 6, f.tr("N.º"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(wPreg, 6, "  PREGUNTA", "1", 0, "L", true, 0, "")
		pdf.CellFormat(wResp, 6, "RESPUESTA", "1", 1, "C", true, 0, "")
	}
	cabecera()

	respuestas := map[string]models.RespuestaInhabilidad{}
	for _, r := range p.Inhabilidades {
		respuestas[r.Codigo] = r
	}
	for i, q := range preguntas {
		pdf.SetFont("Helvetica", "", 8)
		lineas := f.partir(q.Texto, wPreg-4)
		alto := float64(len(lineas))*altoLinea + 2.6
		if pdf.GetY()+alto > altoPagina-margenInf {
			pdf.AddPage()
			cabecera()
			pdf.SetFont("Helvetica", "", 8) // la cabecera deja la fuente en negrita
		}
		y := pdf.GetY()
		if i%2 == 1 {
			f.relleno(colFilaPar)
			pdf.Rect(margen, y, anchoUtil, alto, "F")
		}
		f.trazo(colBorde)
		pdf.Rect(margen, y, wNum, alto, "D")
		pdf.Rect(margen+wNum, y, wPreg, alto, "D")
		pdf.Rect(margen+wNum+wPreg, y, wResp, alto, "D")

		f.texto(colTexto)
		pdf.SetXY(margen, y)
		pdf.CellFormat(wNum, alto, q.Codigo, "", 0, "C", false, 0, "")
		for j, l := range lineas {
			pdf.SetXY(margen+wNum+2, y+1.3+float64(j)*altoLinea)
			pdf.CellFormat(wPreg-4, altoLinea, l, "", 0, "L", false, 0, "")
		}
		r, ok := respuestas[q.Codigo]
		resp := "Sin responder"
		if ok {
			resp = siNo(r.Respuesta)
		}
		pdf.SetFont("Helvetica", "B", 8.5)
		if !ok || (q.ImpidePostular && r.Respuesta != q.RespuestaHabilitante) {
			f.texto(colRojo) // resalta lo que impediría postular
		}
		pdf.SetXY(margen+wNum+wPreg, y)
		pdf.CellFormat(wResp, alto, f.tr(resp), "", 0, "C", false, 0, "")
		pdf.SetXY(margen, y+alto)

		if ok && r.Codigo == codigoPreguntaRemocion && r.Respuesta {
			f.fila([2]string{"Institución", r.RemocionInstitucion}, [2]string{"Cargo", r.RemocionCargo})
			f.fila([2]string{"Fecha de finalización", fechaTexto(r.RemocionFechaFin)}, [2]string{"Motivo de finalización", r.RemocionMotivo})
		}
	}
}

// consentimiento muestra el texto de 1.10.1 y la casilla de aceptación.
func (f *formularioPDF) consentimiento(aceptado bool) {
	pdf := f.pdf
	pdf.SetFont("Helvetica", "", 8.3)
	f.texto(colTexto)
	for _, t := range textoConsentimiento {
		f.espacio(float64(len(f.partir(t, anchoUtil-6)))*4.2 + 2)
		pdf.SetX(margen + 3)
		pdf.MultiCell(anchoUtil-6, 4.2, f.tr(t), "", "J", false)
		pdf.Ln(1.2)
	}
	pdf.Ln(1)
	f.espacio(8)
	y := pdf.GetY()
	f.trazo(colAzulOscuro)
	pdf.SetLineWidth(0.35)
	pdf.Rect(margen+3, y+0.8, 4, 4, "D")
	if aceptado {
		pdf.Line(margen+3.7, y+1.5, margen+6.3, y+4.1)
		pdf.Line(margen+6.3, y+1.5, margen+3.7, y+4.1)
	}
	pdf.SetLineWidth(0.2)
	pdf.SetXY(margen+9, y)
	pdf.SetFont("Helvetica", "B", 8.5)
	f.texto(colAzulOscuro)
	pdf.CellFormat(anchoUtil-9, 5.6, f.tr("He leído y acepto la autorización y declaración que anteceden."), "", 1, "L", false, 0, "")
}

// firma dibuja el bloque de firma (1.10.2) con el recuadro para la huella dactilar.
func (f *formularioPDF) firma(p *models.Postulacion) {
	pdf := f.pdf
	const wFirma, wHuella, alto = 128.0, 46.0, 44.0
	pdf.Ln(6)
	f.espacio(alto + 10)
	y := pdf.GetY()
	xHuella := anchoPagina - margen - wHuella

	f.trazo(colBorde)
	pdf.Rect(margen, y, wFirma, alto, "D")
	pdf.Rect(xHuella, y, wHuella, alto, "D")

	pdf.SetFont("Helvetica", "B", 6.3)
	f.texto(colEtiqueta)
	pdf.SetXY(margen+2, y+1.2)
	pdf.CellFormat(wFirma-4, 3.5, "FIRMA DEL POSTULANTE", "", 0, "L", false, 0, "")
	pdf.SetXY(xHuella+2, y+1.2)
	pdf.CellFormat(wHuella-4, 3.5, f.tr("HUELLA DACTILAR"), "", 0, "L", false, 0, "")

	f.trazo(colTexto)
	pdf.Line(margen+14, y+27, margen+wFirma-14, y+27)
	pdf.SetXY(margen, y+28.5)
	pdf.SetFont("Helvetica", "B", 9)
	f.texto(colTexto)
	pdf.CellFormat(wFirma, 4.6, f.tr(p.Nombres+" "+p.Apellidos), "", 2, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	pdf.CellFormat(wFirma, 4.4, f.tr("C.I. "+p.Cedula), "", 0, "C", false, 0, "")

	pdf.SetXY(margen, y+alto+4)
	pdf.CellFormat(anchoUtil, 5, "Lugar y fecha: ______________________________________________", "", 1, "L", false, 0, "")
}

// generarFormularioPDF arma el formulario de postulación completo (1.1 a 1.10) listo para imprimir y firmar.
func generarFormularioPDF(p *models.Postulacion, d datosFormulario) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margen, margen, margen)
	pdf.SetAutoPageBreak(true, margenInf)
	pdf.SetTitle("Formulario de postulación "+p.Codigo, true)
	pdf.SetAuthor("Instituto Ecuatoriano de Seguridad Social", true)
	pdf.AliasNbPages("")
	f := &formularioPDF{pdf: pdf, tr: pdf.UnicodeTranslatorFromDescriptor("")}

	enviada, generado := "Pendiente de envío", time.Now().Format("02/01/2006 15:04")
	if p.EnviadaEn != nil {
		enviada = p.EnviadaEn.Format("02/01/2006 15:04")
		generado = enviada
	}
	pdf.SetHeaderFunc(func() { f.encabezado(p.Codigo) })
	pdf.SetFooterFunc(func() { f.pie(p.Codigo, generado) })
	pdf.AddPage()

	// Título
	pdf.SetFont("Helvetica", "B", 15)
	f.texto(colAzulOscuro)
	pdf.CellFormat(anchoUtil, 8, f.tr("FORMULARIO DE POSTULACIÓN"), "", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "", 8.5)
	f.texto(colEtiqueta)
	pdf.MultiCell(anchoUtil, 4.2, f.tr("Concurso Público de Méritos y Oposición para la Designación de los Miembros Principales "+
		"y Alternos del Directorio del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), en representación "+
		"de los Afiliados, Jubilados y Empleadores"), "", "C", false)
	pdf.Ln(3)
	f.fila([2]string{"Número de formulario", p.Codigo}, [2]string{"Estado", p.Estado}, [2]string{"Fecha de envío", enviada})

	f.seccion("1.1", "Información personal")
	f.fila([2]string{"Cédula de identidad", p.Cedula}, [2]string{"Nacionalidad", p.Nacionalidad})
	f.fila([2]string{"Nombres", p.Nombres}, [2]string{"Apellidos", p.Apellidos})
	f.fila([2]string{"Género", etiqueta(p.Genero)}, [2]string{"Fecha de nacimiento", fechaTexto(p.FechaNacimiento)},
		[2]string{"Estado civil", etiqueta(p.EstadoCivil)})
	if p.ParejaNombres != "" {
		f.fila([2]string{"Nombres y apellidos de la pareja", p.ParejaNombres}, [2]string{"Cédula de la pareja", p.ParejaCedula})
	}

	f.seccion("1.2", "Información domiciliaria y de contacto")
	f.fila([2]string{"País", d.Pais}, [2]string{"Provincia / Estado", d.Provincia}, [2]string{"Ciudad", d.Ciudad})
	f.fila([2]string{"Calle principal", p.CallePrincipal}, [2]string{"Número", p.Numero}, [2]string{"Calle secundaria", p.CalleSecundaria})
	f.fila([2]string{"Sector", p.Sector}, [2]string{"Referencia", p.Referencia})
	f.fila([2]string{"Teléfono celular", p.TelefonoCelular}, [2]string{"Teléfono de domicilio", p.TelefonoDomicilio})
	f.fila([2]string{"Correo electrónico 1", p.Correo1}, [2]string{"Correo electrónico 2", p.Correo2})

	f.seccion("1.3", "Información adicional")
	f.fila([2]string{"¿Tiene familiares que trabajan en el IESS o BIESS?", siNo(p.TieneFamiliaresIESS)})
	for i, fam := range p.Familiares {
		f.subtitulo(fmt.Sprintf("Familiar %d", i+1))
		f.fila([2]string{"Nombres y apellidos", fam.NombresApellidos}, [2]string{"Parentesco", etiqueta(fam.Parentesco)})
		f.fila([2]string{"Institución", fam.Institucion}, [2]string{"Área", fam.Area}, [2]string{"Cargo", fam.Cargo})
	}

	f.seccion("1.4", "Cargo al que postula")
	verificacion := ""
	if p.VerificacionCargo != nil {
		verificacion = p.VerificacionCargo.Estado
	}
	f.fila([2]string{"Cargo", d.Cargo}, [2]string{"Verificación en el Historial Laboral del IESS", verificacion})
	if p.RespaldadoOrganizacion != nil {
		f.fila([2]string{"¿Cuenta con el respaldo de una organización?", siNo(*p.RespaldadoOrganizacion)})
	}
	if p.OrganizacionNombre != "" {
		f.fila([2]string{"Nombre de la organización", p.OrganizacionNombre},
			[2]string{"Documento de respaldo", nombreDocumento(p.OrganizacionDocumento)})
	}

	f.seccion("1.5", "Formación académica")
	if len(p.Formacion) == 0 {
		f.ninguno()
	}
	for i, x := range p.Formacion {
		f.subtitulo(fmt.Sprintf("Formación %d", i+1))
		f.fila([2]string{"Nivel de instrucción", etiqueta(x.NivelInstruccion)}, [2]string{"Tipo", etiqueta(x.Tipo)})
		f.fila([2]string{"Institución educativa", x.Institucion}, [2]string{"Título obtenido", x.Titulo})
		f.fila([2]string{"Área de estudios", etiqueta(x.AreaEstudios)}, [2]string{"Registro SENESCYT", x.RegistroSenescyt},
			[2]string{"Fecha de registro", fechaTexto(x.FechaRegistroSenescyt)})
		f.fila([2]string{"Documento de respaldo", nombreDocumento(x.DocumentoRespaldo)})
	}

	f.seccion("1.6", "Experiencia profesional")
	if len(p.Experiencia) == 0 {
		f.ninguno()
	}
	for i, x := range p.Experiencia {
		f.subtitulo(fmt.Sprintf("Experiencia %d", i+1))
		f.fila([2]string{"Tipo de experiencia", etiqueta(x.Tipo)}, [2]string{"Cargo o puesto", x.Cargo})
		f.fila([2]string{"Institución o empresa", x.Institucion}, [2]string{"Actividades relacionadas a", etiqueta(x.Actividades)})
		f.fila([2]string{"Fecha de inicio", fechaTexto(x.FechaInicio)}, [2]string{"Fecha de fin", fechaTexto(x.FechaFin)},
			[2]string{"Tiempo", duracionTexto(x.Duracion)})
		f.fila([2]string{"Descripción de funciones", x.DescripcionFunciones})
		f.fila([2]string{"Documento de respaldo", nombreDocumento(x.DocumentoRespaldo)})
	}
	if len(p.Experiencia) > 0 {
		pdf.Ln(1.5)
		f.fila([2]string{"Experiencia total (sin contar periodos superpuestos)", duracionTexto(p.ExperienciaTotal)})
	}

	f.seccion("1.7", "Publicaciones e investigaciones")
	if len(p.Publicaciones) == 0 {
		f.ninguno()
	}
	for i, x := range p.Publicaciones {
		f.subtitulo(fmt.Sprintf("Publicación %d", i+1))
		f.fila([2]string{"Título", x.Titulo})
		f.fila([2]string{"Tipo", etiqueta(x.Tipo)}, [2]string{"Medio de publicación", x.MedioPublicacion},
			[2]string{"Fecha de publicación", fechaTexto(x.FechaPublicacion)})
		f.fila([2]string{"¿Relacionada con su formación?", siNo(x.RelacionadaFormacion)},
			[2]string{"Documento de respaldo", nombreDocumento(x.DocumentoRespaldo)})
	}

	f.seccion("1.8", "Capacitación")
	if len(p.Capacitaciones) == 0 {
		f.ninguno()
	}
	for i, x := range p.Capacitaciones {
		f.subtitulo(fmt.Sprintf("Evento %d", i+1))
		f.fila([2]string{"Tipo de evento", etiqueta(x.TipoEvento)}, [2]string{"Nombre del evento", x.NombreEvento})
		f.fila([2]string{"Institución", x.Institucion}, [2]string{"Fecha de inicio", fechaTexto(x.FechaInicio)},
			[2]string{"Fecha de fin", fechaTexto(x.FechaFin)})
		f.fila([2]string{"Número de horas", fmt.Sprint(x.Horas)}, [2]string{"Documento de respaldo", nombreDocumento(x.DocumentoRespaldo)})
	}

	f.seccion("1.9", "Inhabilidades")
	f.tablaInhabilidades(p, d.Preguntas)

	f.seccion("1.10", "Autorización y declaración")
	f.consentimiento(p.EnviadaEn != nil)
	f.firma(p)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("generar PDF del formulario: %w", err)
	}
	return buf.Bytes(), nil
}
