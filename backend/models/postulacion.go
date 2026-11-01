package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Fecha es una fecha sin hora. En JSON viaja como "AAAA-MM-DD" (input date del navegador);
// al recibir también acepta "DD/MM/AAAA" (formato del documento funcional).
type Fecha struct {
	time.Time
}

const layoutFechaISO = "2006-01-02"

func (f *Fecha) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return errors.New("la fecha debe ser un texto")
	}
	s = strings.TrimSpace(s)
	for _, layout := range []string{layoutFechaISO, "02/01/2006"} {
		if t, err := time.Parse(layout, s); err == nil {
			f.Time = t
			return nil
		}
	}
	return errors.New("fecha inválida: use AAAA-MM-DD o DD/MM/AAAA")
}

func (f Fecha) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Format(layoutFechaISO))
}

// NuevaFecha envuelve un time.Time leído de la base.
func NuevaFecha(t time.Time) *Fecha { return &Fecha{Time: t} }

// ---------------------------------------------------------------------------------------------
// Secciones del formulario EF-12 que se repiten (tablas hijas de postulaciones)
// ---------------------------------------------------------------------------------------------

// Familiar (1.3.1): familiar que trabaja en el IESS o BIESS.
type Familiar struct {
	NombresApellidos string `json:"nombres_apellidos"`
	Parentesco       string `json:"parentesco"`  // CONVIVIENTE, CONYUGE, HIJO, ... (lista en 01_schema.sql)
	Institucion      string `json:"institucion"` // IESS | BIESS
	Area             string `json:"area"`
	Cargo            string `json:"cargo"`
}

// Formacion (1.5): formación académica.
type Formacion struct {
	NivelInstruccion      string `json:"nivel_instruccion"` // TERCER_NIVEL, CUARTO_ESPECIALIDAD, CUARTO_MAESTRIA, CUARTO_PHD
	Institucion           string `json:"institucion"`
	Tipo                  string `json:"tipo"` // NACIONAL | EXTRANJERA
	Titulo                string `json:"titulo"`
	AreaEstudios          string `json:"area_estudios"` // FINANZAS, MERCADO_CAPITALES, ADMINISTRACION, ECONOMIA, DERECHO, AFINES
	RegistroSenescyt      string `json:"registro_senescyt"`
	FechaRegistroSenescyt *Fecha `json:"fecha_registro_senescyt"`
	DocumentoRespaldo     string `json:"documento_respaldo"`
}

// Experiencia (1.6): experiencia profesional.
type Experiencia struct {
	Tipo                 string `json:"tipo"` // GENERAL | DIRECCION
	Cargo                string `json:"cargo"`
	Institucion          string `json:"institucion"`
	FechaInicio          *Fecha `json:"fecha_inicio"`
	FechaFin             *Fecha `json:"fecha_fin"`
	Actividades          string `json:"actividades"` // mismas opciones que area_estudios
	DescripcionFunciones string `json:"descripcion_funciones"`
	DocumentoRespaldo    string `json:"documento_respaldo"`
	// Duracion la calcula el servidor (1.6: años, meses y días). Si el frontend la envía, se ignora.
	Duracion *Duracion `json:"duracion,omitempty"`
}

// Publicacion (1.7): publicaciones e investigaciones.
type Publicacion struct {
	Titulo               string `json:"titulo"`
	Tipo                 string `json:"tipo"` // PUBLICACION | TRABAJO_INVESTIGACION
	MedioPublicacion     string `json:"medio_publicacion"`
	FechaPublicacion     *Fecha `json:"fecha_publicacion"`
	RelacionadaFormacion bool   `json:"relacionada_formacion"`
	DocumentoRespaldo    string `json:"documento_respaldo"`
}

// Capacitacion (1.8).
type Capacitacion struct {
	TipoEvento        string `json:"tipo_evento"` // CURSO | SEMINARIO | TALLER
	NombreEvento      string `json:"nombre_evento"`
	Institucion       string `json:"institucion"`
	FechaInicio       *Fecha `json:"fecha_inicio"`
	FechaFin          *Fecha `json:"fecha_fin"`
	Horas             int    `json:"horas"`
	DocumentoRespaldo string `json:"documento_respaldo"`
}

// RespuestaInhabilidad (1.9): respuesta Sí/No a una pregunta, identificada por su código ("1.9.1" ... "1.9.13").
// Los campos de remoción solo aplican a la pregunta 1.9.7.
type RespuestaInhabilidad struct {
	Codigo              string `json:"codigo"`
	Respuesta           bool   `json:"respuesta"`
	RemocionInstitucion string `json:"remocion_institucion,omitempty"`
	RemocionCargo       string `json:"remocion_cargo,omitempty"`
	RemocionFechaFin    *Fecha `json:"remocion_fecha_fin,omitempty"`
	RemocionMotivo      string `json:"remocion_motivo,omitempty"`
}

// ---------------------------------------------------------------------------------------------
// Petición y respuesta
// ---------------------------------------------------------------------------------------------

