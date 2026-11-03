// Prueba de la subida real de PDF, el formulario imprimible y el envío de correos (necesita Node 22+, el backend y Docker).
//   Desde la carpeta backend:   node pruebas/probar-correo-documentos.mjs
// Usa la cédula 0901234567 (MARIA FERNANDA, afiliada, soltera). El correo de la cuenta es el de la variable CORREO
// (por defecto maria@correo.com): para recibirlo de verdad, ejecútelo con su correo y con SMTP configurado en el .env:
//   PowerShell:  $env:CORREO="su_correo@gmail.com"; node pruebas/probar-correo-documentos.mjs
// Para repetirla desde cero:
//   docker exec biess_redis redis-cli FLUSHALL
//   docker exec biess_postgres psql -U biess_user -d biess_postulacion -c "DELETE FROM postulantes"

const API = process.env.API ?? 'http://localhost:3001/api'
const CORREO = process.env.CORREO ?? 'maria@correo.com'

async function llamar(metodo, ruta, cuerpo, token) {
  const res = await fetch(`${API}${ruta}`, {
    method: metodo,
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: cuerpo ? JSON.stringify(cuerpo) : undefined,
  })
  return { status: res.status, data: await res.json().catch(() => null) }
}

async function subir(token, contenido, nombre, tipo = 'respaldo') {
  const form = new FormData()
  form.append('archivo', new Blob([contenido], { type: 'application/pdf' }), nombre)
  form.append('tipo', tipo)
  const res = await fetch(`${API}/postulacion/documentos`, { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: form })
  return { status: res.status, data: await res.json().catch(() => null) }
}

async function descargar(ruta, token) {
  const res = await fetch(`${API}${ruta}`, { headers: token ? { Authorization: `Bearer ${token}` } : {} })
  return { status: res.status, tipo: res.headers.get('content-type'), datos: Buffer.from(await res.arrayBuffer()) }
}

function mostrar(titulo, r, esperado) {
  const marca = esperado === undefined ? ' ' : r.status === esperado ? 'OK ' : 'FALLO '
  const resumen = JSON.stringify(r.data ?? { tipo: r.tipo, bytes: r.datos?.length })
  console.log(`${marca} [${r.status}] ${titulo}\n      ${resumen.length > 160 ? resumen.slice(0, 160) + '...' : resumen}`)
}

function comprobar(titulo, condicion) {
  console.log(`${condicion ? 'OK ' : 'FALLO '} ${titulo}`)
}

async function sesion(cedula, codigoDactilar, correo) {
  const ident = await llamar('POST', '/auth/verificar-identidad', { cedula, codigo_dactilar: codigoDactilar })
  if (ident.status === 200) {
    await llamar('POST', '/auth/registro', { token_registro: ident.data.token_registro, correo, password: 'Clave1234' })
  }
  return (await llamar('POST', '/auth/login', { cedula, password: 'Clave1234' })).data?.token
}

const pdfValido = '%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n'
const pdfGrande = '%PDF-1.4\n' + 'x'.repeat(2.5 * 1024 * 1024) + '\n%%EOF\n' // 2.5 MB

const token = await sesion('0901234567', '1000000001', CORREO)
const otro = await sesion('0104567896', '1000000004', 'jose@correo.com')

console.log('\n=== Subida de documentos PDF ===')
const subido = await subir(token, pdfValido, 'Título Maestría.pdf')
mostrar('Subir un PDF válido', subido, 201)
mostrar('Subir un ejecutable renombrado a .pdf', await subir(token, 'MZ\x90\x00 programa', 'virus.pdf'), 415)
mostrar('Subir un PDF de 2.5 MB como respaldo (máx. 2 MB)', await subir(token, pdfGrande, 'grande.pdf'), 413)
const org = await subir(token, pdfGrande, 'respaldo-organizacion.pdf', 'organizacion')
mostrar('Subir el mismo PDF como respaldo de organización (máx. 4 MB)', org, 201)

console.log('\n=== Descarga: cada postulante solo ve sus documentos ===')
const propio = await descargar(`/postulacion/documentos/${subido.data.id}`, token)
mostrar('Descargar mi documento', propio, 200)
comprobar('Es el mismo archivo que se subió', propio.datos.toString() === pdfValido && propio.tipo === 'application/pdf')
mostrar('Otro postulante intenta descargarlo', await descargar(`/postulacion/documentos/${subido.data.id}`, otro), 404)
mostrar('Ruta maliciosa (../../.env)', await descargar(`/postulacion/documentos/..%2F..%2F.env`, token), 404)
mostrar('Sin sesión', await descargar(`/postulacion/documentos/${subido.data.id}`), 401)

