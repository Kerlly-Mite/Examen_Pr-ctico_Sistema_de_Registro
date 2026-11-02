package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"biess-backend/models"
	"biess-backend/repositories"
)

// PostulacionService implementa EF-12: consultar, guardar y enviar el formulario de postulación.
type PostulacionService interface {
	// Obtener devuelve la postulación guardada o, si aún no existe, una nueva (estado NUEVA)
	// con la información personal precargada desde el Registro Civil.
	Obtener(ctx context.Context, postulanteID int64) (*models.Postulacion, error)
	// Guardar valida las reglas de negocio y persiste el formulario completo de forma atómica.
	Guardar(ctx context.Context, postulanteID int64, req models.PostulacionReq) (*models.Postulacion, error)
	// VerificarCargo consulta el Historial Laboral para el cargo elegido (1.4.x), sin guardar nada.
	VerificarCargo(ctx context.Context, postulanteID int64, cargoCodigo string) (*models.VerificacionCargo, error)
	// Enviar valida que el formulario esté completo y lo pasa a ENVIADA (1.10).
	Enviar(ctx context.Context, postulanteID int64, req models.EnviarReq) (*models.Postulacion, error)
	// Opciones devuelve los cargos y las preguntas de inhabilidades para armar el formulario.
	Opciones(ctx context.Context) (*models.OpcionesPostulacion, error)
}

// NotificadorPostulacion envía al correo del postulante la confirmación del envío (1.10.3).
// Se SIMULA en consola, igual que el correo de recuperación de contraseña.
type NotificadorPostulacion interface {
	EnviarConfirmacion(ctx context.Context, correo, codigo string) error
}

type notificadorPostulacionConsola struct{}

// NewNotificadorPostulacionConsola devuelve el notificador simulado (escribe en la consola del backend).
func NewNotificadorPostulacionConsola() NotificadorPostulacion {
	return notificadorPostulacionConsola{}
}

func (notificadorPostulacionConsola) EnviarConfirmacion(_ context.Context, correo, codigo string) error {
	log.Printf("[CORREO SIMULADO] Para: %s | Su postulación %s fue enviada. Adjunto: formulario de postulación.", correo, codigo)
	return nil
}

type postulacionService struct {
	postulaciones repositories.PostulacionRepository
	postulantes   repositories.PostulanteRepository
	registro      repositories.RegistroCivilRepository
	historial     repositories.HistorialLaboralRepository
	catalogos     repositories.CatalogoRepository
	notificador   NotificadorPostulacion
	ahora         func() time.Time // inyectable para pruebas
}

