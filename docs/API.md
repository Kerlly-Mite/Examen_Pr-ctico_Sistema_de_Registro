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
Mismo cuerpo que `verificar-identidad`. Si los datos coinciden, se envía un correo **real por SMTP** al correo de la
cuenta con el enlace `FRONTEND_URL/restablecer?token=...` (si `SMTP_HOST` no está configurado, el correo se muestra en la consola del backend).
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

## Postulación (EF-12) — Fase 4 🔒

Todas requieren `Authorization: Bearer <token>`. El usuario se toma del JWT: nadie puede leer ni guardar la postulación de otra persona.

| Método y ruta | Qué hace |
|---|---|
| `GET /postulacion/opciones` | Cargos (1.4) y preguntas de inhabilidades (1.9) para armar el formulario |
| `GET /postulacion` | La postulación de quien tiene la sesión. Si aún no la guarda, `estado: "NUEVA"` con la información personal precargada |
| `PUT /postulacion` | Guarda el formulario **completo** (cabecera + todas las secciones) en **una sola transacción** |

### `GET /postulacion/opciones`
```json
{
  "cargos": [ { "codigo": "AFILIADOS", "nombre": "Miembro Principal y Alterno ...", "sector": "AFILIADOS" }, ... ],
  "inhabilidades": [ { "codigo": "1.9.1", "texto": "¿Es Ecuatoriano/a ...?", "respuesta_habilitante": true }, ... ]
}
```
Cargos y preguntas se identifican por `codigo` (estable), nunca por id.

### `GET /postulacion` y respuesta de `PUT /postulacion`
```json
{
  "id": 1, "codigo": "BIESS-2026-000001", "estado": "BORRADOR", "enviada_en": null, "actualizado_en": "2026-11-01T15:40:57-05:00",
  "cedula": "0104567896", "nombres": "JOSE LUIS", "apellidos": "RAMIREZ SOLIS", "genero": "MASCULINO",
  "fecha_nacimiento": "1990-01-25", "nacionalidad": "ECUATORIANA", "estado_civil": "UNION_DE_HECHO",
  "correo_1": "jose@correo.com",
  "...": "más todos los campos del cuerpo del PUT (abajo)"
}
```
- `estado`: `NUEVA` (nunca guardada: `id` = 0 y `codigo` = ""), `BORRADOR` o `ENVIADA`.
- Información personal (1.1) y `correo_1` (1.2.5) son **solo lectura**: salen del Registro Civil y de la cuenta. Si se envían en el PUT, se ignoran.
- Las listas siempre llegan como `[]`, nunca `null`. Fechas en `AAAA-MM-DD`.

### Cuerpo de `PUT /postulacion`
Se envía el formulario completo cada vez (las listas reemplazan a las anteriores). Sirve para "Guardar progreso": los campos vacíos se permiten mientras no sean obligatorios en la base.
```json
{
  "pareja_nombres": "ANA MARIA PEREZ", "pareja_cedula": "0102030405",
  "pais_id": 52, "provincia_id": 1, "canton_id": 1,
  "provincia_texto": "", "ciudad_texto": "",
  "telefono_celular": "0991111111", "correo_2": "", "calle_principal": "Av. Solano", "numero": "1-23",
  "calle_secundaria": "", "sector": "El Vergel", "referencia": "", "telefono_domicilio": "",

  "tiene_familiares_iess": true,
  "familiares": [ { "nombres_apellidos": "PEDRO RAMIREZ", "parentesco": "TIO", "institucion": "IESS", "area": "Salud", "cargo": "Médico" } ],

  "cargo_codigo": "AFILIADOS", "respaldado_organizacion": false, "organizacion_nombre": "", "organizacion_documento": "",

  "formacion": [ { "nivel_instruccion": "CUARTO_MAESTRIA", "institucion": "Universidad de Cuenca", "tipo": "NACIONAL",
                   "titulo": "Magíster en Finanzas", "area_estudios": "FINANZAS", "registro_senescyt": "1005-2019-123456",
                   "fecha_registro_senescyt": "2019-08-10", "documento_respaldo": "titulo.pdf" } ],
  "experiencia": [ { "tipo": "DIRECCION", "cargo": "Director", "institucion": "Cooperativa Andina",
                     "fecha_inicio": "2020-01-01", "fecha_fin": "2024-06-30", "actividades": "FINANZAS",
                     "descripcion_funciones": "Dirección financiera", "documento_respaldo": "exp.pdf" } ],
  "publicaciones": [ { "titulo": "...", "tipo": "PUBLICACION", "medio_publicacion": "...", "fecha_publicacion": "2022-05-01",
                       "relacionada_formacion": true, "documento_respaldo": "pub.pdf" } ],
  "capacitaciones": [ { "tipo_evento": "CURSO", "nombre_evento": "Riesgo financiero", "institucion": "BCE",
                        "fecha_inicio": "2023-03-01", "fecha_fin": "2023-03-05", "horas": 40, "documento_respaldo": "curso.pdf" } ],
  "inhabilidades": [ { "codigo": "1.9.1", "respuesta": true },
                     { "codigo": "1.9.7", "respuesta": true, "remocion_institucion": "...", "remocion_cargo": "...",
                       "remocion_fecha_fin": "2022-01-31", "remocion_motivo": "..." } ]
}
```
Valores permitidos de las listas desplegables (los valida la base con `CHECK`):

