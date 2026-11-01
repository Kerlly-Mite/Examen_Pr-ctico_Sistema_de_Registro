package repositories

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"biess-backend/models"
)

type postulacionPG struct {
	db *pgxpool.Pool
}

// NewPostulacionRepository recibe el pool por parámetro (Inyección de Dependencias).
func NewPostulacionRepository(db *pgxpool.Pool) PostulacionRepository {
	return &postulacionPG{db: db}
}

// columnasCabecera son las columnas de postulaciones que se escriben al guardar (en este orden: $2, $3, ...).
var columnasCabecera = []string{
	"nombres", "apellidos", "genero", "fecha_nacimiento", "nacionalidad", "estado_civil",
	"pareja_nombres", "pareja_cedula",
	"pais_id", "provincia_id", "canton_id", "provincia_texto", "ciudad_texto",
	"telefono_celular", "correo_1", "correo_2", "calle_principal", "numero", "calle_secundaria",
	"sector", "referencia", "telefono_domicilio",
	"tiene_familiares_iess",
	"cargo_id", "respaldado_organizacion", "organizacion_nombre", "organizacion_documento",
}

// tablasHijas se vacían y se vuelven a llenar en cada guardado (el formulario se envía completo).
var tablasHijas = []string{
	"familiares_iess", "formacion_academica", "experiencia_profesional",
	"publicaciones", "capacitaciones", "postulacion_inhabilidades",
}

// ---------------------------------------------------------------------------------------------
// Escritura: una sola transacción para la cabecera y todas las tablas hijas
// ---------------------------------------------------------------------------------------------

func (r *postulacionPG) Guardar(ctx context.Context, p *models.Postulacion) error {
	// BEGIN
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("iniciar transacción: %w", err)
	}
	// ROLLBACK automático si se sale por cualquier error. Después del COMMIT no hace nada.
	defer tx.Rollback(ctx)

	cargoID, err := buscarCargoID(ctx, tx, p.CargoCodigo)
	if err != nil {
		return err
	}

	// FOR UPDATE bloquea la fila: si la misma persona guarda desde dos pestañas a la vez,
	// la segunda espera a que termine la primera (no se mezclan las tablas hijas).
	var estado string
	err = tx.QueryRow(ctx, `SELECT id, estado FROM postulaciones WHERE postulante_id = $1 FOR UPDATE`, p.PostulanteID).
		Scan(&p.ID, &estado)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		if err := insertarCabecera(ctx, tx, p, cargoID); err != nil {
			return err
		}
	case err != nil:
		return fmt.Errorf("leer postulación: %w", err)
	case estado == "ENVIADA":
		return ErrPostulacionEnviada
	default:
		if err := actualizarCabecera(ctx, tx, p, cargoID); err != nil {
			return err
		}
	}

	for _, tabla := range tablasHijas {
		if _, err := tx.Exec(ctx, `DELETE FROM `+tabla+` WHERE postulacion_id = $1`, p.ID); err != nil {
			return fmt.Errorf("limpiar %s: %w", tabla, err)
		}
	}
	if err := insertarHijas(ctx, tx, p); err != nil {
		return err
	}

	// COMMIT: recién aquí los cambios son visibles para los demás.
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmar transacción: %w", traducirError(err))
	}
	return nil
}

