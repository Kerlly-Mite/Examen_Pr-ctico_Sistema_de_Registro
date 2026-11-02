package services

// Reglas de negocio del formulario de postulación (EF-12, requerimiento IESS v1.4).
// Son funciones PURAS: no consultan la base ni Redis, reciben los datos y devuelven el resultado.
// Por eso se prueban con tests unitarios sin levantar Docker (reglas_postulacion_test.go).

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"biess-backend/models"
)

const (
	codigoPreguntaRemocion  = "1.9.7" // única pregunta con datos adicionales si la respuesta es "Sí"
	maxDescripcionFunciones = 1000    // 1.6.7: máximo 1000 caracteres
)

var reTelefono = regexp.MustCompile(`^\+?[0-9]{7,15}$`)

// ---------------------------------------------------------------------------------------------
// 1.1.6 Nacionalidad
// ---------------------------------------------------------------------------------------------

const msgNacionalidad = "Su postulación no puede continuar: solo pueden postular personas de nacionalidad ecuatoriana."

// esEcuatoriana: si la nacionalidad no es ecuatoriana, el sistema cierra la postulación.
func esEcuatoriana(nacionalidad string) bool {
	return strings.EqualFold(strings.TrimSpace(nacionalidad), "ECUATORIANA")
}

func errNacionalidad() *models.AppError {
	return models.NuevoError(http.StatusForbidden, "nacionalidad_no_permitida", msgNacionalidad)
}

// ---------------------------------------------------------------------------------------------
// 1.4 Cargo al que postula vs. Historial Laboral del IESS
// ---------------------------------------------------------------------------------------------

type reglaCargo struct {
	cumple   func(models.HistorialLaboral) bool
	estadoSi string
	estadoNo string
	rechazo  string
	pregunta string
}

// reglasCargo contiene los textos EXACTOS del documento funcional (1.4.1, 1.4.2 y 1.4.3).
var reglasCargo = map[string]reglaCargo{
	"AFILIADOS": {
		cumple:   func(h models.HistorialLaboral) bool { return h.EsAfiliado },
		estadoSi: "Postulante afiliado al IESS",
		estadoNo: "Postulante No Afiliado al IESS",
		rechazo: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados " +
			"ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS",
		pregunta: "¿Se encuentra respaldado por una organización de trabajadores de representación nacional " +
			"legalmente constituida ante autoridad competente?",
	},
	"JUBILADOS": {
		cumple:   func(h models.HistorialLaboral) bool { return h.EsJubilado },
		estadoSi: "Postulante acredita jubilación",
		estadoNo: "Postulante no acredita jubilación",
		rechazo: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados " +
			"ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación",
		pregunta: "¿Se encuentra respaldado por una organización de jubilados de representación nacional " +
			"legalmente constituida ante autoridad competente?",
	},
	"EMPLEADORES": {
		cumple:   func(h models.HistorialLaboral) bool { return h.EsEmpleador },
		estadoSi: "Postulante es empleador",
		estadoNo: "Postulante no es empleador",
		rechazo: "Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador " +
			"ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador",
		pregunta: "¿Se encuentra respaldado por una federación de cámara de representación nacional " +
			"legalmente constituida ante autoridad competente?",
	},
}

// verificarCargo aplica 1.4.x. Devuelve false en el segundo valor si el código de cargo no existe.
func verificarCargo(codigo string, h models.HistorialLaboral) (*models.VerificacionCargo, bool) {
	regla, ok := reglasCargo[codigo]
	if !ok {
		return nil, false
	}
	v := &models.VerificacionCargo{CargoCodigo: codigo, Pregunta: regla.pregunta}
	if regla.cumple(h) {
		v.Cumple, v.Estado = true, regla.estadoSi
	} else {
		v.Estado, v.Mensaje = regla.estadoNo, regla.rechazo
	}
	return v, true
}

func errCargo(v *models.VerificacionCargo) *models.AppError {
	return &models.AppError{
		Status:  http.StatusForbidden,
		Code:    "cargo_no_habilitado",
		Message: v.Mensaje,
		Extra:   map[string]any{"estado": v.Estado},
	}
}