| Campo | Valores |
|---|---|
| `parentesco` | `CONVIVIENTE`, `CONYUGE`, `CUNADO`, `HERMANO`, `HIJO_POLITICO`, `HIJO`, `MADRE`, `NIETO`, `NUERA`, `PADRE`, `PRIMO_HERMANO`, `SOBRINO`, `SUEGRO`, `TIO`, `YERNO` |
| `institucion` (familiar) | `IESS`, `BIESS` |
| `nivel_instruccion` | `TERCER_NIVEL`, `CUARTO_ESPECIALIDAD`, `CUARTO_MAESTRIA`, `CUARTO_PHD` |
| `tipo` (formación) | `NACIONAL`, `EXTRANJERA` |
| `area_estudios`, `actividades` | `FINANZAS`, `MERCADO_CAPITALES`, `ADMINISTRACION`, `ECONOMIA`, `DERECHO`, `AFINES` |
| `tipo` (experiencia) | `GENERAL`, `DIRECCION` |
| `tipo` (publicación) | `PUBLICACION`, `TRABAJO_INVESTIGACION` |
| `tipo_evento` | `CURSO`, `SEMINARIO`, `TALLER` |

- Ecuador: enviar `provincia_id` y `canton_id` de los catálogos. Otro país: `provincia_texto` y `ciudad_texto`.
- Si `tiene_familiares_iess` es `false`, la lista `familiares` se descarta. Si `respaldado_organizacion` no es `true`, se descartan los datos de la organización.
- `documento_respaldo` y `organizacion_documento`: el `id` que devuelve `POST /postulacion/documentos` (ver "Documentos PDF, formulario y correo").

**Errores de `PUT /postulacion`**

| HTTP | `error` | Cuándo | Campos extra |
|---|---|---|---|
| 400 | `solicitud_invalida` | JSON roto, fecha mal escrita, cargo o pregunta de inhabilidad inexistente | — |
| 400 | `dato_invalido` | La base rechazó un dato: campo obligatorio vacío en una sección, valor fuera de la lista, fecha fin antes de la de inicio, cantón que no pertenece a la provincia | `seccion` (tabla), `restriccion` |
| 401 | `no_autorizado` | Sin sesión | — |
| 409 | `postulacion_enviada` | La postulación ya se envió y no se puede modificar | — |

### Transacción (para la sustentación)
`repositories/postulacion_pg.go → Guardar`: `BEGIN` → `SELECT ... FOR UPDATE` (bloquea la fila contra dos guardados simultáneos) →
`INSERT` o `UPDATE` de `postulaciones` → `DELETE` + `INSERT` de las 6 tablas hijas → `COMMIT`.
Si cualquier paso falla, `defer tx.Rollback(ctx)` deshace todo: no quedan datos a medias ni huérfanos.
Se puede comprobar con `node pruebas/probar-fase4.mjs`.

## Reglas de negocio (EF-12) — Fase 5 🔒

