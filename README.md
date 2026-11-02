# Sistema de postulación al Directorio del BIESS

Módulo web para el concurso público de méritos y oposición de los Miembros Principales y Alternos del Directorio del BIESS, en cumplimiento de la Resolución C.D. 701 del IESS. Implementa el requerimiento funcional EF-11 (registro y autenticación) y EF-12 (formulario de postulación). Proyecto del examen práctico del primer parcial (Frameworks de Programación Web), desarrollado por tres integrantes.

## Qué hace

- **Registro** en dos pasos: verificación de identidad con cédula y fecha de expedición o código dactilar (simulado), y luego correo y contraseña.
- **Login** con JWT (8 horas), recuperación de contraseña y **bloqueo de 10 minutos** tras 3 intentos fallidos.
- **Formulario de postulación** por secciones: información personal, domicilio con cascada País → Provincia → Cantón, familiares en el IESS/BIESS, cargo, formación, experiencia, publicaciones, capacitación e inhabilidades.
- **Reglas de negocio** en el servidor: solo nacionalidad ecuatoriana, validación de la calidad de afiliado, jubilado o empleador, datos de la pareja y cálculo de la experiencia sin contar dos veces los periodos que se cruzan.
- **Envío** con consentimiento, firma, código de postulación único y correo de confirmación (simulado).

## Tecnologías

| Capa | Tecnología |
|---|---|
| Frontend | React + TypeScript, Vite, Tailwind CSS, react-router-dom |
| Backend | Go (Fiber), capas Handler → Service → Repository con inyección de dependencias |
| Base de datos | PostgreSQL 16 |
| Caché y bloqueos | Redis |
| Autenticación | JWT |
| Infraestructura | Docker Compose (PostgreSQL y Redis) |

## Prerrequisitos

| Herramienta | Versión con la que se probó |
|---|---|
| Git | cualquier versión reciente |
| Go | 1.27.1 |
| Node.js (incluye npm) | 24.18.0 |
| Docker Desktop con Docker Compose | cualquier versión reciente |
| PostgreSQL | 16 (imagen `postgres:16`, se levanta con Docker Compose) |
| Redis | imagen `redis:alpine` (se levanta con Docker Compose) |

No hace falta instalar PostgreSQL ni Redis en la máquina: los levanta `docker-compose.yml`.

## Estructura

## Cómo ejecutarlo

```bash
git clone https://github.com/Kerlly-Mite/Examen_Pr-ctico_Sistema_de_Registro.git
cd Examen_Pr-ctico_Sistema_de_Registro
git checkout develop
copy .env.example .env        # en Mac/Linux: cp .env.example .env
```

Edita `.env` y cambia `JWT_SECRET` por un texto propio. Luego:

```bash
docker compose up -d          # PostgreSQL y Redis
cd backend
go mod tidy
go run .                      # API en http://localhost:3001
```

En otra terminal:

```bash
cd frontend
npm install
npm run dev                   # http://localhost:5180
```

| Servicio | Puerto |
|---|---|
| API | 3001 |
| Frontend | 5180 |
| PostgreSQL | 5433 |
| Redis | 6380 |

Vite reenvía las llamadas a `/api` hacia el backend.

## Base de datos

Scripts SQL del repositorio (esquema DDL y datos semilla de la división política):
- `backend/database/01_schema.sql`
- `backend/database/02_seed_catalogos.sql`
- `backend/database/03_seed_simulados.sql`

Si cambia el esquema, hay que recrear la base: `docker compose down -v` y luego `docker compose up -d`.

## Pruebas

Con el backend corriendo, desde `backend/`:

```bash
node pruebas/probar-fase2-3.mjs    # registro, login, bloqueo con Redis y caché
node pruebas/probar-fase4.mjs      # guardado transaccional de la postulación
node pruebas/probar-fase5.mjs      # reglas de negocio y envío
go test ./services/                # reglas de negocio sin base de datos
```

Para repetir las pruebas sin esperar los 10 minutos del bloqueo:

```bash
docker exec biess_redis redis-cli FLUSHALL
docker exec biess_postgres psql -U biess_user -d biess_postulacion -c "DELETE FROM postulantes"
```

Usuarios de prueba (contraseña `Clave1234`, el correo puede ser cualquiera):

| Cédula | Código dactilar | Para qué sirve |
|---|---|---|
| 1702345677 | 1000000002 | Carlos: flujo completo |
| 2403456789 | 1000000003 | Lucía, jubilada: cargo de Jubilados y prueba de inhabilidades |
| 1705678900 | 1000000005 | Extranjero: la postulación se cierra por nacionalidad |
| 0901234567 | 1000000001 | Prueba del bloqueo por intentos fallidos |

## Decisiones técnicas

- **Transacción**: guardar la postulación hace `BEGIN`, `SELECT ... FOR UPDATE`, `INSERT` o `UPDATE` de la cabecera, `DELETE` e `INSERT` de las seis tablas hijas y `COMMIT`. Si algo falla, `ROLLBACK` y no queda nada a medias.
- **Bloqueo optimista**: al enviar se compara `actualizado_en`; si hubo otro cambio, responde 409.
- **Redis**: contadores de intentos y bloqueo (10 min), tokens de registro (15 min) y de recuperación (30 min), y caché de catálogos (24 h).
- **Frontend**: la sesión vive en un contexto y se valida con `GET /api/auth/me` al recargar. El cliente de API convierte los errores `{error, message}` en un `ApiError` y cierra la sesión ante un 401 `no_autorizado`. El formulario completo vive en otro contexto y se envía entero en cada guardado.
- **Errores**: el frontend decide con el código estable `error` y muestra el `message`.

## Reparto del trabajo

| Integrante | Fases | Aporte |
|---|---|---|
| A (Kerlly) | 0 a 3 | Estructura, base de datos, registro, login, recuperación, bloqueo con Redis y caché de catálogos |
| B | 4 y 5 | Guardado transaccional, reglas de negocio, verificación del cargo y envío |
| C (Estefany) | 6 y 7 | Frontend: rutas protegidas, autenticación, formulario completo y consumo de la API |
| Las tres | 8 | README, revisión cruzada y sustentación |

## Flujo de trabajo en Git

Ramas permanentes `main` y `develop`; una rama por fase (`feature/...`, `docs/...`); Pull Request hacia `develop` revisado por otra integrante; commits con prefijo `feat:`, `fix:`, `docs:` o `chore:`; nadie hace push directo a `main` ni a `develop`.

## Alcance y limitaciones

- El enunciado del examen evalúa EF-11 y EF-12. Quedan fuera el módulo de administración (EF-13) y el banco de preguntas con evaluación teórica (EF-14).
- El Registro Civil, el Historial Laboral y el envío de correos están **simulados**: usan tablas de prueba y los correos se imprimen en la consola del backend.
- Los documentos PDF se validan en el cliente (formato y tamaño), pero por ahora solo se guarda el nombre del archivo.