package services

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"biess-backend/models"
	"biess-backend/repositories"
)

// PostulacionService implementa EF-12: consultar y guardar el formulario de postulación.
type PostulacionService interface {
	// Obtener devuelve la postulación guardada o, si aún no existe, una nueva (estado NUEVA)
	// con la información personal precargada desde el Registro Civil.
	Obtener(ctx context.Context, postulanteID int64) (*models.Postulacion, error)
	// Guardar persiste el formulario completo de forma atómica y devuelve lo que quedó en la base.
	Guardar(ctx context.Context, postulanteID int64, req models.PostulacionReq) (*models.Postulacion, error)
	// Opciones devuelve los cargos y las preguntas de inhabilidades para armar el formulario.
	Opciones(ctx context.Context) (*models.OpcionesPostulacion, error)
}

type postulacionService struct {
	postulaciones repositories.PostulacionRepository
	postulantes   repositories.PostulanteRepository
	registro      repositories.RegistroCivilRepository
}

// NewPostulacionService recibe sus dependencias como interfaces (Inyección de Dependencias).
func NewPostulacionService(
	postulaciones repositories.PostulacionRepository,
	postulantes repositories.PostulanteRepository,
	registro repositories.RegistroCivilRepository,
) PostulacionService {
	return &postulacionService{postulaciones: postulaciones, postulantes: postulantes, registro: registro}
}

// seccionesFormulario traduce la tabla donde falló la base a la sección del formulario que el usuario debe revisar.
var seccionesFormulario = map[string]string{
	"postulaciones":             "Información personal, domiciliaria y cargo",
	"familiares_iess":           "Familiares en el IESS o BIESS (1.3)",
	"formacion_academica":       "Formación académica (1.5)",
	"experiencia_profesional":   "Experiencia profesional (1.6)",
	"publicaciones":             "Publicaciones (1.7)",
	"capacitaciones":            "Capacitación (1.8)",
	"postulacion_inhabilidades": "Inhabilidades (1.9)",
}

func (s *postulacionService) Obtener(ctx context.Context, postulanteID int64) (*models.Postulacion, error) {
	perfil, err := s.perfil(ctx, postulanteID)
	if err != nil {
		return nil, err
	}

	p, err := s.postulaciones.BuscarPorPostulante(ctx, postulanteID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, repositories.ErrNoEncontrado) {
		return nil, models.ErrInterno(err)
	}

	// Primera vez: formulario vacío con los datos del Registro Civil y el correo de la cuenta.
	nueva, err := s.armarPostulacion(ctx, postulanteID, perfil, models.PostulacionReq{})
	if err != nil {
		return nil, err
	}
	nueva.Estado = "NUEVA"
	return nueva, nil
}