// NewPostulacionService recibe sus dependencias como interfaces (Inyección de Dependencias).
// El repositorio de catálogos que llega desde main.go es el que tiene caché Redis.
func NewPostulacionService(
	postulaciones repositories.PostulacionRepository,
	postulantes repositories.PostulanteRepository,
	registro repositories.RegistroCivilRepository,
	historial repositories.HistorialLaboralRepository,
	catalogos repositories.CatalogoRepository,
	notificador NotificadorPostulacion,
) PostulacionService {
	return &postulacionService{
		postulaciones: postulaciones,
		postulantes:   postulantes,
		registro:      registro,
		historial:     historial,
		catalogos:     catalogos,
		notificador:   notificador,
		ahora:         time.Now,
	}
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
	switch {
	case errors.Is(err, repositories.ErrNoEncontrado):
		// Primera vez: formulario vacío con los datos del Registro Civil y el correo de la cuenta.
		if p, err = s.armarPostulacion(ctx, postulanteID, perfil, models.PostulacionReq{}); err != nil {
			return nil, err
		}
		p.Estado = "NUEVA"
	case err != nil:
		return nil, models.ErrInterno(err)
	}

	if err := s.completarCalculados(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
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

	// 1.1.6 Si la nacionalidad no es ecuatoriana, no se permite continuar.
	if !esEcuatoriana(p.Nacionalidad) {
		return nil, errNacionalidad()
	}
	ecuadorID, err := s.ecuadorID(ctx)
	if err != nil {
		return nil, err
	}
	if errRegla := validarGuardado(p, ecuadorID, s.hoy()); errRegla != nil {
		return nil, errRegla
	}
	// 1.4.x El cargo elegido exige ser afiliado, jubilado o empleador según el Historial Laboral.
	if p.CargoCodigo != "" {
		v, err := s.verificar(ctx, p.Cedula, p.CargoCodigo)
		if err != nil {
			return nil, err
		}
		if !v.Cumple {
			return nil, errCargo(v)
		}
	}

	if err := s.postulaciones.Guardar(ctx, p); err != nil {
		return nil, traducirErrorPostulacion(err)
	}
	// Se devuelve lo que realmente quedó guardado (código único, estado, fechas, cálculos).
	return s.Obtener(ctx, postulanteID)
}

func (s *postulacionService) VerificarCargo(ctx context.Context, postulanteID int64, cargoCodigo string) (*models.VerificacionCargo, error) {
	perfil, err := s.perfil(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	return s.verificar(ctx, perfil.Cedula, strings.ToUpper(strings.TrimSpace(cargoCodigo)))
}

func (s *postulacionService) Enviar(ctx context.Context, postulanteID int64, req models.EnviarReq) (*models.Postulacion, error) {
	// 1.10.1 Sin aceptar el texto de consentimiento no se puede enviar.
	if !req.ConsentimientoAceptado {
		return nil, models.ErrSolicitudInvalida("Debe aceptar la autorización y declaración de consentimiento para enviar su postulación")
	}

	p, err := s.Obtener(ctx, postulanteID)
	if err != nil {
		return nil, err
	}
	switch {
	case p.Estado == "NUEVA":
		return nil, models.NuevoError(http.StatusUnprocessableEntity, "postulacion_incompleta", "Guarde su postulación antes de enviarla")
	case p.Estado == "ENVIADA":
		return nil, models.ErrConflicto("postulacion_enviada", "Su postulación ya fue enviada.")
	case p.Bloqueo != nil:
		return nil, models.NuevoError(http.StatusForbidden, p.Bloqueo.Codigo, p.Bloqueo.Mensaje)
	}

	ecuadorID, err := s.ecuadorID(ctx)
	if err != nil {
		return nil, err
	}
	preguntas, err := s.postulaciones.ListarPreguntasInhabilidad(ctx)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	if errRegla := validarEnvio(p, ecuadorID, preguntas); errRegla != nil {
		return nil, errRegla
	}

	if err := s.postulaciones.MarcarEnviada(ctx, postulanteID, *p.ActualizadoEn); err != nil {
		return nil, traducirErrorPostulacion(err)
	}
	// 1.10.3 El formulario se envía al correo del postulante. Si el correo falla, el envío ya quedó registrado.
	if err := s.notificador.EnviarConfirmacion(ctx, p.Correo1, p.Codigo); err != nil {
		log.Printf("[postulacion] no se pudo enviar el correo de confirmación de %s: %v", p.Codigo, err)
	}
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

// completarCalculados agrega lo que calcula el servidor: duración de cada experiencia y total (1.6),
// verificación del cargo (1.4.x) y bloqueo por nacionalidad o por cargo (el frontend cierra el formulario).
func (s *postulacionService) completarCalculados(ctx context.Context, p *models.Postulacion) error {
	for i := range p.Experiencia {
		e := &p.Experiencia[i]
		if e.FechaInicio != nil && e.FechaFin != nil {
			d := calcularDuracion(e.FechaInicio.Time, e.FechaFin.Time)
			e.Duracion = &d
		}
	}
	total := experienciaTotal(p.Experiencia)
	p.ExperienciaTotal = &total

	if !esEcuatoriana(p.Nacionalidad) {
		p.Bloqueo = &models.Bloqueo{Codigo: "nacionalidad_no_permitida", Mensaje: msgNacionalidad}
		return nil
	}
	if p.CargoCodigo != "" {
		v, err := s.verificar(ctx, p.Cedula, p.CargoCodigo)
		if err != nil {
			return err
		}
		p.VerificacionCargo = v
		if !v.Cumple {
			p.Bloqueo = &models.Bloqueo{Codigo: "cargo_no_habilitado", Mensaje: v.Mensaje}
		}
	}
	return nil
}

// verificar consulta el Historial Laboral (simulado) y aplica la regla del cargo.
func (s *postulacionService) verificar(ctx context.Context, cedula, cargoCodigo string) (*models.VerificacionCargo, error) {
	h, err := s.historial.BuscarPorCedula(ctx, cedula)
	switch {
	case errors.Is(err, repositories.ErrNoEncontrado):
		h = &models.HistorialLaboral{} // sin historial: no es afiliado, ni jubilado, ni empleador
	case err != nil:
		return nil, models.ErrInterno(err)
	}
	v, ok := verificarCargo(cargoCodigo, *h)
	if !ok {
		return nil, models.ErrSolicitudInvalida("El cargo seleccionado no existe")
	}
	return v, nil
}

// ecuadorID busca el id de Ecuador en el catálogo de países (con caché Redis): los ids pueden variar entre bases.
func (s *postulacionService) ecuadorID(ctx context.Context) (int, error) {
	paises, err := s.catalogos.ListarPaises(ctx)
	if err != nil {
		return 0, models.ErrInterno(err)
	}
	for _, p := range paises {
		if strings.EqualFold(p.Nombre, "Ecuador") {
			return p.ID, nil
		}
	}
	return 0, models.ErrInterno(fmt.Errorf("el país Ecuador no está en el catálogo"))
}

func (s *postulacionService) hoy() time.Time {
	y, m, d := s.ahora().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
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

// normalizar limpia el formato de la petición (mayúsculas, espacios). Las reglas que descartan
// datos dependientes ("No" en una pregunta) están en reglas_postulacion.go.
func normalizar(req models.PostulacionReq) models.PostulacionReq {
	req.CargoCodigo = strings.ToUpper(strings.TrimSpace(req.CargoCodigo))
	req.Correo2 = strings.ToLower(strings.TrimSpace(req.Correo2))
	for i := range req.Experiencia {
		req.Experiencia[i].Duracion = nil // lo calcula el servidor
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
	case errors.Is(err, repositories.ErrPostulacionModificada):
		return models.ErrConflicto("postulacion_modificada", "Su postulación se modificó mientras se enviaba. Revísela e intente enviarla nuevamente.")
	case errors.Is(err, repositories.ErrNoEncontrado):
		return models.NuevoError(http.StatusUnprocessableEntity, "postulacion_incompleta", "Guarde su postulación antes de enviarla")
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