### Rutas nuevas
| Método y ruta | Qué hace |
|---|---|
| `GET /postulacion/verificar-cargo/:codigo` | Consulta el Historial Laboral para el cargo (`AFILIADOS`, `JUBILADOS`, `EMPLEADORES`) sin guardar. Úsalo al elegir el cargo en el formulario (1.4.x) |
| `POST /postulacion/enviar` | Envía la postulación (1.10). Cuerpo: `{ "consentimiento_aceptado": true }` |

`GET /postulacion/verificar-cargo/AFILIADOS` → **200**
```json
{ "cargo_codigo": "AFILIADOS", "cumple": false, "estado": "Postulante No Afiliado al IESS",
  "mensaje": "Su postulación para el cargo de Miembro Principal y Alterno ... NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS",
  "pregunta": "¿Se encuentra respaldado por una organización de trabajadores ...?" }
```
Si `cumple` es `true`, mostrar `estado` y la `pregunta` del respaldo de organización (Sí/No). Si es `false`, mostrar `mensaje` y cerrar el formulario.

### Campos nuevos en la respuesta de `GET`/`PUT /postulacion`
```json
{
  "experiencia": [ { "...": "...", "duracion": { "anios": 5, "meses": 0, "dias": 0 } } ],
  "experiencia_total": { "anios": 6, "meses": 0, "dias": 0 },
  "verificacion_cargo": { "cargo_codigo": "JUBILADOS", "cumple": true, "estado": "Postulante acredita jubilación", "...": "..." },
  "bloqueo": null
}
```
- `duracion` (por experiencia) y `experiencia_total` los calcula el servidor (1.6). El total **no cuenta dos veces** los periodos que se cruzan.
- `bloqueo`: si no es `null` (`{ "codigo": "nacionalidad_no_permitida" | "cargo_no_habilitado", "mensaje": "..." }`), el frontend muestra el mensaje y no deja continuar.

### Reglas que valida el servidor
**Al guardar** (`PUT /postulacion`, "Guardar progreso": se permite incompleto, pero lo ingresado debe ser coherente):
| Regla | Error |
|---|---|
| 1.1.6 Solo nacionalidad ecuatoriana | 403 `nacionalidad_no_permitida` |
| 1.4.x El cargo exige ser afiliado / jubilado / empleador (Historial Laboral) | 403 `cargo_no_habilitado` (con `estado`) |
| 1.1.7.1 Casado o unión de hecho: nombres y cédula **válida** de la pareja (distinta de la propia). Otro estado civil: se descartan | 400 `solicitud_invalida` |
| 1.2 Ecuador: provincia y cantón del catálogo. Otro país: `provincia_texto` y `ciudad_texto` (se descartan los ids). Teléfonos solo números; correo 2 válido | 400 `solicitud_invalida` |
| 1.3 / 1.4 / 1.5-1.8 Cada registro agregado debe tener sus campos obligatorios; fechas no futuras; fecha fin ≥ inicio; horas > 0; descripción de funciones ≤ 1000 caracteres; documentos de respaldo en **PDF** | 400 `solicitud_invalida` (el `message` dice la sección y el número de registro) |
| 1.9.7 Si responde "Sí": institución y motivo obligatorios | 400 `solicitud_invalida` |

**Al enviar** (`POST /postulacion/enviar`):
| Regla | Error |
|---|---|
| Debe aceptar el consentimiento (1.10.1) | 400 `solicitud_invalida` |
| Debe estar guardada antes | 422 `postulacion_incompleta` |
| Domicilio (país, provincia/cantón o textos, celular, calle principal), cargo y respuesta del respaldo de organización completos; si dijo que tiene familiares, al menos uno | 422 `postulacion_incompleta` (con `seccion`) |
| Las 13 preguntas de inhabilidades respondidas | 422 `postulacion_incompleta` (con `preguntas` faltantes) |
| Ninguna respuesta inhabilitante (según `respuesta_habilitante` de `/opciones`) | 403 `postulante_inhabilitado` (con `preguntas`) |
| Ya enviada | 409 `postulacion_enviada` |
| Se guardó otro cambio mientras se enviaba (bloqueo optimista con `actualizado_en`) | 409 `postulacion_modificada` |

Al enviarse, el estado pasa a `ENVIADA` (ya no se puede modificar) y se envía por correo (SMTP) la confirmación con el formulario en PDF adjunto (1.10.3).

