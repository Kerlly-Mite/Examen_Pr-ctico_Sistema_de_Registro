# Contrato de la API — Fases 2 y 3

Base: `http://localhost:3001/api` (en el frontend, Vite lo reenvía: se llama simplemente a `/api/...`).
Todas las respuestas son JSON. Los errores siempre tienen la forma:

```json
{ "error": "codigo_estable", "message": "Texto en español para mostrar al usuario", ...campos extra }
```

El frontend debe decidir QUÉ HACER según `error` (código estable) y mostrar `message` al usuario.

## Códigos de error

| HTTP | `error` | Cuándo | Campos extra |
|---|---|---|---|
| 400 | `solicitud_invalida` | Formato incorrecto (cédula inválida, correo mal escrito, contraseña débil, JSON roto) | — |
| 400 | `verificacion_expirada` | El `token_registro` expiró (15 min) | — |
| 400 | `enlace_invalido` | El token de recuperación no existe o expiró (30 min) | — |
| 401 | `informacion_incorrecta` | Cédula + fecha/código dactilar no coinciden con el Registro Civil | `intentos_restantes` |
| 401 | `credenciales_incorrectas` | Cédula o contraseña incorrectas en el login | `intentos_restantes` |
| 401 | `no_autorizado` | Falta el token JWT, es inválido o expiró | — |
| 404 | `no_encontrado` | Recurso inexistente (p. ej. país o provincia) | — |
| 409 | `cedula_registrada` | La cédula ya tiene cuenta | — |
| 409 | `correo_registrado` | El correo ya está en uso | — |
| 429 | `cuenta_bloqueada` | 3 intentos fallidos consecutivos: bloqueo de 10 min | `segundos_restantes` |
| 500 | `error_interno` | Error inesperado | — |

Mensajes del documento funcional (EF-11): `"La información ingresada es incorrecta"`,
`"Registro fallido, inténtelo nuevamente en 10 minutos"`, `"Sus credenciales de ingreso son incorrectas"`.

## Autenticación (EF-11)

### 1. Verificar identidad — `POST /auth/verificar-identidad`
Paso 1 del registro. Enviar la cédula y **una** de las dos pruebas (fecha de expedición **o** código dactilar).

```json
{ "cedula": "0901234567", "fecha_expedicion": "21/06/2018" }
{ "cedula": "0901234567", "codigo_dactilar": "1000000001" }
```
`fecha_expedicion` acepta `DD/MM/AAAA` o `AAAA-MM-DD`. `codigo_dactilar`: 10 dígitos.

**200**
```json
{ "token_registro": "9f2c...", "nombres": "MARIA FERNANDA", "apellidos": "TORRES LEON", "expira_en_segundos": 900 }
```
Errores: 400, 401 `informacion_incorrecta` (con `intentos_restantes`), 409 `cedula_registrada`, 429 `cuenta_bloqueada`.

### 2. Registrar — `POST /auth/registro`
Paso 2: crea la cuenta con el token del paso 1.
```json
{ "token_registro": "9f2c...", "correo": "maria@correo.com", "password": "Clave1234" }
```
Contraseña: 8 a 72 caracteres, con al menos una letra y un número.
**201** `{ "message": "Registro exitoso. ..." }` → el frontend redirige al login.
Errores: 400 (`solicitud_invalida`, `verificacion_expirada`), 409 (`correo_registrado`, `cedula_registrada`).

### 3. Login — `POST /auth/login`
```json
{ "cedula": "0901234567", "password": "Clave1234" }
```
**200**
```json
{
  "token": "eyJhbGciOi...",
  "postulante": { "id": 1, "cedula": "0901234567", "correo": "maria@correo.com", "nombres": "MARIA FERNANDA", "apellidos": "TORRES LEON" }
}
```
Guardar `token` y enviarlo en cada petición protegida: `Authorization: Bearer <token>`. Dura 8 horas.
Errores: 401 `credenciales_incorrectas` (con `intentos_restantes`), 429 `cuenta_bloqueada`.

### 4. Sesión actual — `GET /auth/me` 🔒
Requiere `Authorization: Bearer <token>`. Devuelve el mismo objeto `postulante` del login.
Sirve para saber si la sesión sigue vigente al recargar la página (rutas protegidas del frontend).
Errores: 401 `no_autorizado`.

### 5. Olvidé mi contraseña — `POST /auth/recuperar/solicitar`
Mismo cuerpo que `verificar-identidad`. Si los datos coinciden, se "envía" un correo (**simulado**: el enlace
aparece en la consola del backend) con el enlace `FRONTEND_URL/restablecer?token=...`.
**200** `{ "message": "Si los datos son correctos, enviaremos un correo..." }` (siempre el mismo mensaje).
Errores: 400, 401 `informacion_incorrecta`, 429 `cuenta_bloqueada`.

### 6. Restablecer contraseña — `POST /auth/recuperar/restablecer`
```json
{ "token": "ab12...", "password": "NuevaClave99" }
```
**200** `{ "message": "Su contraseña fue actualizada. ..." }`. Error: 400 `enlace_invalido`.
El frontend necesita una ruta `/restablecer` que lea `?token=` de la URL.

## Catálogos (públicos, con caché Redis)

| Método y ruta | Respuesta |
|---|---|
| `GET /catalogos/paises` | `{ "data": [ { "id": 54, "nombre": "Ecuador" }, ... ] }` |
| `GET /catalogos/paises/:id/provincias` | `{ "data": [ { "id": 1, "pais_id": 54, "nombre": "Azuay" }, ... ] }` — lista vacía si el país no tiene provincias cargadas (se escribe a mano) |
| `GET /catalogos/provincias/:id/cantones` | `{ "data": [ { "id": 1, "provincia_id": 1, "nombre": "Cuenca" }, ... ] }` |

Los ids de la base pueden variar: el frontend debe usar los que devuelve la API, nunca números fijos.

## Para la integrante B (Fases 4-5): cómo proteger rutas
En `routes.go` ya existe `d.RequiereAuth`. Ejemplo:
```go
post := api.Group("/postulacion", d.RequiereAuth)
```
Dentro del handler, quién es el usuario:
```go
id, _ := c.Locals(middleware.LocalPostulanteID).(int64)
cedula, _ := c.Locals(middleware.LocalCedula).(string)
```

## Llaves que usa Redis (útil para la sustentación)

| Llave | Contenido | TTL |
|---|---|---|
| `intentos:{contexto}:{cedula}` | contador de fallos consecutivos | 10 min |
| `bloqueo:{contexto}:{cedula}` | marca de bloqueo | 10 min |
| `token:registro:{token}` | cédula verificada | 15 min |
| `token:reset:{token}` | id del postulante | 30 min |
| `catalogo:paises`, `catalogo:provincias:{id}`, `catalogo:cantones:{id}` | listas en JSON | 24 h |

`contexto` = `registro`, `login` o `recuperacion` (cada flujo cuenta aparte).