// ---------------------------------------------------------------------------------------------
// Validaciones al GUARDAR (Guardar progreso). Se permite el formulario incompleto, pero lo que se
// ingresa debe ser coherente. Además limpia los datos que no corresponden (evita datos huérfanos).
// ---------------------------------------------------------------------------------------------

func errDato(seccion, msg string) *models.AppError {
	return &models.AppError{
		Status:  http.StatusBadRequest,
		Code:    "solicitud_invalida",
		Message: seccion + ": " + msg,
	}
}

type campo struct{ nombre, valor string }

// exigir verifica que los campos obligatorios (*) de un registro de una lista no estén vacíos.
func exigir(seccion string, i int, campos ...campo) *models.AppError {
	for _, c := range campos {
		if strings.TrimSpace(c.valor) == "" {
			return errDato(seccion, fmt.Sprintf("en el registro %d complete el campo \"%s\"", i+1, c.nombre))
		}
	}
	return nil
}

// esContenidoPDF revisa la "firma" del archivo: todo PDF empieza con "%PDF-" y termina con "%%EOF".
// Así no se acepta, por ejemplo, un .exe o un .docx renombrado a .pdf.
func esContenidoPDF(datos []byte) bool {
	final := datos
	if len(final) > 2048 {
		final = final[len(final)-2048:]
	}
	return bytes.HasPrefix(datos, []byte("%PDF-")) && bytes.Contains(final, []byte("%%EOF"))
}

// esPDF: los documentos de respaldo se aceptan solo en formato PDF.
func esPDF(nombre string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimSpace(nombre)), ".pdf")
}

func exigirPDF(seccion string, i int, nombre string) *models.AppError {
	if !esPDF(nombre) {
		return errDato(seccion, fmt.Sprintf("en el registro %d el documento de respaldo debe ser un archivo PDF", i+1))
	}
	return nil
}

func exigirFecha(seccion string, i int, nombre string, f *models.Fecha, hoy time.Time) *models.AppError {
	if f == nil {
		return errDato(seccion, fmt.Sprintf("en el registro %d complete el campo \"%s\"", i+1, nombre))
	}
	if f.After(hoy) {
		return errDato(seccion, fmt.Sprintf("en el registro %d la \"%s\" no puede ser futura", i+1, nombre))
	}
	return nil
}

func exigirRango(seccion string, i int, inicio, fin *models.Fecha, hoy time.Time) *models.AppError {
	if err := exigirFecha(seccion, i, "fecha de inicio", inicio, hoy); err != nil {
		return err
	}
	if err := exigirFecha(seccion, i, "fecha de finalización", fin, hoy); err != nil {
		return err
	}
	if fin.Before(inicio.Time) {
		return errDato(seccion, fmt.Sprintf("en el registro %d la fecha de finalización es anterior a la de inicio", i+1))
	}
	return nil
}

