package services

import (
	"context"
	"testing"
	"time"

	"biess-backend/models"
	"biess-backend/repositories"
)

func fecha(s string) *models.Fecha {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return models.NuevaFecha(t)
}

func codigoDe(err *models.AppError) string {
	if err == nil {
		return ""
	}
	return err.Code
}

var hoyPrueba = time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

const ecuadorPrueba = 52

// ---------------------------------------------------------------------------------------------
// Cálculo de experiencia (1.6)
// ---------------------------------------------------------------------------------------------

func TestCalcularDuracion(t *testing.T) {
	casos := []struct {
		inicio, fin string
		want        models.Duracion
	}{
		{"2020-01-01", "2020-12-31", models.Duracion{Anios: 1}},
		{"2015-01-01", "2019-12-31", models.Duracion{Anios: 5}},
		{"2020-01-15", "2020-03-14", models.Duracion{Meses: 2}},
		{"2023-05-10", "2023-05-20", models.Duracion{Dias: 11}},
		{"2019-03-20", "2024-06-30", models.Duracion{Anios: 5, Meses: 3, Dias: 11}},
	}
	for _, c := range casos {
		got := calcularDuracion(fecha(c.inicio).Time, fecha(c.fin).Time)
		if got != c.want {
			t.Errorf("calcularDuracion(%s, %s) = %+v; se esperaba %+v", c.inicio, c.fin, got, c.want)
		}
	}
}

func TestExperienciaTotalNoCuentaDosVecesLosCruces(t *testing.T) {
	exps := []models.Experiencia{
		{FechaInicio: fecha("2018-01-01"), FechaFin: fecha("2020-12-31")}, // se cruza con la siguiente
		{FechaInicio: fecha("2015-01-01"), FechaFin: fecha("2019-12-31")},
		{FechaInicio: fecha("2022-01-01"), FechaFin: fecha("2022-06-30")}, // separada
	}
	got := experienciaTotal(exps)
	want := models.Duracion{Anios: 6, Meses: 6} // 2015-2020 (6 años) + 6 meses de 2022
	if got != want {
		t.Errorf("experienciaTotal = %+v; se esperaba %+v", got, want)
	}
}

// ---------------------------------------------------------------------------------------------
// Nacionalidad y cargo (1.1.6 y 1.4)
// ---------------------------------------------------------------------------------------------

func TestEsEcuatoriana(t *testing.T) {
	if !esEcuatoriana(" ecuatoriana ") || esEcuatoriana("ESTADOUNIDENSE") {
		t.Error("la validación de nacionalidad no es correcta")
	}
}

func TestVerificarCargo(t *testing.T) {
	afiliado := models.HistorialLaboral{EsAfiliado: true}

	v, ok := verificarCargo("AFILIADOS", afiliado)
	if !ok || !v.Cumple || v.Estado != "Postulante afiliado al IESS" {
		t.Errorf("un afiliado debe cumplir para AFILIADOS: %+v", v)
	}
	v, _ = verificarCargo("JUBILADOS", afiliado)
	if v.Cumple || v.Estado != "Postulante no acredita jubilación" || v.Mensaje == "" {
		t.Errorf("un afiliado NO debe cumplir para JUBILADOS: %+v", v)
	}
	if _, ok := verificarCargo("PRESIDENTE", afiliado); ok {
		t.Error("un cargo inexistente debe rechazarse")
	}
}

// ---------------------------------------------------------------------------------------------
// Reglas al guardar
// ---------------------------------------------------------------------------------------------

func postulacionBase() *models.Postulacion {
	return &models.Postulacion{Cedula: "0104567896", Nacionalidad: "ECUATORIANA", EstadoCivil: "SOLTERO"}
}