console.log('\n=== Máximo 50 MB de documentos por postulante (EF-1, Restricciones) con 1702345677 ===')
const carlos = await sesion('1702345677', '1000000002', 'carlos@correo.com')
const pdf39 = '%PDF-1.4\n' + 'x'.repeat(3.9 * 1024 * 1024) + '\n%%EOF\n' // 3.9 MB (bajo el límite de 4 MB por archivo)
let aceptados = 0
let rechazo = null
for (let i = 1; i <= 14 && !rechazo; i++) {
  const r = await subir(carlos, pdf39, `respaldo-${i}.pdf`, 'organizacion')
  if (r.status === 201) aceptados++
  else rechazo = r
}
comprobar(`Se aceptan ${aceptados} archivos de 3.9 MB (${(aceptados * 3.9).toFixed(1)} MB) y el siguiente se rechaza`, aceptados === 12)
mostrar('Subir un archivo más allá de los 50 MB', rechazo ?? { status: 0, data: null }, 413)
comprobar(`Código del error: ${rechazo?.data?.error}`, rechazo?.data?.error === 'limite_documentos')

console.log('\n=== Guardar solo acepta documentos subidos de verdad ===')
const paises = (await llamar('GET', '/catalogos/paises')).data.data
const ecuador = paises.find((p) => p.nombre === 'Ecuador')
const guayas = (await llamar('GET', `/catalogos/paises/${ecuador.id}/provincias`)).data.data.find((p) => p.nombre === 'Guayas')
const gye = (await llamar('GET', `/catalogos/provincias/${guayas.id}/cantones`)).data.data.find((c) => c.nombre === 'Guayaquil')
const opciones = (await llamar('GET', '/postulacion/opciones', null, token)).data

const formacion = (doc) => [{
  nivel_instruccion: 'CUARTO_MAESTRIA', institucion: 'ESPOL', tipo: 'NACIONAL', titulo: 'Magíster en Finanzas',
  area_estudios: 'FINANZAS', registro_senescyt: '1006-2018-778899', fecha_registro_senescyt: '2018-09-14', documento_respaldo: doc,
}]
const postulacion = {
  pais_id: ecuador.id, provincia_id: guayas.id, canton_id: gye.id,
  telefono_celular: '0993333333', calle_principal: 'Av. 9 de Octubre', numero: '100', sector: 'Centro',
  cargo_codigo: 'AFILIADOS', respaldado_organizacion: true,
  organizacion_nombre: 'Confederación de Trabajadores', organizacion_documento: org.data.id,
  formacion: formacion(subido.data.id),
  inhabilidades: opciones.inhabilidades.map((q) => ({ codigo: q.codigo, respuesta: q.respuesta_habilitante })),
}
const inventado = 'a'.repeat(32) + '__no-existe.pdf'
mostrar('Guardar con un documento que nunca se subió', await llamar('PUT', '/postulacion', { ...postulacion, formacion: formacion(inventado) }, token), 400)
mostrar('Guardar con documentos reales', await llamar('PUT', '/postulacion', postulacion, token), 200)

console.log('\n=== Formulario imprimible (1.10.2) ===')
const formulario = await descargar('/postulacion/formulario.pdf', token)
mostrar('Descargar formulario.pdf', formulario, 200)
comprobar(`Es un PDF real (${formulario.datos.length} bytes)`, formulario.datos.subarray(0, 5).toString() === '%PDF-')

console.log(`\n=== Correos reales (llegan a ${CORREO} si SMTP está configurado; si no, salen en la consola del backend) ===`)
mostrar('Solicitar recuperación de contraseña', await llamar('POST', '/auth/recuperar/solicitar', { cedula: '0901234567', codigo_dactilar: '1000000001' }), 200)
const enviada = await llamar('POST', '/postulacion/enviar', { consentimiento_aceptado: true }, token)
mostrar('Enviar postulación (correo con el formulario PDF adjunto)', enviada, 200)
comprobar(`Estado ${enviada.data?.estado}`, enviada.data?.estado === 'ENVIADA')
mostrar('Subir documentos después de enviar', await subir(token, pdfValido, 'tarde.pdf'), 409)
console.log('\nListo. Revise la bandeja de entrada (y la carpeta de spam) o la consola del backend.')
