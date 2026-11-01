// Prueba rápida de la Fase 5: reglas de negocio de EF-12 (necesita Node 22+, el backend y Docker encendidos).
//   Desde la carpeta backend:   node pruebas/probar-fase5.mjs
// Usa las personas simuladas de 03_seed_simulados.sql. Como al final la postulación queda ENVIADA,
// para repetirla hay que borrar las cuentas de prueba:
//   docker exec biess_redis redis-cli FLUSHALL
//   docker exec biess_postgres psql -U biess_user -d biess_postulacion -c "DELETE FROM postulantes"

const API = process.env.API ?? 'http://localhost:3001/api'

async function llamar(metodo, ruta, cuerpo, token) {
  const res = await fetch(`${API}${ruta}`, {
    method: metodo,
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: cuerpo ? JSON.stringify(cuerpo) : undefined,
  })
  return { status: res.status, data: await res.json().catch(() => null) }
}

function mostrar(titulo, r, esperado) {
  const marca = esperado === undefined ? ' ' : r.status === esperado ? 'OK ' : 'FALLO '
  const resumen = JSON.stringify(r.data)
  console.log(`${marca} [${r.status}] ${titulo}\n      ${resumen.length > 160 ? resumen.slice(0, 160) + '...' : resumen}`)
}

function comprobar(titulo, condicion) {
  console.log(`${condicion ? 'OK ' : 'FALLO '} ${titulo}`)
}

// Crea la cuenta (si no existe) e inicia sesión. Devuelve el token JWT.
async function sesion(cedula, codigoDactilar, correo) {
  const ident = await llamar('POST', '/auth/verificar-identidad', { cedula, codigo_dactilar: codigoDactilar })
  if (ident.status === 200) {
    await llamar('POST', '/auth/registro', { token_registro: ident.data.token_registro, correo, password: 'Clave1234' })
  }
  const login = await llamar('POST', '/auth/login', { cedula, password: 'Clave1234' })
  return login.data?.token
}

// Ids reales del catálogo (nunca números fijos).
const paises = (await llamar('GET', '/catalogos/paises')).data.data
const ecuador = paises.find((p) => p.nombre === 'Ecuador')
const pichincha = (await llamar('GET', `/catalogos/paises/${ecuador.id}/provincias`)).data.data.find((p) => p.nombre === 'Pichincha')
const quito = (await llamar('GET', `/catalogos/provincias/${pichincha.id}/cantones`)).data.data.find((c) => c.nombre === 'Quito')

const domicilio = {
  pais_id: ecuador.id, provincia_id: pichincha.id, canton_id: quito.id,
  telefono_celular: '0992222222', calle_principal: 'Av. Amazonas', numero: 'N24-10',
}

console.log('\n=== 1.1.6 Nacionalidad: 1705678900 (JOHN, estadounidense) ===')
const tExtranjero = await sesion('1705678900', '1000000005', 'john@correo.com')
const ext = await llamar('GET', '/postulacion', null, tExtranjero)
mostrar('GET /postulacion', ext, 200)
comprobar(`Viene bloqueada: ${ext.data?.bloqueo?.codigo}`, ext.data?.bloqueo?.codigo === 'nacionalidad_no_permitida')
mostrar('Intentar guardar', await llamar('PUT', '/postulacion', domicilio, tExtranjero), 403)

console.log('\n=== 1.1.7.1 Pareja y 1.4 cargo: 1702345677 (CARLOS, casado, empleador) ===')
const tCarlos = await sesion('1702345677', '1000000002', 'carlos@correo.com')
mostrar('Guardar sin datos de la pareja', await llamar('PUT', '/postulacion', domicilio, tCarlos), 400)
mostrar('Guardar con cédula de pareja inválida',
  await llamar('PUT', '/postulacion', { ...domicilio, pareja_nombres: 'ROSA VERA', pareja_cedula: '1234567890' }, tCarlos), 400)
const conPareja = { ...domicilio, pareja_nombres: 'ROSA VERA', pareja_cedula: '0102030400' }
mostrar('Guardar con pareja correcta', await llamar('PUT', '/postulacion', conPareja, tCarlos), 200)