func TestParejaObligatoriaSiEsCasado(t *testing.T) {
	p := postulacionBase()
	p.EstadoCivil = "CASADO"
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); codigoDe(err) != "solicitud_invalida" {
		t.Error("un postulante casado debe ingresar los datos de su pareja")
	}

	p.ParejaNombres, p.ParejaCedula = "ANA PEREZ", "0102030405" // dígito verificador incorrecto
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err == nil {
		t.Error("la cédula de la pareja debe ser válida")
	}

	p.ParejaCedula = "0102030400"
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err != nil {
		t.Errorf("datos de pareja correctos rechazados: %v", err)
	}
}

func TestParejaSeDescartaSiEsSoltero(t *testing.T) {
	p := postulacionBase()
	p.ParejaNombres, p.ParejaCedula = "ANA PEREZ", "0102030400"
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err != nil || p.ParejaNombres != "" || p.ParejaCedula != "" {
		t.Errorf("los datos de pareja de un soltero deben descartarse (err=%v)", err)
	}
}

func TestDomicilioEnOtroPaisUsaTexto(t *testing.T) {
	p := postulacionBase()
	otro, prov, canton := 40, 1, 1
	p.PaisID, p.ProvinciaID, p.CantonID = &otro, &prov, &canton
	p.ProvinciaTexto, p.CiudadTexto = "Florida", "Miami"
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err != nil || p.ProvinciaID != nil || p.CantonID != nil {
		t.Errorf("fuera de Ecuador no se usan los catálogos de provincia y cantón (err=%v)", err)
	}
}

func TestExperienciaConFechasInvalidas(t *testing.T) {
	p := postulacionBase()
	p.Experiencia = []models.Experiencia{{
		Tipo: "GENERAL", Cargo: "Analista", Institucion: "BCE", Actividades: "FINANZAS", DescripcionFunciones: "x",
		FechaInicio: fecha("2020-01-01"), FechaFin: fecha("2019-01-01"), DocumentoRespaldo: "exp.pdf",
	}}
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err == nil {
		t.Error("la fecha de finalización no puede ser anterior a la de inicio")
	}

	p.Experiencia[0].FechaFin = fecha("2027-01-01")
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err == nil {
		t.Error("la fecha de finalización no puede ser futura")
	}

	p.Experiencia[0].FechaFin = fecha("2021-01-01")
	p.Experiencia[0].DocumentoRespaldo = "exp.docx"
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err == nil {
		t.Error("el documento de respaldo debe ser PDF")
	}
}

func TestPreguntaRemocionExigeDatos(t *testing.T) {
	p := postulacionBase()
	p.Inhabilidades = []models.RespuestaInhabilidad{{Codigo: "1.9.7", Respuesta: true}}
	if err := validarGuardado(p, ecuadorPrueba, hoyPrueba); err == nil {
		t.Error("si responde Sí a 1.9.7 debe ingresar institución y motivo")
	}
}

// ---------------------------------------------------------------------------------------------
// Reglas al enviar
// ---------------------------------------------------------------------------------------------

func postulacionCompleta() *models.Postulacion {
	p := postulacionBase()
	ec, prov, canton, no := ecuadorPrueba, 1, 1, false
	p.PaisID, p.ProvinciaID, p.CantonID = &ec, &prov, &canton
	p.TelefonoCelular, p.CallePrincipal = "0991111111", "Av. Solano"
	p.CargoCodigo, p.RespaldadoOrganizacion = "AFILIADOS", &no
	p.Inhabilidades = []models.RespuestaInhabilidad{{Codigo: "1.9.1", Respuesta: true}, {Codigo: "1.9.2", Respuesta: false}}
	return p
}

var preguntasPrueba = []models.PreguntaInhabilidad{
	{Codigo: "1.9.1", RespuestaHabilitante: true},
	{Codigo: "1.9.2", RespuestaHabilitante: false},
}