// MarcarEnviada usa un UPDATE condicional: solo cambia el estado si sigue en BORRADOR y si actualizado_en
// es el mismo que se leyó al validar (nadie guardó cambios entre la validación y el envío).
func (r *postulacionPG) MarcarEnviada(ctx context.Context, postulanteID int64, actualizadoEn time.Time) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE postulaciones
		SET estado = 'ENVIADA', consentimiento_aceptado = TRUE, enviada_en = now(), actualizado_en = now()
		WHERE postulante_id = $1 AND estado = 'BORRADOR' AND actualizado_en = $2`, postulanteID, actualizadoEn)
	if err != nil {
		return fmt.Errorf("enviar postulación: %w", traducirError(err))
	}
	if tag.RowsAffected() == 1 {
		return nil
	}
	// No se actualizó: se averigua por qué.
	var estado string
	err = r.db.QueryRow(ctx, `SELECT estado FROM postulaciones WHERE postulante_id = $1`, postulanteID).Scan(&estado)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNoEncontrado
	case err != nil:
		return fmt.Errorf("verificar estado de la postulación: %w", err)
	case estado == "ENVIADA":
		return ErrPostulacionEnviada
	default:
		return ErrPostulacionModificada
	}
}

func buscarCargoID(ctx context.Context, tx pgx.Tx, codigo string) (*int, error) {
	c := nulo(codigo)
	if c == nil {
		return nil, nil
	}
	var id int
	err := tx.QueryRow(ctx, `SELECT id FROM cargos WHERE codigo = $1`, *c).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &CodigoInexistenteError{Catalogo: "cargos", Codigo: *c}
	}
	if err != nil {
		return nil, fmt.Errorf("buscar cargo: %w", err)
	}
	return &id, nil
}

func valoresCabecera(p *models.Postulacion, cargoID *int) []any {
	return []any{
		p.Nombres, p.Apellidos, p.Genero, fechaDB(p.FechaNacimiento), p.Nacionalidad, p.EstadoCivil,
		nulo(p.ParejaNombres), nulo(p.ParejaCedula),
		p.PaisID, p.ProvinciaID, p.CantonID, nulo(p.ProvinciaTexto), nulo(p.CiudadTexto),
		nulo(p.TelefonoCelular), nulo(p.Correo1), nulo(p.Correo2), nulo(p.CallePrincipal), nulo(p.Numero), nulo(p.CalleSecundaria),
		nulo(p.Sector), nulo(p.Referencia), nulo(p.TelefonoDomicilio),
		p.TieneFamiliaresIESS,
		cargoID, p.RespaldadoOrganizacion, nulo(p.OrganizacionNombre), nulo(p.OrganizacionDocumento),
	}
}

func insertarCabecera(ctx context.Context, tx pgx.Tx, p *models.Postulacion, cargoID *int) error {
	marcas := make([]string, len(columnasCabecera))
	for i := range marcas {
		marcas[i] = "$" + strconv.Itoa(i+2)
	}
	sql := `INSERT INTO postulaciones (postulante_id, ` + strings.Join(columnasCabecera, ", ") + `)
	        VALUES ($1, ` + strings.Join(marcas, ", ") + `)
	        RETURNING id, codigo, estado`
	args := append([]any{p.PostulanteID}, valoresCabecera(p, cargoID)...)
	if err := tx.QueryRow(ctx, sql, args...).Scan(&p.ID, &p.Codigo, &p.Estado); err != nil {
		return fmt.Errorf("insertar postulación: %w", traducirError(err))
	}
	return nil
}

func actualizarCabecera(ctx context.Context, tx pgx.Tx, p *models.Postulacion, cargoID *int) error {
	asignaciones := make([]string, len(columnasCabecera))
	for i, col := range columnasCabecera {
		asignaciones[i] = col + " = $" + strconv.Itoa(i+2)
	}
	sql := `UPDATE postulaciones SET ` + strings.Join(asignaciones, ", ") + `, actualizado_en = now()
	        WHERE id = $1
	        RETURNING codigo, estado`
	args := append([]any{p.ID}, valoresCabecera(p, cargoID)...)
	if err := tx.QueryRow(ctx, sql, args...).Scan(&p.Codigo, &p.Estado); err != nil {
		return fmt.Errorf("actualizar postulación: %w", traducirError(err))
	}
	return nil
}

func insertarHijas(ctx context.Context, tx pgx.Tx, p *models.Postulacion) error {
	for _, f := range p.Familiares {
		if _, err := tx.Exec(ctx, `
			INSERT INTO familiares_iess (postulacion_id, nombres_apellidos, parentesco, institucion, area, cargo)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			p.ID, nulo(f.NombresApellidos), nulo(f.Parentesco), nulo(f.Institucion), nulo(f.Area), nulo(f.Cargo)); err != nil {
			return fmt.Errorf("insertar familiar: %w", traducirError(err))
		}
	}
	for _, f := range p.Formacion {
		if _, err := tx.Exec(ctx, `
			INSERT INTO formacion_academica (postulacion_id, nivel_instruccion, institucion, tipo, titulo, area_estudios,
			                                 registro_senescyt, fecha_registro_senescyt, documento_respaldo)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			p.ID, nulo(f.NivelInstruccion), nulo(f.Institucion), nulo(f.Tipo), nulo(f.Titulo), nulo(f.AreaEstudios),
			nulo(f.RegistroSenescyt), fechaDB(f.FechaRegistroSenescyt), nulo(f.DocumentoRespaldo)); err != nil {
			return fmt.Errorf("insertar formación: %w", traducirError(err))
		}
	}
	for _, e := range p.Experiencia {
		if _, err := tx.Exec(ctx, `
			INSERT INTO experiencia_profesional (postulacion_id, tipo, cargo, institucion, fecha_inicio, fecha_fin,
			                                     actividades, descripcion_funciones, documento_respaldo)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			p.ID, nulo(e.Tipo), nulo(e.Cargo), nulo(e.Institucion), fechaDB(e.FechaInicio), fechaDB(e.FechaFin),
			nulo(e.Actividades), nulo(e.DescripcionFunciones), nulo(e.DocumentoRespaldo)); err != nil {
			return fmt.Errorf("insertar experiencia: %w", traducirError(err))
		}
	}
	for _, pub := range p.Publicaciones {
		if _, err := tx.Exec(ctx, `
			INSERT INTO publicaciones (postulacion_id, titulo, tipo, medio_publicacion, fecha_publicacion,
			                           relacionada_formacion, documento_respaldo)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			p.ID, nulo(pub.Titulo), nulo(pub.Tipo), nulo(pub.MedioPublicacion), fechaDB(pub.FechaPublicacion),
			pub.RelacionadaFormacion, nulo(pub.DocumentoRespaldo)); err != nil {
			return fmt.Errorf("insertar publicación: %w", traducirError(err))
		}
	}
	for _, c := range p.Capacitaciones {
		if _, err := tx.Exec(ctx, `
			INSERT INTO capacitaciones (postulacion_id, tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin,
			                            horas, documento_respaldo)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			p.ID, nulo(c.TipoEvento), nulo(c.NombreEvento), nulo(c.Institucion), fechaDB(c.FechaInicio), fechaDB(c.FechaFin),
			c.Horas, nulo(c.DocumentoRespaldo)); err != nil {
			return fmt.Errorf("insertar capacitación: %w", traducirError(err))
		}
	}
	for _, i := range p.Inhabilidades {
		// La pregunta se identifica por su código ("1.9.7"); el id interno lo resuelve la base.
		tag, err := tx.Exec(ctx, `
			INSERT INTO postulacion_inhabilidades (postulacion_id, pregunta_id, respuesta,
			            remocion_institucion, remocion_cargo, remocion_fecha_fin, remocion_motivo)
			SELECT $1, id, $3, $4, $5, $6, $7 FROM inhabilidad_preguntas WHERE codigo = $2`,
			p.ID, strings.TrimSpace(i.Codigo), i.Respuesta,
			nulo(i.RemocionInstitucion), nulo(i.RemocionCargo), fechaDB(i.RemocionFechaFin), nulo(i.RemocionMotivo))
		if err != nil {
			return fmt.Errorf("insertar inhabilidad: %w", traducirError(err))
		}
		if tag.RowsAffected() == 0 {
			return &CodigoInexistenteError{Catalogo: "inhabilidad_preguntas", Codigo: i.Codigo}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------
// Lectura
// ---------------------------------------------------------------------------------------------

func (r *postulacionPG) BuscarPorPostulante(ctx context.Context, postulanteID int64) (*models.Postulacion, error) {
	p := models.Postulacion{PostulanteID: postulanteID}
	var fechaNac time.Time
	err := r.db.QueryRow(ctx, `
		SELECT p.id, p.codigo, p.estado, p.enviada_en, p.actualizado_en,
		       rc.cedula, p.nombres, p.apellidos, p.genero, p.fecha_nacimiento, p.nacionalidad, p.estado_civil,
		       COALESCE(p.pareja_nombres, ''), COALESCE(p.pareja_cedula, ''),
		       p.pais_id, p.provincia_id, p.canton_id, COALESCE(p.provincia_texto, ''), COALESCE(p.ciudad_texto, ''),
		       COALESCE(p.telefono_celular, ''), COALESCE(p.correo_1, ''), COALESCE(p.correo_2, ''),
		       COALESCE(p.calle_principal, ''), COALESCE(p.numero, ''), COALESCE(p.calle_secundaria, ''),
		       COALESCE(p.sector, ''), COALESCE(p.referencia, ''), COALESCE(p.telefono_domicilio, ''),
		       p.tiene_familiares_iess,
		       COALESCE(c.codigo, ''), p.respaldado_organizacion,
		       COALESCE(p.organizacion_nombre, ''), COALESCE(p.organizacion_documento, '')
		FROM postulaciones p
		JOIN postulantes po ON po.id = p.postulante_id
		JOIN registro_civil_simulado rc ON rc.cedula = po.cedula
		LEFT JOIN cargos c ON c.id = p.cargo_id
		WHERE p.postulante_id = $1`, postulanteID).
		Scan(&p.ID, &p.Codigo, &p.Estado, &p.EnviadaEn, &p.ActualizadoEn,
			&p.Cedula, &p.Nombres, &p.Apellidos, &p.Genero, &fechaNac, &p.Nacionalidad, &p.EstadoCivil,
			&p.ParejaNombres, &p.ParejaCedula,
			&p.PaisID, &p.ProvinciaID, &p.CantonID, &p.ProvinciaTexto, &p.CiudadTexto,
			&p.TelefonoCelular, &p.Correo1, &p.Correo2,
			&p.CallePrincipal, &p.Numero, &p.CalleSecundaria,
			&p.Sector, &p.Referencia, &p.TelefonoDomicilio,
			&p.TieneFamiliaresIESS,
			&p.CargoCodigo, &p.RespaldadoOrganizacion,
			&p.OrganizacionNombre, &p.OrganizacionDocumento)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("leer postulación: %w", err)
	}
	p.FechaNacimiento = models.NuevaFecha(fechaNac)

	if p.Familiares, err = leerHijas(ctx, r.db, `
		SELECT nombres_apellidos, parentesco, institucion, area, cargo
		FROM familiares_iess WHERE postulacion_id = $1 ORDER BY id`, p.ID,
		func(row pgx.CollectableRow) (models.Familiar, error) {
			var f models.Familiar
			err := row.Scan(&f.NombresApellidos, &f.Parentesco, &f.Institucion, &f.Area, &f.Cargo)
			return f, err
		}); err != nil {
		return nil, err
	}

	if p.Formacion, err = leerHijas(ctx, r.db, `
		SELECT nivel_instruccion, institucion, tipo, titulo, area_estudios, registro_senescyt,
		       fecha_registro_senescyt, documento_respaldo
		FROM formacion_academica WHERE postulacion_id = $1 ORDER BY id`, p.ID,
		func(row pgx.CollectableRow) (models.Formacion, error) {
			var f models.Formacion
			var fecha time.Time
			err := row.Scan(&f.NivelInstruccion, &f.Institucion, &f.Tipo, &f.Titulo, &f.AreaEstudios,
				&f.RegistroSenescyt, &fecha, &f.DocumentoRespaldo)
			f.FechaRegistroSenescyt = models.NuevaFecha(fecha)
			return f, err
		}); err != nil {
		return nil, err
	}

	if p.Experiencia, err = leerHijas(ctx, r.db, `
		SELECT tipo, cargo, institucion, fecha_inicio, fecha_fin, actividades, descripcion_funciones, documento_respaldo
		FROM experiencia_profesional WHERE postulacion_id = $1 ORDER BY id`, p.ID,
		func(row pgx.CollectableRow) (models.Experiencia, error) {
			var e models.Experiencia
			var inicio, fin time.Time
			err := row.Scan(&e.Tipo, &e.Cargo, &e.Institucion, &inicio, &fin, &e.Actividades,
				&e.DescripcionFunciones, &e.DocumentoRespaldo)
			e.FechaInicio, e.FechaFin = models.NuevaFecha(inicio), models.NuevaFecha(fin)
			return e, err
		}); err != nil {
		return nil, err
	}

	if p.Publicaciones, err = leerHijas(ctx, r.db, `
		SELECT titulo, tipo, medio_publicacion, fecha_publicacion, relacionada_formacion, documento_respaldo
		FROM publicaciones WHERE postulacion_id = $1 ORDER BY id`, p.ID,
		func(row pgx.CollectableRow) (models.Publicacion, error) {
			var pub models.Publicacion
			var fecha time.Time
			err := row.Scan(&pub.Titulo, &pub.Tipo, &pub.MedioPublicacion, &fecha, &pub.RelacionadaFormacion,
				&pub.DocumentoRespaldo)
			pub.FechaPublicacion = models.NuevaFecha(fecha)
			return pub, err
		}); err != nil {
		return nil, err
	}

	if p.Capacitaciones, err = leerHijas(ctx, r.db, `
		SELECT tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin, horas, documento_respaldo
		FROM capacitaciones WHERE postulacion_id = $1 ORDER BY id`, p.ID,
		func(row pgx.CollectableRow) (models.Capacitacion, error) {
			var c models.Capacitacion
			var inicio, fin time.Time
			err := row.Scan(&c.TipoEvento, &c.NombreEvento, &c.Institucion, &inicio, &fin, &c.Horas,
				&c.DocumentoRespaldo)
			c.FechaInicio, c.FechaFin = models.NuevaFecha(inicio), models.NuevaFecha(fin)
			return c, err
		}); err != nil {
		return nil, err
	}

	if p.Inhabilidades, err = leerHijas(ctx, r.db, `
		SELECT q.codigo, i.respuesta, COALESCE(i.remocion_institucion, ''), COALESCE(i.remocion_cargo, ''),
		       i.remocion_fecha_fin, COALESCE(i.remocion_motivo, '')
		FROM postulacion_inhabilidades i
		JOIN inhabilidad_preguntas q ON q.id = i.pregunta_id
		WHERE i.postulacion_id = $1 ORDER BY q.id`, p.ID,
		func(row pgx.CollectableRow) (models.RespuestaInhabilidad, error) {
			var i models.RespuestaInhabilidad
			var fecha *time.Time
			err := row.Scan(&i.Codigo, &i.Respuesta, &i.RemocionInstitucion, &i.RemocionCargo, &fecha, &i.RemocionMotivo)
			if fecha != nil {
				i.RemocionFechaFin = models.NuevaFecha(*fecha)
			}
			return i, err
		}); err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *postulacionPG) ListarCargos(ctx context.Context) ([]models.Cargo, error) {
	rows, err := r.db.Query(ctx, `SELECT codigo, nombre, sector FROM cargos ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("listar cargos: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Cargo, error) {
		var c models.Cargo
		err := row.Scan(&c.Codigo, &c.Nombre, &c.Sector)
		return c, err
	})
}

func (r *postulacionPG) ListarPreguntasInhabilidad(ctx context.Context) ([]models.PreguntaInhabilidad, error) {
	rows, err := r.db.Query(ctx, `SELECT codigo, texto, respuesta_habilitante FROM inhabilidad_preguntas ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("listar preguntas de inhabilidad: %w", err)
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.PreguntaInhabilidad, error) {
		var q models.PreguntaInhabilidad
		err := row.Scan(&q.Codigo, &q.Texto, &q.RespuestaHabilitante)
		return q, err
	})
}