const noAfiliado = await llamar('GET', '/postulacion/verificar-cargo/AFILIADOS', null, tCarlos)
mostrar('Verificar cargo AFILIADOS', noAfiliado, 200)
comprobar(`Estado: "${noAfiliado.data?.estado}"`, noAfiliado.data?.cumple === false && noAfiliado.data?.estado === 'Postulante No Afiliado al IESS')
mostrar('Guardar postulando por AFILIADOS', await llamar('PUT', '/postulacion', { ...conPareja, cargo_codigo: 'AFILIADOS' }, tCarlos), 403)
const empleador = await llamar('GET', '/postulacion/verificar-cargo/EMPLEADORES', null, tCarlos)
comprobar(`Verificar EMPLEADORES: "${empleador.data?.estado}"`, empleador.data?.cumple === true)

console.log('\n=== Flujo completo: 2403456789 (LUCIA, jubilada) ===')
const tLucia = await sesion('2403456789', '1000000003', 'lucia@correo.com')
const opciones = (await llamar('GET', '/postulacion/opciones', null, tLucia)).data
const habilitantes = opciones.inhabilidades.map((q) => ({ codigo: q.codigo, respuesta: q.respuesta_habilitante }))

const experiencia = (cargo, inicio, fin) => ({
  tipo: 'DIRECCION', cargo, institucion: 'Banco del Pacífico', fecha_inicio: inicio, fecha_fin: fin,
  actividades: 'FINANZAS', descripcion_funciones: 'Gestión financiera', documento_respaldo: 'exp.pdf',
})
const lucia = {
  ...domicilio,
  cargo_codigo: 'JUBILADOS', respaldado_organizacion: true,
  organizacion_nombre: 'Federación Nacional de Jubilados', organizacion_documento: 'respaldo.pdf',
  experiencia: [experiencia('Gerente', '2015-01-01', '2019-12-31'), experiencia('Asesora', '2018-01-01', '2020-12-31')],
  inhabilidades: habilitantes.slice(0, 5), // aún no responde todas
}

const g = await llamar('PUT', '/postulacion', lucia, tLucia)
mostrar('Guardar progreso (JUBILADOS, 2 experiencias que se cruzan)', g, 200)
comprobar(`Verificación del cargo: "${g.data?.verificacion_cargo?.estado}"`, g.data?.verificacion_cargo?.cumple === true)
comprobar(`Duración de la 1.ª experiencia: ${JSON.stringify(g.data?.experiencia?.[0]?.duracion)}`, g.data?.experiencia?.[0]?.duracion?.anios === 5)
comprobar(`Experiencia total sin contar el cruce dos veces: ${JSON.stringify(g.data?.experiencia_total)}`, g.data?.experiencia_total?.anios === 6)
mostrar('Documento de respaldo que no es PDF',
  await llamar('PUT', '/postulacion', { ...lucia, organizacion_documento: 'respaldo.docx' }, tLucia), 400)

console.log('\n=== 1.10 Envío de la postulación ===')
mostrar('Enviar sin aceptar el consentimiento', await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: false }, tLucia), 400)
mostrar('Enviar con preguntas sin responder', await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: true }, tLucia), 422)

const conInhabilidad = habilitantes.map((r) => (r.codigo === '1.9.8' ? { ...r, respuesta: true } : r)) // deudas con el SRI
await llamar('PUT', '/postulacion', { ...lucia, inhabilidades: conInhabilidad }, tLucia)
mostrar('Enviar con una inhabilidad (1.9.8 deudas con el SRI)', await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: true }, tLucia), 403)

await llamar('PUT', '/postulacion', { ...lucia, inhabilidades: habilitantes }, tLucia)
const enviada = await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: true }, tLucia)
mostrar('Enviar postulación completa (mira el correo simulado en la consola del backend)', enviada, 200)
comprobar(`Estado ${enviada.data?.estado}, código ${enviada.data?.codigo}`, enviada.data?.estado === 'ENVIADA' && enviada.data?.enviada_en)
mostrar('Modificar una postulación ya enviada', await llamar('PUT', '/postulacion', lucia, tLucia), 409)
mostrar('Enviarla otra vez', await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: true }, tLucia), 409)
console.log('\nListo.')