func TestEnvio(t *testing.T) {
	if err := validarEnvio(postulacionCompleta(), ecuadorPrueba, preguntasPrueba); err != nil {
		t.Fatalf("una postulación completa debe poder enviarse: %v", err)
	}

	p := postulacionCompleta()
	p.Inhabilidades = p.Inhabilidades[:1]
	if err := validarEnvio(p, ecuadorPrueba, preguntasPrueba); codigoDe(err) != "postulacion_incompleta" {
		t.Errorf("faltan preguntas por responder: se obtuvo %q", codigoDe(err))
	}

	p = postulacionCompleta()
	p.Inhabilidades[1].Respuesta = true // "Sí, estoy inhabilitado para ejercer el comercio"
	if err := validarEnvio(p, ecuadorPrueba, preguntasPrueba); codigoDe(err) != "postulante_inhabilitado" {
		t.Errorf("una respuesta inhabilitante debe impedir el envío: se obtuvo %q", codigoDe(err))
	}
}

// ---------------------------------------------------------------------------------------------
// Service con repositorios FALSOS: gracias a la Inyección de Dependencias se prueba sin base de datos.
// ---------------------------------------------------------------------------------------------

type fakePostulaciones struct {
	repositories.PostulacionRepository // métodos no usados en estas pruebas
	guardados                          int
}

func (f *fakePostulaciones) Guardar(context.Context, *models.Postulacion) error {
	f.guardados++
	return nil
}

func (f *fakePostulaciones) BuscarPorPostulante(context.Context, int64) (*models.Postulacion, error) {
	return nil, repositories.ErrNoEncontrado
}

type fakePostulantes struct {
	repositories.PostulanteRepository
}

func (fakePostulantes) BuscarPerfilPorID(_ context.Context, id int64) (*models.PerfilPostulante, error) {
	cedulas := map[int64]string{1: "0104567896", 2: "1705678900"}
	return &models.PerfilPostulante{ID: id, Cedula: cedulas[id], Correo: "x@correo.com"}, nil
}

type fakeRegistro struct{}

func (fakeRegistro) BuscarPorCedula(_ context.Context, cedula string) (*models.RegistroCivil, error) {
	nac := "ECUATORIANA"
	if cedula == "1705678900" {
		nac = "ESTADOUNIDENSE"
	}
	return &models.RegistroCivil{Cedula: cedula, Nacionalidad: nac, EstadoCivil: "SOLTERO"}, nil
}

type fakeHistorial struct{}

func (fakeHistorial) BuscarPorCedula(context.Context, string) (*models.HistorialLaboral, error) {
	return &models.HistorialLaboral{EsAfiliado: true}, nil
}

type fakeCatalogos struct {
	repositories.CatalogoRepository
}

func (fakeCatalogos) ListarPaises(context.Context) ([]models.Pais, error) {
	return []models.Pais{{ID: ecuadorPrueba, Nombre: "Ecuador"}}, nil
}

func nuevoServicioPrueba(repo *fakePostulaciones) PostulacionService {
	return NewPostulacionService(repo, fakePostulantes{}, fakeRegistro{}, fakeHistorial{}, fakeCatalogos{},
		NewNotificadorPostulacionConsola())
}

func TestServicioRechazaExtranjeroSinGuardar(t *testing.T) {
	repo := &fakePostulaciones{}
	_, err := nuevoServicioPrueba(repo).Guardar(context.Background(), 2, models.PostulacionReq{})
	if ae, ok := models.EsAppError(err); !ok || ae.Code != "nacionalidad_no_permitida" {
		t.Fatalf("se esperaba nacionalidad_no_permitida, se obtuvo %v", err)
	}
	if repo.guardados != 0 {
		t.Error("la postulación de un extranjero NO debe llegar a la base de datos")
	}
}

func TestServicioRechazaCargoNoHabilitadoSinGuardar(t *testing.T) {
	repo := &fakePostulaciones{}
	_, err := nuevoServicioPrueba(repo).Guardar(context.Background(), 1, models.PostulacionReq{CargoCodigo: "jubilados"})
	if ae, ok := models.EsAppError(err); !ok || ae.Code != "cargo_no_habilitado" {
		t.Fatalf("un afiliado que postula por JUBILADOS debe ser rechazado, se obtuvo %v", err)
	}
	if repo.guardados != 0 {
		t.Error("una postulación rechazada NO debe llegar a la base de datos")
	}
}