// ---------------------------------------------------------------------------------------------
// Utilidades
// ---------------------------------------------------------------------------------------------

// leerHijas ejecuta la consulta de una tabla hija y arma la lista (vacía, nunca nil, para que el JSON sea []).
func leerHijas[T any](ctx context.Context, db *pgxpool.Pool, sql string, postulacionID int64,
	scan func(pgx.CollectableRow) (T, error)) ([]T, error) {
	rows, err := db.Query(ctx, sql, postulacionID)
	if err != nil {
		return nil, fmt.Errorf("leer tabla hija: %w", err)
	}
	lista, err := pgx.CollectRows(rows, scan)
	if err != nil {
		return nil, fmt.Errorf("leer tabla hija: %w", err)
	}
	if lista == nil {
		lista = []T{}
	}
	return lista, nil
}

// nulo convierte el texto vacío en NULL: así las columnas NOT NULL de la base rechazan los campos obligatorios vacíos.
func nulo(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func fechaDB(f *models.Fecha) *time.Time {
	if f == nil {
		return nil
	}
	t := f.Time
	return &t
}

// traducirError convierte las violaciones de restricciones de PostgreSQL en DatoInvalidoError.
// 23514 CHECK | 23502 NOT NULL | 23503 FOREIGN KEY | 22001 texto muy largo | 22007/22008 fecha inválida.
func traducirError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23514", "23502", "23503", "22001", "22007", "22008":
		restriccion := pgErr.ConstraintName
		if restriccion == "" {
			restriccion = pgErr.ColumnName
		}
		return &DatoInvalidoError{Tabla: pgErr.TableName, Restriccion: restriccion, Causa: err}
	}
	return err
}
