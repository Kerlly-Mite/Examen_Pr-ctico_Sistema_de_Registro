// Prueba rápida de las Fases 2 y 3 (necesita Node 22+, el backend y Docker encendidos).
//   Desde la carpeta backend:   node pruebas/probar-fase2-3.mjs
// Para repetirla sin esperar 10 minutos, borra los bloqueos y las cuentas de prueba:
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

console.log('\n=== FASE 3: caché de catálogos (mira la consola del backend: MISS la 1.ª vez, HIT la 2.ª) ===')
mostrar('GET paises (1.ª vez)', await llamar('GET', '/catalogos/paises'), 200)
mostrar('GET paises (2.ª vez)', await llamar('GET', '/catalogos/paises'), 200)

console.log('\n=== FASE 2: bloqueo en el registro (EF-11) con la cédula 0901234567 ===')
const mala = { cedula: '0901234567', fecha_expedicion: '01/01/2000' }
mostrar('Intento fallido 1', await llamar('POST', '/auth/verificar-identidad', mala), 401)
mostrar('Intento fallido 2', await llamar('POST', '/auth/verificar-identidad', mala), 401)
mostrar('Intento fallido 3 -> bloqueo', await llamar('POST', '/auth/verificar-identidad', mala), 429)
mostrar('Datos CORRECTOS pero bloqueada', await llamar('POST', '/auth/verificar-identidad', { cedula: '0901234567', fecha_expedicion: '21/06/2018' }), 429)

console.log('\n=== FASE 2: registro completo con la cédula 1702345677 ===')
const id = await llamar('POST', '/auth/verificar-identidad', { cedula: '1702345677', codigo_dactilar: '1000000002' })
mostrar('Verificar identidad (código dactilar)', id, 200)
const token = id.data?.token_registro
mostrar('Registro con contraseña débil', await llamar('POST', '/auth/registro', { token_registro: token, correo: 'carlos@correo.com', password: '123' }), 400)
mostrar('Registro correcto', await llamar('POST', '/auth/registro', { token_registro: token, correo: 'carlos@correo.com', password: 'Clave1234' }), 201)
mostrar('Reusar el mismo token', await llamar('POST', '/auth/registro', { token_registro: token, correo: 'otro@correo.com', password: 'Clave1234' }), 400)

console.log('\n=== FASE 2: login y ruta protegida ===')
const login = await llamar('POST', '/auth/login', { cedula: '1702345677', password: 'Clave1234' })
mostrar('Login correcto', login, 200)
mostrar('GET /auth/me sin token', await llamar('GET', '/auth/me'), 401)
mostrar('GET /auth/me con token', await llamar('GET', '/auth/me', null, login.data?.token), 200)

console.log('\n=== FASE 2: bloqueo en el login ===')
const mal = { cedula: '1702345677', password: 'Incorrecta99' }
mostrar('Login fallido 1', await llamar('POST', '/auth/login', mal), 401)
mostrar('Login fallido 2', await llamar('POST', '/auth/login', mal), 401)
mostrar('Login fallido 3 -> bloqueo', await llamar('POST', '/auth/login', mal), 429)
mostrar('Contraseña CORRECTA pero bloqueada', await llamar('POST', '/auth/login', { cedula: '1702345677', password: 'Clave1234' }), 429)

console.log('\n=== FASE 2: recuperación de contraseña (el enlace sale en la consola del backend) ===')
mostrar('Solicitar recuperación', await llamar('POST', '/auth/recuperar/solicitar', { cedula: '1702345677', fecha_expedicion: '2019-02-11' }), 200)
mostrar('Restablecer con token falso', await llamar('POST', '/auth/recuperar/restablecer', { token: 'falso', password: 'NuevaClave99' }), 400)
console.log('\nListo.')
