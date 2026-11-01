// Prueba rápida de la Fase 4: guardado de la postulación en UNA transacción (necesita Node 22+, el backend y Docker).
//   Desde la carpeta backend:   node pruebas/probar-fase4.mjs
// Usa la cédula 0104567896 (JOSE LUIS, afiliado, unión de hecho). Para repetirla desde cero:
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

console.log('\n=== Preparación: cuenta y sesión de 0104567896 ===')
const ident = await llamar('POST', '/auth/verificar-identidad', { cedula: '0104567896', codigo_dactilar: '1000000004' })
if (ident.status === 200) {
  await llamar('POST', '/auth/registro', { token_registro: ident.data.token_registro, correo: 'jose@correo.com', password: 'Clave1234' })
}
const login = await llamar('POST', '/auth/login', { cedula: '0104567896', password: 'Clave1234' })
mostrar('Login', login, 200)
const token = login.data?.token

console.log('\n=== Rutas protegidas y datos iniciales ===')
mostrar('GET /postulacion sin token', await llamar('GET', '/postulacion'), 401)
mostrar('GET /postulacion/opciones', await llamar('GET', '/postulacion/opciones', null, token), 200)
mostrar('GET /postulacion (precargada del Registro Civil)', await llamar('GET', '/postulacion', null, token), 200)

// Ids reales de la cascada País -> Provincia -> Cantón (nunca números fijos).
const paises = (await llamar('GET', '/catalogos/paises')).data.data
const ecuador = paises.find((p) => p.nombre === 'Ecuador')
const provincias = (await llamar('GET', `/catalogos/paises/${ecuador.id}/provincias`)).data.data
const azuay = provincias.find((p) => p.nombre === 'Azuay')
const pichincha = provincias.find((p) => p.nombre === 'Pichincha')
const cuenca = (await llamar('GET', `/catalogos/provincias/${azuay.id}/cantones`)).data.data.find((c) => c.nombre === 'Cuenca')
const quito = (await llamar('GET', `/catalogos/provincias/${pichincha.id}/cantones`)).data.data.find((c) => c.nombre === 'Quito')

const experiencia = (cargo, inicio, fin) => ({
  tipo: 'DIRECCION', cargo, institucion: 'Cooperativa Andina', fecha_inicio: inicio, fecha_fin: fin,
  actividades: 'FINANZAS', descripcion_funciones: 'Dirección financiera', documento_respaldo: 'exp.pdf',
})

const valida = {
  pareja_nombres: 'ANA MARIA PEREZ', pareja_cedula: '0102030400',
  pais_id: ecuador.id, provincia_id: azuay.id, canton_id: cuenca.id,
  telefono_celular: '0991111111', calle_principal: 'Av. Solano', numero: '1-23', sector: 'El Vergel',
  tiene_familiares_iess: true,
  familiares: [{ nombres_apellidos: 'PEDRO RAMIREZ', parentesco: 'TIO', institucion: 'IESS', area: 'Salud', cargo: 'Médico' }],
  cargo_codigo: 'AFILIADOS', respaldado_organizacion: false,
  formacion: [{
    nivel_instruccion: 'CUARTO_MAESTRIA', institucion: 'Universidad de Cuenca', tipo: 'NACIONAL', titulo: 'Magíster en Finanzas',
    area_estudios: 'FINANZAS', registro_senescyt: '1005-2019-123456', fecha_registro_senescyt: '2019-08-10', documento_respaldo: 'titulo.pdf',
  }],
  experiencia: [experiencia('Gerente', '2015-01-01', '2019-12-31'), experiencia('Director', '2020-01-01', '2024-06-30')],
  publicaciones: [],
  capacitaciones: [{
    tipo_evento: 'CURSO', nombre_evento: 'Riesgo financiero', institucion: 'BCE', fecha_inicio: '2023-03-01',
    fecha_fin: '2023-03-05', horas: 40, documento_respaldo: 'curso.pdf',
  }],
  inhabilidades: [{ codigo: '1.9.1', respuesta: true }, { codigo: '1.9.2', respuesta: false }],
}

console.log('\n=== Guardado transaccional (cabecera + 6 tablas hijas) ===')
const guardada = await llamar('PUT', '/postulacion', valida, token)
mostrar('Guardar postulación completa', guardada, 200)
comprobar(`Código único asignado: ${guardada.data?.codigo}`, /^BIESS-\d{4}-\d{6}$/.test(guardada.data?.codigo ?? ''))
comprobar('Se guardaron 2 experiencias, 1 familiar y 2 inhabilidades',
  guardada.data?.experiencia?.length === 2 && guardada.data?.familiares?.length === 1 && guardada.data?.inhabilidades?.length === 2)

console.log('\n=== ROLLBACK: un error en una tabla hija deshace TODO el guardado ===')
const conError = {
  ...valida,
  telefono_celular: '0999999999', // cambio en la cabecera...
  experiencia: [experiencia('Solo una', '2015-01-01', '2019-12-31')], // ...y en otra tabla hija...
  inhabilidades: [{ codigo: '1.9.1', respuesta: true }, { codigo: '9.9.9', respuesta: false }], // ...pero esta pregunta no existe
}
mostrar('Guardar con una pregunta de inhabilidad inexistente', await llamar('PUT', '/postulacion', conError, token), 400)
let actual = (await llamar('GET', '/postulacion', null, token)).data
comprobar('La cabecera NO cambió (teléfono sigue en 0991111111)', actual.telefono_celular === '0991111111')
comprobar('Las experiencias NO cambiaron (siguen siendo 2)', actual.experiencia.length === 2)

const fechasAlReves = { ...valida, telefono_celular: '0988888888', capacitaciones: [{ ...valida.capacitaciones[0], fecha_fin: '2023-02-01' }] }
mostrar('Guardar con fecha fin < fecha inicio (regla del servicio)', await llamar('PUT', '/postulacion', fechasAlReves, token), 400)
actual = (await llamar('GET', '/postulacion', null, token)).data
comprobar('Tras el ROLLBACK el teléfono sigue en 0991111111', actual.telefono_celular === '0991111111')

console.log('\n=== Integridad de la cascada (FK compuesta en la base) ===')
mostrar('Cantón Quito con provincia Azuay', await llamar('PUT', '/postulacion', { ...valida, canton_id: quito.id }, token), 400)
mostrar('Cargo inexistente', await llamar('PUT', '/postulacion', { ...valida, cargo_codigo: 'PRESIDENTE' }, token), 400)

console.log('\n=== Volver a guardar reemplaza las tablas hijas ===')
const segunda = await llamar('PUT', '/postulacion', { ...valida, experiencia: [experiencia('Director', '2020-01-01', '2024-06-30')] }, token)
mostrar('Guardar con 1 experiencia', segunda, 200)
comprobar('Mismo código y ahora 1 experiencia', segunda.data?.codigo === guardada.data?.codigo && segunda.data?.experiencia?.length === 1)
console.log('\nListo.')