// PostulacionReq es lo que envía el frontend al guardar (PUT /api/postulacion).
// No incluye la información personal (1.1) ni el correo 1: el servidor los toma del Registro Civil
// y de la cuenta, para que el postulante no pueda alterarlos.
type PostulacionReq struct {
	// 1.1.7.1 Datos de la pareja (solo si es CASADO o UNION_DE_HECHO)
	ParejaNombres string `json:"pareja_nombres"`
	ParejaCedula  string `json:"pareja_cedula"`

	// 1.2 Información domiciliaria y de contacto
	PaisID            *int   `json:"pais_id"`
	ProvinciaID       *int   `json:"provincia_id"`
	CantonID          *int   `json:"canton_id"`
	ProvinciaTexto    string `json:"provincia_texto"` // si el país NO es Ecuador
	CiudadTexto       string `json:"ciudad_texto"`    // si el país NO es Ecuador
	TelefonoCelular   string `json:"telefono_celular"`
	Correo2           string `json:"correo_2"`
	CallePrincipal    string `json:"calle_principal"`
	Numero            string `json:"numero"`
	CalleSecundaria   string `json:"calle_secundaria"`
	Sector            string `json:"sector"`
	Referencia        string `json:"referencia"`
	TelefonoDomicilio string `json:"telefono_domicilio"`

	// 1.3 Información adicional
	TieneFamiliaresIESS bool       `json:"tiene_familiares_iess"`
	Familiares          []Familiar `json:"familiares"`

	// 1.4 Cargo al que postula (AFILIADOS | JUBILADOS | EMPLEADORES) y respaldo de organización
	CargoCodigo            string `json:"cargo_codigo"`
	RespaldadoOrganizacion *bool  `json:"respaldado_organizacion"`
	OrganizacionNombre     string `json:"organizacion_nombre"`
	OrganizacionDocumento  string `json:"organizacion_documento"`

	// 1.5 a 1.9
	Formacion      []Formacion            `json:"formacion"`
	Experiencia    []Experiencia          `json:"experiencia"`
	Publicaciones  []Publicacion          `json:"publicaciones"`
	Capacitaciones []Capacitacion         `json:"capacitaciones"`
	Inhabilidades  []RespuestaInhabilidad `json:"inhabilidades"`
}

// Postulacion es el formulario completo: cabecera + secciones. Es lo que se guarda y lo que se devuelve.
type Postulacion struct {
	ID            int64      `json:"id"`
	Codigo        string     `json:"codigo"` // código de numeración único, p. ej. BIESS-2026-000001
	Estado        string     `json:"estado"` // NUEVA (aún no guardada) | BORRADOR | ENVIADA
	PostulanteID  int64      `json:"-"`
	EnviadaEn     *time.Time `json:"enviada_en"`
	ActualizadoEn *time.Time `json:"actualizado_en"`

	// 1.1 Información personal (Registro Civil, solo lectura)
	Cedula          string `json:"cedula"`
	Nombres         string `json:"nombres"`
	Apellidos       string `json:"apellidos"`
	Genero          string `json:"genero"`
	FechaNacimiento *Fecha `json:"fecha_nacimiento"`
	Nacionalidad    string `json:"nacionalidad"`
	EstadoCivil     string `json:"estado_civil"`

	// 1.2.5 Correo 1 = correo de la cuenta (solo lectura)
	Correo1 string `json:"correo_1"`

	PostulacionReq

	// Calculados por el servidor (Fase 5, reglas de negocio)
	ExperienciaTotal  *Duracion          `json:"experiencia_total"`  // sin contar dos veces los periodos que se cruzan
	VerificacionCargo *VerificacionCargo `json:"verificacion_cargo"` // resultado del Historial Laboral para el cargo elegido
	Bloqueo           *Bloqueo           `json:"bloqueo"`            // si no es null, el postulante NO puede continuar
}

// Duracion expresa un tiempo de experiencia en años, meses y días.
type Duracion struct {
	Anios int `json:"anios"`
	Meses int `json:"meses"`
	Dias  int `json:"dias"`
}

// HistorialLaboral es la fuente SIMULADA del Historial Laboral del IESS (tabla historial_laboral_simulado).
type HistorialLaboral struct {
	EsAfiliado  bool
	EsJubilado  bool
	EsEmpleador bool
}

// VerificacionCargo es el resultado de 1.4.x: si el postulante cumple la calidad que exige el cargo.
type VerificacionCargo struct {
	CargoCodigo string `json:"cargo_codigo"`
	Cumple      bool   `json:"cumple"`
	Estado      string `json:"estado"`   // p. ej. "Postulante afiliado al IESS" / "Postulante No Afiliado al IESS"
	Mensaje     string `json:"mensaje"`  // mensaje de rechazo del documento funcional (vacío si cumple)
	Pregunta    string `json:"pregunta"` // pregunta del respaldo de organización que corresponde al cargo
}

// Bloqueo indica por qué la postulación quedó cerrada (p. ej. nacionalidad no ecuatoriana).
type Bloqueo struct {
	Codigo  string `json:"codigo"`
	Mensaje string `json:"mensaje"`
}

// EnviarReq es el cuerpo de POST /api/postulacion/enviar (1.10).
type EnviarReq struct {
	ConsentimientoAceptado bool `json:"consentimiento_aceptado"`
}

// Cargo es una opción de 1.4.
type Cargo struct {
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Sector string `json:"sector"`
}

// PreguntaInhabilidad es una pregunta de 1.9. RespuestaHabilitante es la respuesta con la que SÍ puede postular.
type PreguntaInhabilidad struct {
	Codigo               string `json:"codigo"`
	Texto                string `json:"texto"`
	RespuestaHabilitante bool   `json:"respuesta_habilitante"`
}

// OpcionesPostulacion agrupa los catálogos propios del formulario (cargos y preguntas de inhabilidades).
type OpcionesPostulacion struct {
	Cargos        []Cargo               `json:"cargos"`
	Inhabilidades []PreguntaInhabilidad `json:"inhabilidades"`
}