// validarGuardado aplica las reglas que deben cumplirse en CADA guardado. Modifica p para descartar
// los datos que no corresponden. ecuadorID es el id del país Ecuador en el catálogo; hoy, la fecha actual.
func validarGuardado(p *models.Postulacion, ecuadorID int, hoy time.Time) *models.AppError {
	if err := validarPareja(p); err != nil {
		return err
	}
	if err := validarDomicilio(p, ecuadorID); err != nil {
		return err
	}

	// 1.3 Familiares: solo si respondió "Sí".
	if !p.TieneFamiliaresIESS {
		p.Familiares = nil
	}
	for i, f := range p.Familiares {
		if err := exigir("Familiares en el IESS o BIESS", i,
			campo{"Nombres y apellidos", f.NombresApellidos}, campo{"Parentesco", f.Parentesco},
			campo{"Institución", f.Institucion}, campo{"Área", f.Area}, campo{"Cargo", f.Cargo}); err != nil {
			return err
		}
	}

	// 1.4 Respaldo de organización: nombre y documento obligatorios (*) solo si respondió "Sí".
	if p.RespaldadoOrganizacion != nil && *p.RespaldadoOrganizacion {
		if strings.TrimSpace(p.OrganizacionNombre) == "" {
			return errDato("Cargo al que postula", "ingrese el nombre de la organización que lo respalda")
		}
		if !esPDF(p.OrganizacionDocumento) {
			return errDato("Cargo al que postula", "adjunte el documento de respaldo de la organización en PDF")
		}
	} else {
		p.OrganizacionNombre, p.OrganizacionDocumento = "", ""
	}

	for i, f := range p.Formacion {
		const s = "Formación académica"
		if err := exigir(s, i, campo{"Nivel de instrucción", f.NivelInstruccion}, campo{"Institución educativa", f.Institucion},
			campo{"Tipo", f.Tipo}, campo{"Título obtenido", f.Titulo}, campo{"Área de estudios", f.AreaEstudios},
			campo{"Número de registro SENESCYT", f.RegistroSenescyt}); err != nil {
			return err
		}
		if err := exigirFecha(s, i, "fecha de registro SENESCYT", f.FechaRegistroSenescyt, hoy); err != nil {
			return err
		}
		if err := exigirPDF(s, i, f.DocumentoRespaldo); err != nil {
			return err
		}
	}

	for i, e := range p.Experiencia {
		const s = "Experiencia profesional"
		if err := exigir(s, i, campo{"Tipo de experiencia", e.Tipo}, campo{"Cargo o puesto", e.Cargo},
			campo{"Institución o empresa", e.Institucion}, campo{"Actividades relacionadas a", e.Actividades},
			campo{"Descripción de funciones", e.DescripcionFunciones}); err != nil {
			return err
		}
		if utf8.RuneCountInString(e.DescripcionFunciones) > maxDescripcionFunciones {
			return errDato(s, fmt.Sprintf("en el registro %d la descripción de funciones supera los %d caracteres", i+1, maxDescripcionFunciones))
		}
		if err := exigirRango(s, i, e.FechaInicio, e.FechaFin, hoy); err != nil {
			return err
		}
		if err := exigirPDF(s, i, e.DocumentoRespaldo); err != nil {
			return err
		}
	}

	for i, pub := range p.Publicaciones {
		const s = "Publicaciones"
		if err := exigir(s, i, campo{"Título", pub.Titulo}, campo{"Tipo", pub.Tipo},
			campo{"Medio de publicación", pub.MedioPublicacion}); err != nil {
			return err
		}
		if err := exigirFecha(s, i, "fecha de publicación", pub.FechaPublicacion, hoy); err != nil {
			return err
		}
		if err := exigirPDF(s, i, pub.DocumentoRespaldo); err != nil {
			return err
		}
	}

	for i, c := range p.Capacitaciones {
		const s = "Capacitación"
		if err := exigir(s, i, campo{"Tipo de evento", c.TipoEvento}, campo{"Nombre del evento", c.NombreEvento},
			campo{"Institución", c.Institucion}); err != nil {
			return err
		}
		if err := exigirRango(s, i, c.FechaInicio, c.FechaFin, hoy); err != nil {
			return err
		}
		if c.Horas <= 0 {
			return errDato(s, fmt.Sprintf("en el registro %d el número de horas debe ser mayor a cero", i+1))
		}
		if err := exigirPDF(s, i, c.DocumentoRespaldo); err != nil {
			return err
		}
	}

	return validarInhabilidades(p, hoy)
}

// validarPareja (1.1.7.1): si es CASADO o UNION_DE_HECHO, los datos de la pareja son obligatorios.
// Si no lo es, se descartan (no deben quedar datos de una pareja que no existe).
func validarPareja(p *models.Postulacion) *models.AppError {
	if p.EstadoCivil != "CASADO" && p.EstadoCivil != "UNION_DE_HECHO" {
		p.ParejaNombres, p.ParejaCedula = "", ""
		return nil
	}
	const s = "Información personal"
	if strings.TrimSpace(p.ParejaNombres) == "" || strings.TrimSpace(p.ParejaCedula) == "" {
		return errDato(s, "por su estado civil, ingrese los nombres, apellidos y la cédula de su pareja")
	}
	p.ParejaCedula = strings.TrimSpace(p.ParejaCedula)
	if !CedulaValida(p.ParejaCedula) {
		return errDato(s, "la cédula de la pareja no es válida")
	}
	if p.ParejaCedula == p.Cedula {
		return errDato(s, "la cédula de la pareja no puede ser la suya")
	}
	return nil
}

