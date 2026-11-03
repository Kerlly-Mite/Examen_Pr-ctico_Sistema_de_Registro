package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"biess-backend/models"
	"biess-backend/repositories"
)

const (
	// Contextos del limitador: cada flujo lleva su propio contador de intentos fallidos.
	ctxRegistro     = "registro"
	ctxLogin        = "login"
	ctxRecuperacion = "recuperacion"

	// Prefijos de los tokens temporales guardados en Redis.
	prefTokenRegistro = "registro"
	prefTokenReset    = "reset"

	ttlTokenRegistro = 15 * time.Minute
	ttlTokenReset    = 30 * time.Minute

	layoutFecha = "2006-01-02"
)

// AuthService implementa EF-11: registro con validación de identidad, inicio de sesión y recuperación de contraseña.
type AuthService interface {
	// VerificarIdentidad es el paso 1 del registro: valida cédula + fecha de expedición o código dactilar.
	VerificarIdentidad(ctx context.Context, req models.VerificarIdentidadReq) (*models.VerificarIdentidadResp, error)
	// Registrar es el paso 2: con el token del paso 1 crea la cuenta (correo + contraseña).
	Registrar(ctx context.Context, req models.RegistroReq) error
	Login(ctx context.Context, req models.LoginReq) (*models.LoginResp, error)
	Perfil(ctx context.Context, postulanteID int64) (*models.PerfilPostulante, error)
	SolicitarRecuperacion(ctx context.Context, req models.VerificarIdentidadReq) error
	RestablecerPassword(ctx context.Context, req models.RestablecerReq) error
}

type authService struct {
	registro    repositories.RegistroCivilRepository
	postulantes repositories.PostulanteRepository
	limitador   repositories.LimitadorIntentos
	tokens      repositories.TokenStore
	jwt         GestorTokens
	notificador Notificador
	frontendURL string
	hashFalso   []byte // se compara contra él cuando la cédula no existe, para que tarde lo mismo (evita enumerar cuentas)
}

// NewAuthService recibe TODAS sus dependencias como interfaces (Inyección de Dependencias).
func NewAuthService(
	registro repositories.RegistroCivilRepository,
	postulantes repositories.PostulanteRepository,
	limitador repositories.LimitadorIntentos,
	tokens repositories.TokenStore,
	jwt GestorTokens,
	notificador Notificador,
	frontendURL string,
) AuthService {
	hashFalso, _ := bcrypt.GenerateFromPassword([]byte("contraseña-falsa-para-igualar-tiempos"), bcrypt.DefaultCost)
	return &authService{
		registro:    registro,
		postulantes: postulantes,
		limitador:   limitador,
		tokens:      tokens,
		jwt:         jwt,
		notificador: notificador,
		frontendURL: strings.TrimRight(frontendURL, "/"),
		hashFalso:   hashFalso,
	}
}

// ---------------------------------------------------------------------------------------------
// Registro (EF-11, puntos 4 y 5)
// ---------------------------------------------------------------------------------------------

func (s *authService) VerificarIdentidad(ctx context.Context, req models.VerificarIdentidadReq) (*models.VerificarIdentidadResp, error) {
	rc, err := s.validarIdentidad(ctx, ctxRegistro, req)
	if err != nil {
		return nil, err
	}

	// Solo se revela que la cédula ya tiene cuenta DESPUÉS de comprobar la identidad (no se puede usar para enumerar cuentas).
	if _, err := s.postulantes.BuscarPorCedula(ctx, rc.Cedula); err == nil {
		return nil, models.ErrConflicto("cedula_registrada", "Esta cédula ya está registrada. Inicie sesión.")
	} else if !errors.Is(err, repositories.ErrNoEncontrado) {
		return nil, models.ErrInterno(err)
	}

	token, err := nuevoToken()
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	if err := s.tokens.Guardar(ctx, prefTokenRegistro, token, rc.Cedula, ttlTokenRegistro); err != nil {
		return nil, models.ErrInterno(err)
	}
	return &models.VerificarIdentidadResp{
		TokenRegistro:    token,
		Nombres:          rc.Nombres,
		Apellidos:        rc.Apellidos,
		ExpiraEnSegundos: int(ttlTokenRegistro.Seconds()),
	}, nil
}