func (s *postulacionService) Guardar(ctx context.Context, postulanteID int64, req models.PostulacionReq) (*models.Postulacion, error) {
	perfil, err := s.perfil(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	p, err := s.armarPostulacion(ctx, postulanteID, perfil, normalizar(req))
	if err != nil {
		return nil, err
	}

	if err := s.postulaciones.Guardar(ctx, p); err != nil {
		return nil, traducirErrorPostulacion(err)
	}
	// Se devuelve lo que realmente quedó guardado (código único, estado, fechas).
	return s.Obtener(ctx, postulanteID)
}

func (s *postulacionService) Opciones(ctx context.Context) (*models.OpcionesPostulacion, error) {
	cargos, err := s.postulaciones.ListarCargos(ctx)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	preguntas, err := s.postulaciones.ListarPreguntasInhabilidad(ctx)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return &models.OpcionesPostulacion{Cargos: cargos, Inhabilidades: preguntas}, nil
}

// perfil obtiene la cuenta de la sesión. Si ya no existe (cuenta borrada), la sesión no es válida.
func (s *postulacionService) perfil(ctx context.Context, postulanteID int64) (*models.PerfilPostulante, error) {
	perfil, err := s.postulantes.BuscarPerfilPorID(ctx, postulanteID)
	if errors.Is(err, repositories.ErrNoEncontrado) {
		return nil, models.ErrNoAutorizado("Su sesión no es válida. Inicie sesión nuevamente.")
	}
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return perfil, nil
}

// armarPostulacion combina lo que envía el usuario con la información personal (1.1) del Registro Civil
// y el correo de la cuenta (1.2.5). Estos datos NUNCA se toman del cuerpo de la petición.
func (s *postulacionService) armarPostulacion(ctx context.Context, postulanteID int64, perfil *models.PerfilPostulante,
	req models.PostulacionReq) (*models.Postulacion, error) {
	rc, err := s.registro.BuscarPorCedula(ctx, perfil.Cedula)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return &models.Postulacion{
		PostulanteID:    postulanteID,
		Cedula:          rc.Cedula,
		Nombres:         rc.Nombres,
		Apellidos:       rc.Apellidos,
		Genero:          rc.Genero,
		FechaNacimiento: models.NuevaFecha(rc.FechaNacimiento),
		Nacionalidad:    rc.Nacionalidad,
		EstadoCivil:     rc.EstadoCivil,
		Correo1:         perfil.Correo,
		PostulacionReq:  conListasVacias(req),
	}, nil
}

// normalizar limpia la petición para no guardar datos huérfanos: si una respuesta es "No",
// se descartan los datos que dependen de ella.
func normalizar(req models.PostulacionReq) models.PostulacionReq {
	req.CargoCodigo = strings.ToUpper(strings.TrimSpace(req.CargoCodigo))
	req.Correo2 = strings.ToLower(strings.TrimSpace(req.Correo2))
	if !req.TieneFamiliaresIESS {
		req.Familiares = nil
	}
	if req.RespaldadoOrganizacion == nil || !*req.RespaldadoOrganizacion {
		req.OrganizacionNombre, req.OrganizacionDocumento = "", ""
	}
	return req
}

// conListasVacias garantiza [] (y no null) en el JSON de respuesta.
func conListasVacias(req models.PostulacionReq) models.PostulacionReq {
	if req.Familiares == nil {
		req.Familiares = []models.Familiar{}
	}
	if req.Formacion == nil {
		req.Formacion = []models.Formacion{}
	}
	if req.Experiencia == nil {
		req.Experiencia = []models.Experiencia{}
	}
	if req.Publicaciones == nil {
		req.Publicaciones = []models.Publicacion{}
	}
	if req.Capacitaciones == nil {
		req.Capacitaciones = []models.Capacitacion{}
	}
	if req.Inhabilidades == nil {
		req.Inhabilidades = []models.RespuestaInhabilidad{}
	}
	return req
}

// traducirErrorPostulacion convierte los errores del repositorio en errores de negocio para el usuario.
func traducirErrorPostulacion(err error) error {
	var dato *repositories.DatoInvalidoError
	var codigo *repositories.CodigoInexistenteError
	switch {
	case errors.Is(err, repositories.ErrPostulacionEnviada):
		return models.ErrConflicto("postulacion_enviada", "Su postulación ya fue enviada y no se puede modificar.")
	case errors.As(err, &codigo):
		if codigo.Catalogo == "cargos" {
			return models.ErrSolicitudInvalida("El cargo seleccionado no existe")
		}
		return models.ErrSolicitudInvalida("La pregunta de inhabilidad " + codigo.Codigo + " no existe")
	case errors.As(err, &dato):
		seccion, ok := seccionesFormulario[dato.Tabla]
		if !ok {
			seccion = "formulario"
		}
		return &models.AppError{
			Status:  http.StatusBadRequest,
			Code:    "dato_invalido",
			Message: "Revise la sección \"" + seccion + "\": hay un campo obligatorio vacío, un valor no permitido o un texto demasiado largo.",
			Extra:   map[string]any{"seccion": dato.Tabla, "restriccion": dato.Restriccion},
			Err:     err,
		}
	default:
		return models.ErrInterno(err)
	}
}