### Para la sustentación
- Las reglas están en `services/reglas_postulacion.go` como funciones puras; se prueban sin base de datos con `go test ./services/` (incluye pruebas del servicio con repositorios falsos gracias a la inyección de dependencias).
- El Historial Laboral está detrás de la interfaz `HistorialLaboralRepository`: para usar el servicio real del IESS solo cambia la implementación en `main.go`.
- Prueba completa contra la base: `node pruebas/probar-fase5.mjs`.

## Documentos PDF, formulario y correo 🔒

### Subir un PDF — `POST /postulacion/documentos`
`multipart/form-data` con dos campos:
- `archivo`: el PDF.
- `tipo`: `respaldo` (formación, experiencia, publicaciones, capacitación; **máx. 2 MB**) u `organizacion` (respaldo de organización del cargo; **máx. 4 MB**).

**201**
```json
{ "id": "9f2c...e1__titulo-maestria.pdf", "nombre": "titulo-maestria.pdf", "tamano": 48213 }
```
Guarda el `id` en `documento_respaldo` u `organizacion_documento` y luego haz el `PUT /postulacion`.
Al guardar, el servidor comprueba que cada documento exista y sea de quien guarda.

Ejemplo en el frontend:
```ts
const form = new FormData()
form.append('archivo', archivo)            // File del <input type="file" accept="application/pdf">
form.append('tipo', 'respaldo')
const res = await fetch('/api/postulacion/documentos', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, body: form })
// No poner Content-Type a mano: el navegador agrega el boundary del multipart.
```

| HTTP | `error` | Cuándo |
|---|---|---|
| 400 | `solicitud_invalida` | Falta el campo `archivo` o el `tipo` no es válido |
| 413 | `archivo_muy_grande` | Supera 2 MB (respaldo) o 4 MB (organización) |
| 415 | `formato_no_permitido` | El contenido no es un PDF (se revisa la firma `%PDF-` ... `%%EOF`, no solo la extensión) |
| 409 | `postulacion_enviada` | La postulación ya se envió |

### Ver un PDF — `GET /postulacion/documentos/:id`
Devuelve el PDF (`application/pdf`). Solo el dueño puede verlo (otro usuario recibe 404). Como la ruta pide el token,
en el frontend se descarga con `fetch` y se abre con `URL.createObjectURL(await res.blob())`.

### Formulario para imprimir — `GET /postulacion/formulario.pdf`
Genera el formulario completo (1.1 a 1.9), el texto de autorización y declaración (1.10.1) y la línea de **firma**
con nombres, apellidos y cédula (1.10.2). Sirve para el botón "Imprimir formulario". 422 si la postulación aún no se guarda.

### Correos reales (SMTP)
| Cuándo | Para | Contenido |
|---|---|---|
| `POST /auth/recuperar/solicitar` (EF-11) | Correo de la cuenta | Enlace para restablecer la contraseña (30 min) |
| `POST /postulacion/enviar` (1.10.3) | Correo 1 del postulante | Confirmación con el **formulario en PDF adjunto** |

Configuración en `.env` (ver `.env.example`): `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`.
Con Gmail: `SMTP_HOST=smtp.gmail.com`, `SMTP_PORT=587` y una **contraseña de aplicación** (no la contraseña normal).
Si `SMTP_HOST` está vacío, los correos se muestran en la consola del backend (el resto del sistema funciona igual).
El correo de confirmación se envía en segundo plano (goroutine): la respuesta del envío no espera al servidor de correo.

### Para la sustentación
- `services/correo.go`: interfaz `EnviadorCorreo` con dos implementaciones (SMTP real con STARTTLS/TLS y consola). `main.go` elige cuál según el `.env`.
- `services/notificador_correo.go`: un solo notificador implementa la interfaz de A (`Notificador`) y la de B (`NotificadorPostulacion`).
- `repositories/documentos_disco.go`: interfaz `DocumentoStore`; los PDF se guardan en `UPLOADS_DIR/<postulante_id>/` con nombre aleatorio (no se puede adivinar ni salir de la carpeta: protección contra *path traversal*).
- Pruebas: `go test ./...` (incluye un servidor SMTP de prueba en memoria) y `node pruebas/probar-correo-documentos.mjs`.