func (s *authService) Registrar(ctx context.Context, req models.RegistroReq) error {
	correo := strings.ToLower(strings.TrimSpace(req.Correo))
	if !correoValido(correo) {
		return models.ErrSolicitudInvalida("Ingrese un correo electrónico válido")
	}
	if errPass := validarPassword(req.Password); errPass != nil {
		return errPass
	}

	cedula, err := s.tokens.Obtener(ctx, prefTokenRegistro, req.TokenRegistro)
	if errors.Is(err, repositories.ErrNoEncontrado) {
		return models.NuevoError(400, "verificacion_expirada", "La verificación de identidad expiró. Vuelva a validar su identidad.")
	}
	if err != nil {
		return models.ErrInterno(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.ErrInterno(fmt.Errorf("generar hash: %w", err))
	}

	if _, err := s.postulantes.Crear(ctx, cedula, correo, string(hash)); err != nil {
		switch {
		case errors.Is(err, repositories.ErrCorreoDuplicado):
			return models.ErrConflicto("correo_registrado", "Este correo ya está en uso. Use otro correo electrónico.")
		case errors.Is(err, repositories.ErrCedulaDuplicada):
			return models.ErrConflicto("cedula_registrada", "Esta cédula ya está registrada. Inicie sesión.")
		default:
			return models.ErrInterno(err)
		}
	}

	// El token es de un solo uso: si falla el borrado, igual expira solo por TTL.
	_ = s.tokens.Eliminar(ctx, prefTokenRegistro, req.TokenRegistro)
	return nil
}

// ---------------------------------------------------------------------------------------------
// Inicio de sesión (EF-11, punto 6)
// ---------------------------------------------------------------------------------------------

func (s *authService) Login(ctx context.Context, req models.LoginReq) (*models.LoginResp, error) {
	cedula := strings.TrimSpace(req.Cedula)
	const msgIncorrectas = "Sus credenciales de ingreso son incorrectas"

	// Cédula mal formada o contraseña vacía: se rechaza sin tocar la base ni el contador de intentos.
	if !CedulaValida(cedula) || req.Password == "" {
		return nil, models.ErrIntentoFallido("credenciales_incorrectas", msgIncorrectas, 0)
	}

	if seg, err := s.limitador.Verificar(ctx, ctxLogin, cedula); err != nil {
		return nil, models.ErrInterno(err)
	} else if seg > 0 {
		return nil, s.errBloqueado(ctxLogin, seg)
	}

	p, err := s.postulantes.BuscarPorCedula(ctx, cedula)
	switch {
	case errors.Is(err, repositories.ErrNoEncontrado):
		_ = bcrypt.CompareHashAndPassword(s.hashFalso, []byte(req.Password)) // mismo tiempo de respuesta que una cuenta real
		return nil, s.registrarFallo(ctx, ctxLogin, cedula, "credenciales_incorrectas", msgIncorrectas)
	case err != nil:
		return nil, models.ErrInterno(err)
	}

	if bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(req.Password)) != nil {
		return nil, s.registrarFallo(ctx, ctxLogin, cedula, "credenciales_incorrectas", msgIncorrectas)
	}

	if err := s.limitador.Reiniciar(ctx, ctxLogin, cedula); err != nil {
		return nil, models.ErrInterno(err)
	}
	perfil, err := s.postulantes.BuscarPerfilPorID(ctx, p.ID)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	token, err := s.jwt.Generar(p.ID, p.Cedula)
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return &models.LoginResp{Token: token, Postulante: *perfil}, nil
}

func (s *authService) Perfil(ctx context.Context, postulanteID int64) (*models.PerfilPostulante, error) {
	perfil, err := s.postulantes.BuscarPerfilPorID(ctx, postulanteID)
	if errors.Is(err, repositories.ErrNoEncontrado) {
		return nil, models.ErrNoAutorizado("Su sesión no es válida. Inicie sesión nuevamente.")
	}
	if err != nil {
		return nil, models.ErrInterno(err)
	}
	return perfil, nil
}

// ---------------------------------------------------------------------------------------------
// Recuperación de contraseña (EF-11, punto 6)
// ---------------------------------------------------------------------------------------------

func (s *authService) SolicitarRecuperacion(ctx context.Context, req models.VerificarIdentidadReq) error {
	rc, err := s.validarIdentidad(ctx, ctxRecuperacion, req)
	if err != nil {
		return err
	}

	p, err := s.postulantes.BuscarPorCedula(ctx, rc.Cedula)
	if errors.Is(err, repositories.ErrNoEncontrado) {
		return nil // misma respuesta que si existiera: no se revela quién tiene cuenta
	}
	if err != nil {
		return models.ErrInterno(err)
	}

	token, err := nuevoToken()
	if err != nil {
		return models.ErrInterno(err)
	}
	if err := s.tokens.Guardar(ctx, prefTokenReset, token, strconv.FormatInt(p.ID, 10), ttlTokenReset); err != nil {
		return models.ErrInterno(err)
	}
	enlace := s.frontendURL + "/restablecer?token=" + token
	if err := s.notificador.EnviarRecuperacion(ctx, p.Correo, enlace); err != nil {
		return models.ErrInterno(err)
	}
	return nil
}