// validarDomicilio (1.2.1-1.2.3): si el país es Ecuador, provincia y cantón se eligen de los catálogos
// (cascada); si es otro país, se escriben a mano. La pertenencia cantón -> provincia -> país la garantiza
// además la base con claves foráneas compuestas.
func validarDomicilio(p *models.Postulacion, ecuadorID int) *models.AppError {
	const s = "Información domiciliaria"
	switch {
	case p.PaisID == nil:
		if p.ProvinciaID != nil || p.CantonID != nil {
			return errDato(s, "seleccione primero el país")
		}
	case *p.PaisID == ecuadorID:
		p.ProvinciaTexto, p.CiudadTexto = "", ""
		if p.CantonID != nil && p.ProvinciaID == nil {
			return errDato(s, "seleccione la provincia antes que el cantón")
		}
	default:
		p.ProvinciaID, p.CantonID = nil, nil
	}

	if t := strings.TrimSpace(p.TelefonoCelular); t != "" && !reTelefono.MatchString(t) {
		return errDato(s, "el teléfono celular debe tener solo números (7 a 15 dígitos)")
	}
	if t := strings.TrimSpace(p.TelefonoDomicilio); t != "" && !reTelefono.MatchString(t) {
		return errDato(s, "el teléfono de domicilio debe tener solo números (7 a 15 dígitos)")
	}
	if p.Correo2 != "" && !correoValido(p.Correo2) {
		return errDato(s, "el correo electrónico 2 no es válido")
	}
	return nil
}

