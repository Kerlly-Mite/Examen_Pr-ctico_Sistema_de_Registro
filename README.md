# Sistema de Registro y Postulación al Directorio del BIESS

Examen práctico del primer parcial — **Frameworks de Programación Web** (UPSE).
Módulo de registro (EF-11) y formulario de postulación (EF-12) del Concurso Público de Méritos y Oposición del Directorio del BIESS, según el Requerimiento Funcional IESS v1.4 (Resolución C.D. 701). Proyecto desarrollado por tres integrantes.

## Qué hace

- **Registro** en dos pasos: verificación de identidad con cédula y fecha de expedición o código dactilar (simulado), y luego correo y contraseña.
- **Login** con JWT (8 horas), recuperación de contraseña y **bloqueo de 10 minutos** tras 3 intentos fallidos.
- **Formulario de postulación** por secciones: información personal, domicilio con cascada País → Provincia → Cantón, familiares en el IESS/BIESS, cargo, formación, experiencia, publicaciones, capacitación e inhabilidades.
- **Reglas de negocio** en el servidor: solo nacionalidad ecuatoriana, validación de la calidad de afiliado, jubilado o empleador, datos de la pareja y cálculo de la experiencia sin contar dos veces los periodos que se cruzan.
- **Envío** con consentimiento, firma, código de postulación único y correo de confirmación (simulado).

## Arquitectura

~~~
React + Vite + Tailwind (SPA)  --HTTP/JSON-->  API Go (Fiber)  -->  PostgreSQL 16
                                                     |
                                                     +------------>  Redis
~~~

Backend en capas con **inyección de dependencias** (cada capa recibe la inferior como *interfaz*; `main.go` es el único lugar donde se construyen e inyectan las piezas):

| Capa | Carpeta | Responsabilidad |
|---|---|---|
| Handler | `backend/handlers` | HTTP: lee la petición, llama al service, responde JSON |
| Service | `backend/services` | Reglas de negocio y validaciones |
| Repository | `backend/repositories` | SQL (PostgreSQL) y Redis |
| Models | `backend/models` | Estructuras y errores compartidos |
| Config | `backend/config` | Variables de entorno, pool de BD, cliente Redis |

## Tecnologías

| Capa | Tecnología |
|---|---|
| Frontend | React + TypeScript, Vite, Tailwind CSS, react-router-dom |
| Backend | Go (Fiber) |
| Base de datos | PostgreSQL 16 |
| Caché y bloqueos | Redis |
| Autenticación | JWT |
| Infraestructura | Docker Compose (PostgreSQL y Redis) |

## Prerrequisitos

| Herramienta | Versión mínima | Probado con |
|---|---|---|
| Go | 1.26 | 1.27.1 |
| Node.js (con npm) | 22 | 24.18.0 |
| Docker Desktop (Docker Compose) | reciente | — |
| PostgreSQL | 16 | imagen `postgres:16` (Docker Compose) |
| Redis | — | imagen `redis:alpine` (Docker Compose) |
| Git | reciente | — |

No hace falta instalar PostgreSQL ni Redis en la máquina: los levanta `docker-compose.yml`. Si no usas Docker, necesitas PostgreSQL 16 y Redis instalados y ejecutar los scripts de `backend/database` en orden (`01`, `02`, `03`).

## Estructura

~~~
backend/      API en Go (config, database, handlers, middleware, models, repositories, routes, services, pruebas)
frontend/     Aplicación React (src/api, components, context, hooks, lib, pages, types)
docs/API.md   Contrato de la API (rutas, JSON y códigos de error)
docker-compose.yml, .env.example
~~~

## Ejecución local

~~~bash
git clone https://github.com/Kerlly-Mite/Examen_Pr-ctico_Sistema_de_Registro.git
cd Examen_Pr-ctico_Sistema_de_Registro
git checkout develop
copy .env.example .env        # en Mac/Linux: cp .env.example .env
~~~

Edita `.env` y cambia `JWT_SECRET` por un texto propio. Luego:

~~~bash
docker compose up -d          # PostgreSQL y Redis (la primera vez ejecuta los .sql)
cd backend
go mod tidy                   # solo la primera vez
go run .                      # API en http://localhost:3001
~~~

En otra terminal:

~~~bash
cd frontend
npm install
npm run dev                   # http://localhost:5180
~~~

| Servicio | Puerto |
|---|---|
| API | 3001 |
| Frontend | 5180 |
| PostgreSQL | 5433 |
| Redis | 6380 |

Vite reenvía las llamadas a `/api` hacia el backend. Comprobaciones rápidas:

~~~bash
curl http://localhost:3001/api/health
curl http://localhost:3001/api/catalogos/paises
~~~

## Base de datos

Scripts en `backend/database/` (se ejecutan en orden alfabético):

- `01_schema.sql` — DDL: catálogos, cuentas, postulaciones y tablas hijas, con restricciones (`CHECK`, `FOREIGN KEY` compuestas para garantizar la cascada país → provincia → cantón).
- `02_seed_catalogos.sql` — división política del documento funcional (194 países, 24 provincias, 199 cantones), cargos y preguntas de inhabilidades.
- `03_seed_simulados.sql` — personas **ficticias** que reemplazan la interoperabilidad con el Registro Civil y el Historial Laboral del IESS (incluye un extranjero para probar la restricción de nacionalidad).

Para **reiniciar la base de datos** desde cero: `docker compose down -v && docker compose up -d`.

## Pruebas

Con el backend corriendo, desde `backend/`:

~~~bash
node pruebas/probar-fase2-3.mjs    # registro, login, bloqueo con Redis y caché
node pruebas/probar-fase4.mjs      # guardado transaccional de la postulación
node pruebas/probar-fase5.mjs      # reglas de negocio y envío
go test ./services/                # reglas de negocio sin base de datos
~~~

Para repetir las pruebas sin esperar los 10 minutos del bloqueo:

~~~bash
docker exec biess_redis redis-cli FLUSHALL
docker exec biess_postgres psql -U biess_user -d biess_postulacion -c "DELETE FROM postulantes"
~~~

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

## Estado del proyecto y reparto

| Fase | Contenido | Responsable | Estado |
|---|---|---|---|
| 0-1 | Estructura, Docker, SQL, arquitectura en capas, catálogos | A (Kerlly) | Hecha |
| 2 | Registro, login, JWT y bloqueo de 10 min con Redis (EF-11) | A (Kerlly) | Hecha |
| 3 | Caché de catálogos en Redis | A (Kerlly) | Hecha |
| 4-5 | Postulación con transacciones y reglas de negocio (EF-12) | B | Hecha |
| 6-7 | Frontend: autenticación, formulario completo, rutas protegidas, estados de carga y error | C (Estefany) | Hecha |
| 8 | README final, revisión cruzada y guion de sustentación | Las tres | En curso |

## Flujo de Git

Ramas permanentes `main` (estable) y `develop` (integración); una rama por fase (`feature/...`, `docs/...`); Pull Request hacia `develop` revisado por otra integrante; commits con prefijo `feat:`, `fix:`, `docs:` o `chore:`; nadie hace push directo a `main` ni a `develop`.

## Alcance y limitaciones

- El enunciado del examen evalúa EF-11 y EF-12. Quedan fuera el módulo de administración (EF-13) y el banco de preguntas con evaluación teórica (EF-14).
- El Registro Civil, el Historial Laboral y el envío de correos están **simulados**: usan tablas de prueba y los correos se imprimen en la consola del backend.
- Los documentos PDF se validan en el cliente (formato y tamaño), pero por ahora solo se guarda el nombre del archivo.