func (s *authService) RestablecerPassword(ctx context.Context, req models.RestablecerReq) error {
	if errPass := validarPassword(req.Password); errPass != nil {
		return errPass
	}
	valor, err := s.tokens.Obtener(ctx, prefTokenReset, req.Token)
	if errors.Is(err, repositories.ErrNoEncontrado) {
		return models.NuevoError(400, "enlace_invalido", "El enlace de recuperación no es válido o ya expiró. Solicite uno nuevo.")
	}
	if err != nil {
		return models.ErrInterno(err)
	}
	id, err := strconv.ParseInt(valor, 10, 64)
	if err != nil {
		return models.ErrInterno(fmt.Errorf("token de recuperación corrupto: %w", err))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.ErrInterno(fmt.Errorf("generar hash: %w", err))
	}
	if err := s.postulantes.ActualizarPassword(ctx, id, string(hash)); err != nil {
		return models.ErrInterno(err)
	}
	_ = s.tokens.Eliminar(ctx, prefTokenReset, req.Token)
	return nil
}

// ---------------------------------------------------------------------------------------------
// Apoyo: validación de identidad con control de intentos
// ---------------------------------------------------------------------------------------------

// validarIdentidad comprueba cédula + (fecha de expedición | código dactilar) contra el Registro Civil simulado.
// Aplica el bloqueo de EF-11: al tercer fallo consecutivo bloquea 10 minutos ese contexto para esa cédula.
func (s *authService) validarIdentidad(ctx context.Context, contexto string, req models.VerificarIdentidadReq) (*models.RegistroCivil, error) {
	cedula := strings.TrimSpace(req.Cedula)
	fecha := strings.TrimSpace(req.FechaExpedicion)
	dactilar := strings.TrimSpace(req.CodigoDactilar)

	// 1) Validaciones de formato: son errores de captura, NO cuentan como intento fallido.
	if !CedulaValida(cedula) {
		return nil, models.ErrSolicitudInvalida("La cédula ingresada no es válida")
	}
	if fecha == "" && dactilar == "" {
		return nil, models.ErrSolicitudInvalida("Ingrese la fecha de expedición o el código dactilar")
	}
	var fechaIngresada time.Time
	if fecha != "" {
		t, err := parseFecha(fecha)
		if err != nil {
			return nil, models.ErrSolicitudInvalida("La fecha de expedición debe tener el formato DD/MM/AAAA")
		}
		fechaIngresada = t
	}
	if dactilar != "" && !reDactilar.MatchString(dactilar) {
		return nil, models.ErrSolicitudInvalida("El código dactilar debe tener 10 caracteres (letras y números)")
	}

	// 2) ¿Hay un bloqueo vigente? Aunque los datos sean correctos, no se deja continuar.
	if seg, err := s.limitador.Verificar(ctx, contexto, cedula); err != nil {
		return nil, models.ErrInterno(err)
	} else if seg > 0 {
		return nil, s.errBloqueado(contexto, seg)
	}

	// 3) Comparación con el Registro Civil (simulado).
	rc, err := s.registro.BuscarPorCedula(ctx, cedula)
	if err != nil && !errors.Is(err, repositories.ErrNoEncontrado) {
		return nil, models.ErrInterno(err)
	}
	coincide := err == nil
	if coincide && fecha != "" {
		coincide = rc.FechaExpedicion.Format(layoutFecha) == fechaIngresada.Format(layoutFecha)
	}
	if coincide && dactilar != "" {
		coincide = strings.EqualFold(strings.TrimSpace(rc.CodigoDactilar), dactilar)
	}
	if !coincide {
		return nil, s.registrarFallo(ctx, contexto, cedula, "informacion_incorrecta", "La información ingresada es incorrecta")
	}

	if err := s.limitador.Reiniciar(ctx, contexto, cedula); err != nil {
		return nil, models.ErrInterno(err)
	}
	return rc, nil
}

// registrarFallo suma un intento fallido en Redis y devuelve el error adecuado:
// "intentos restantes" o, si fue el último, el bloqueo de 10 minutos.
func (s *authService) registrarFallo(ctx context.Context, contexto, id, codigo, mensaje string) error {
	estado, err := s.limitador.RegistrarFallo(ctx, contexto, id)
	if err != nil {
		return models.ErrInterno(err)
	}
	if estado.Bloqueado {
		return s.errBloqueado(contexto, estado.SegundosBloqueo)
	}
	return models.ErrIntentoFallido(codigo, mensaje, estado.Restantes)
}

func (s *authService) errBloqueado(contexto string, segundos int) *models.AppError {
	tiempo := textoTiempo(segundos)
	if contexto == ctxRegistro {
		return models.ErrBloqueado("Registro fallido, inténtelo nuevamente en "+tiempo, segundos)
	}
	return models.ErrBloqueado("Demasiados intentos fallidos. Inténtelo nuevamente en "+tiempo, segundos)
}

// textoTiempo convierte segundos a "N minutos" (redondeando hacia arriba).
func textoTiempo(segundos int) string {
	minutos := (segundos + 59) / 60
	if minutos <= 1 {
		return "1 minuto"
	}
	return fmt.Sprintf("%d minutos", minutos)
}

// nuevoToken genera un token aleatorio criptográficamente seguro de 256 bits (64 caracteres hexadecimales).
func nuevoToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generar token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