// validarInhabilidades (1.9): no se repiten preguntas y, si respondió "Sí" a 1.9.7, son obligatorios
// la institución y el motivo de finalización (*).
func validarInhabilidades(p *models.Postulacion, hoy time.Time) *models.AppError {
	const s = "Inhabilidades"
	vistas := map[string]bool{}
	for i := range p.Inhabilidades {
		r := &p.Inhabilidades[i]
		r.Codigo = strings.TrimSpace(r.Codigo)
		if vistas[r.Codigo] {
			return errDato(s, "la pregunta "+r.Codigo+" está respondida más de una vez")
		}
		vistas[r.Codigo] = true

		if r.Codigo != codigoPreguntaRemocion || !r.Respuesta {
			r.RemocionInstitucion, r.RemocionCargo, r.RemocionMotivo, r.RemocionFechaFin = "", "", "", nil
			continue
		}
		if strings.TrimSpace(r.RemocionInstitucion) == "" || strings.TrimSpace(r.RemocionMotivo) == "" {
			return errDato(s, "en la pregunta 1.9.7 ingrese la institución y el motivo de finalización")
		}
		if r.RemocionFechaFin != nil && r.RemocionFechaFin.After(hoy) {
			return errDato(s, "en la pregunta 1.9.7 la fecha de finalización no puede ser futura")
		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------
// Validaciones al ENVIAR (1.10): el formulario debe estar completo y el postulante, habilitado.
// ---------------------------------------------------------------------------------------------

func errIncompleta(seccion, msg string) *models.AppError {
	return &models.AppError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "postulacion_incompleta",
		Message: "Antes de enviar, complete la sección \"" + seccion + "\": " + msg,
		Extra:   map[string]any{"seccion": seccion},
	}
}

func validarEnvio(p *models.Postulacion, ecuadorID int, preguntas []models.PreguntaInhabilidad) *models.AppError {
	const dom = "Información domiciliaria y de contacto"
	switch {
	case p.PaisID == nil:
		return errIncompleta(dom, "seleccione el país")
	case *p.PaisID == ecuadorID && (p.ProvinciaID == nil || p.CantonID == nil):
		return errIncompleta(dom, "seleccione la provincia y el cantón")
	case *p.PaisID != ecuadorID && (strings.TrimSpace(p.ProvinciaTexto) == "" || strings.TrimSpace(p.CiudadTexto) == ""):
		return errIncompleta(dom, "ingrese la provincia o estado y la ciudad")
	case strings.TrimSpace(p.TelefonoCelular) == "":
		return errIncompleta(dom, "ingrese su teléfono celular")
	case strings.TrimSpace(p.CallePrincipal) == "":
		return errIncompleta(dom, "ingrese la calle principal de su domicilio")
	}

	if p.TieneFamiliaresIESS && len(p.Familiares) == 0 {
		return errIncompleta("Información adicional", "agregue al menos un familiar o cambie la respuesta a \"No\"")
	}
	if p.CargoCodigo == "" {
		return errIncompleta("Cargo al que postula", "seleccione el cargo")
	}
	if p.RespaldadoOrganizacion == nil {
		return errIncompleta("Cargo al que postula", "responda si está respaldado por una organización")
	}

	// 1.9: todas las preguntas respondidas y con la respuesta habilitante.
	respuestas := map[string]bool{}
	for _, r := range p.Inhabilidades {
		respuestas[r.Codigo] = r.Respuesta
	}
	var faltan, inhabilitan []string
	for _, q := range preguntas {
		r, ok := respuestas[q.Codigo]
		switch {
		case !ok:
			faltan = append(faltan, q.Codigo)
		case r != q.RespuestaHabilitante:
			inhabilitan = append(inhabilitan, q.Codigo)
		}
	}
	if len(faltan) > 0 {
		e := errIncompleta("Inhabilidades", "responda las preguntas "+strings.Join(faltan, ", "))
		e.Extra["preguntas"] = faltan
		return e
	}
	if len(inhabilitan) > 0 {
		return &models.AppError{
			Status: http.StatusForbidden,
			Code:   "postulante_inhabilitado",
			Message: "Su postulación no puede enviarse: según sus respuestas se encuentra inmerso en una inhabilidad " +
				"(preguntas " + strings.Join(inhabilitan, ", ") + ").",
			Extra: map[string]any{"preguntas": inhabilitan},
		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------
// 1.6 Cálculo automático de la experiencia (años, meses y días)
// ---------------------------------------------------------------------------------------------

// calcularDuracion cuenta el tiempo entre dos fechas INCLUSIVE (del 01/01/2020 al 31/12/2020 = 1 año).
func calcularDuracion(inicio, fin time.Time) models.Duracion {
	hasta := fin.AddDate(0, 0, 1) // el último día también se trabajó
	anios := hasta.Year() - inicio.Year()
	meses := int(hasta.Month()) - int(inicio.Month())
	dias := hasta.Day() - inicio.Day()
	if dias < 0 {
		meses--
		// días del mes anterior a "hasta" (el día 0 de un mes es el último día del mes previo)
		dias += time.Date(hasta.Year(), hasta.Month(), 0, 0, 0, 0, 0, time.UTC).Day()
	}
	if meses < 0 {
		anios--
		meses += 12
	}
	return models.Duracion{Anios: anios, Meses: meses, Dias: dias}
}

type periodo struct{ inicio, fin time.Time }

// experienciaTotal suma la experiencia SIN contar dos veces los periodos que se cruzan
// (dos trabajos al mismo tiempo cuentan una sola vez). Para sumar, 30 días = 1 mes y 12 meses = 1 año.
func experienciaTotal(exps []models.Experiencia) models.Duracion {
	var periodos []periodo
	for _, e := range exps {
		if e.FechaInicio != nil && e.FechaFin != nil && !e.FechaFin.Before(e.FechaInicio.Time) {
			periodos = append(periodos, periodo{e.FechaInicio.Time, e.FechaFin.Time})
		}
	}
	sort.Slice(periodos, func(i, j int) bool { return periodos[i].inicio.Before(periodos[j].inicio) })

	// Une los periodos que se cruzan o son consecutivos.
	var unidos []periodo
	for _, p := range periodos {
		if n := len(unidos); n > 0 && !p.inicio.After(unidos[n-1].fin.AddDate(0, 0, 1)) {
			if p.fin.After(unidos[n-1].fin) {
				unidos[n-1].fin = p.fin
			}
			continue
		}
		unidos = append(unidos, p)
	}

	var total models.Duracion
	for _, p := range unidos {
		d := calcularDuracion(p.inicio, p.fin)
		total.Anios += d.Anios
		total.Meses += d.Meses
		total.Dias += d.Dias
	}
	total.Meses += total.Dias / 30
	total.Dias %= 30
	total.Anios += total.Meses / 12
	total.Meses %= 12
	return total
}
