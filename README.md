# Sistema de Registro y Postulación al Directorio del BIESS

Examen práctico del primer parcial — **Frameworks de Programación Web** (UPSE).
Módulo de registro (EF-11) y formulario de postulación (EF-12) del Concurso Público de Méritos y Oposición del Directorio del BIESS, según el Requerimiento Funcional IESS v1.4.

## Arquitectura

```
React + Vite + Tailwind (SPA)  ──HTTP/JSON──▶  API Go (Fiber)  ──▶  PostgreSQL 16
                                                     │
                                                     └──────────▶  Redis
```

Backend en capas con **Inyección de Dependencias** (cada capa recibe la inferior como *interfaz*; `main.go` es el único lugar donde se construyen e inyectan las piezas):

| Capa | Carpeta | Responsabilidad |
|---|---|---|
| Handler | `backend/handlers` | HTTP: lee la petición, llama al service, responde JSON |
| Service | `backend/services` | Reglas de negocio y validaciones |
| Repository | `backend/repositories` | SQL (PostgreSQL) y Redis |
| Models | `backend/models` | Estructuras y errores compartidos |
| Config | `backend/config` | Variables de entorno, pool de BD, cliente Redis |

## Prerrequisitos

| Herramienta | Versión |
|---|---|
| Go | 1.26 o superior |
| Node.js | 22 o superior (con npm) |
| Docker Desktop | para PostgreSQL 16 y Redis |
| Git | cualquier versión reciente |

Si no usas Docker, necesitas PostgreSQL 16 y Redis instalados y ejecutar los scripts de `backend/database` en orden (`01`, `02`, `03`).

## Ejecución local

```bash
# 1. Variables de entorno
cp .env.example .env            # en Windows: copy .env.example .env

# 2. Base de datos y Redis (la primera vez ejecuta los .sql de backend/database)
docker compose up -d

# 3. Backend (http://localhost:3000)
cd backend
go mod tidy                     # solo la primera vez: descarga dependencias y crea go.sum
go run .

# 4. Frontend (http://localhost:5173), en otra terminal
cd frontend
npm install
npm run dev
```

Comprobaciones rápidas:

```bash
curl http://localhost:3000/api/health
curl http://localhost:3000/api/catalogos/paises
curl http://localhost:3000/api/catalogos/paises/1/provincias   # reemplaza 1 por el id de Ecuador
```

Para **reiniciar la base de datos** desde cero: `docker compose down -v && docker compose up -d`.

## Base de datos

Scripts en `backend/database/` (se ejecutan en orden alfabético):

- `01_schema.sql` — DDL: catálogos, cuentas, postulaciones y tablas hijas, con restricciones (`CHECK`, `FOREIGN KEY` compuestas para garantizar la cascada país → provincia → cantón).
- `02_seed_catalogos.sql` — división política del documento funcional (194 países, 24 provincias, 199 cantones), cargos y preguntas de inhabilidades.
- `03_seed_simulados.sql` — personas **ficticias** que reemplazan la interoperabilidad con el Registro Civil y el Historial Laboral del IESS (incluye un extranjero para probar la restricción de nacionalidad).

## Estado del proyecto

| Fase | Contenido | Estado |
|---|---|---|
| 0-1 | Estructura, Docker, SQL, arquitectura en capas, catálogos sin caché | ✅ |
| 2 | Registro, login, JWT y bloqueo de 10 min con Redis (EF-11) | Pendiente |
| 3 | Caché de catálogos en Redis | Pendiente |
| 4-5 | Postulación con transacciones y reglas de negocio (EF-12) | Pendiente |
| 6-7 | Frontend: formularios, rutas protegidas, estados de carga y error | Pendiente |
| 8 | README final, ramas y guion de sustentación | Pendiente |

## Flujo de Git

Ramas: `main` (estable) ← `develop` ← `feature/*`. Commits pequeños con prefijo (`feat:`, `fix:`, `docs:`, `chore:`).
