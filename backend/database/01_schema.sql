-- =====================================================================
-- 01_schema.sql  |  DDL del Sistema de Registro y Postulación - Directorio BIESS
-- Base de datos: PostgreSQL 16
-- Cubre EF-11 (registro y autenticación) y EF-12 (formulario de postulación).
-- Se ejecuta automáticamente la PRIMERA vez que se levanta el contenedor (volumen vacío).
-- =====================================================================

BEGIN;

-- ---------------------------------------------------------------------
-- 1. Catálogos de división política (país -> provincia -> cantón)
-- ---------------------------------------------------------------------
CREATE TABLE paises (
    id      SERIAL PRIMARY KEY,
    nombre  VARCHAR(80) NOT NULL UNIQUE
);

CREATE TABLE provincias (
    id       SERIAL PRIMARY KEY,
    pais_id  INT NOT NULL REFERENCES paises(id),
    nombre   VARCHAR(80) NOT NULL,
    UNIQUE (pais_id, nombre),
    UNIQUE (id, pais_id)            -- permite la FK compuesta desde postulaciones
);

CREATE TABLE cantones (
    id            SERIAL PRIMARY KEY,
    provincia_id  INT NOT NULL REFERENCES provincias(id),
    nombre        VARCHAR(80) NOT NULL,
    UNIQUE (provincia_id, nombre),
    UNIQUE (id, provincia_id)       -- permite la FK compuesta desde postulaciones
);

CREATE INDEX idx_provincias_pais ON provincias(pais_id);
CREATE INDEX idx_cantones_provincia ON cantones(provincia_id);

-- ---------------------------------------------------------------------
-- 2. Otros catálogos
-- ---------------------------------------------------------------------
CREATE TABLE cargos (
    id      SERIAL PRIMARY KEY,
    codigo  VARCHAR(20)  NOT NULL UNIQUE,
    nombre  VARCHAR(200) NOT NULL,
    sector  VARCHAR(20)  NOT NULL CHECK (sector IN ('AFILIADOS','JUBILADOS','EMPLEADORES'))
);

CREATE TABLE inhabilidad_preguntas (
    id                   SERIAL PRIMARY KEY,
    codigo               VARCHAR(10)  NOT NULL UNIQUE,   -- 1.9.1 ... 1.9.13
    texto                TEXT         NOT NULL,
    respuesta_habilitante BOOLEAN     NOT NULL           -- respuesta con la que SÍ puede continuar
);

-- ---------------------------------------------------------------------
-- 3. Fuentes SIMULADAS (en producción: interoperabilidad con Registro Civil e Historial Laboral IESS)
-- ---------------------------------------------------------------------
CREATE TABLE registro_civil_simulado (
    cedula            CHAR(10) PRIMARY KEY,
    nombres           VARCHAR(100) NOT NULL,
    apellidos         VARCHAR(100) NOT NULL,
    genero            VARCHAR(10)  NOT NULL CHECK (genero IN ('MASCULINO','FEMENINO')),
    fecha_nacimiento  DATE         NOT NULL,
    nacionalidad      VARCHAR(40)  NOT NULL,
    estado_civil      VARCHAR(20)  NOT NULL
                      CHECK (estado_civil IN ('SOLTERO','CASADO','UNION_DE_HECHO','DIVORCIADO','VIUDO')),
    fecha_expedicion  DATE         NOT NULL,
    codigo_dactilar   CHAR(10)     NOT NULL
);

CREATE TABLE historial_laboral_simulado (
    cedula       CHAR(10) PRIMARY KEY REFERENCES registro_civil_simulado(cedula),
    es_afiliado  BOOLEAN NOT NULL DEFAULT FALSE,
    es_jubilado  BOOLEAN NOT NULL DEFAULT FALSE,
    es_empleador BOOLEAN NOT NULL DEFAULT FALSE
);

-- ---------------------------------------------------------------------
-- 4. Cuentas de postulantes (EF-11)
--    Los intentos fallidos y el bloqueo de 10 min NO van aquí: viven en Redis (con TTL).
-- ---------------------------------------------------------------------
CREATE TABLE postulantes (
    id             BIGSERIAL PRIMARY KEY,
    cedula         CHAR(10)     NOT NULL UNIQUE REFERENCES registro_civil_simulado(cedula),
    correo         VARCHAR(150) NOT NULL UNIQUE,
    password_hash  VARCHAR(100) NOT NULL,                 -- bcrypt
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- 5. Postulación (EF-12): cabecera + tablas hijas
-- ---------------------------------------------------------------------
CREATE SEQUENCE postulacion_codigo_seq;

CREATE TABLE postulaciones (
    id              BIGSERIAL PRIMARY KEY,
    -- "El formulario contará con un código de numeración único"
    codigo          VARCHAR(25) NOT NULL UNIQUE
                    DEFAULT ('BIESS-' || to_char(now(), 'YYYY') || '-' || lpad(nextval('postulacion_codigo_seq')::text, 6, '0')),
    postulante_id   BIGINT NOT NULL UNIQUE REFERENCES postulantes(id) ON DELETE CASCADE,
    estado          VARCHAR(10) NOT NULL DEFAULT 'BORRADOR' CHECK (estado IN ('BORRADOR','ENVIADA')),

    -- 1.1 Información personal (viene del Registro Civil)
    nombres         VARCHAR(100) NOT NULL,
    apellidos       VARCHAR(100) NOT NULL,
    genero          VARCHAR(10)  NOT NULL CHECK (genero IN ('MASCULINO','FEMENINO')),
    fecha_nacimiento DATE        NOT NULL,
    -- Restricción estricta: solo ecuatorianos pueden postular (se valida también en el servicio)
    nacionalidad    VARCHAR(40)  NOT NULL CHECK (upper(nacionalidad) = 'ECUATORIANA'),
    estado_civil    VARCHAR(20)  NOT NULL
                    CHECK (estado_civil IN ('SOLTERO','CASADO','UNION_DE_HECHO','DIVORCIADO','VIUDO')),
    pareja_nombres  VARCHAR(200),
    pareja_cedula   CHAR(10),
    CONSTRAINT ck_pareja CHECK (
        estado_civil NOT IN ('CASADO','UNION_DE_HECHO')
        OR (pareja_nombres IS NOT NULL AND pareja_cedula IS NOT NULL)
    ),

    -- 1.2 Información domiciliaria y de contacto
    pais_id         INT,
    provincia_id    INT,
    canton_id       INT,
    provincia_texto VARCHAR(80),          -- si el país NO es Ecuador, se ingresa manualmente
    ciudad_texto    VARCHAR(80),
    telefono_celular  VARCHAR(15),
    correo_1        VARCHAR(150),
    correo_2        VARCHAR(150),
    calle_principal VARCHAR(120),
    numero          VARCHAR(20),
    calle_secundaria VARCHAR(120),
    sector          VARCHAR(120),
    referencia      VARCHAR(250),
    telefono_domicilio VARCHAR(15),
    -- Cascada garantizada por la BD: la provincia pertenece al país y el cantón a la provincia
    CONSTRAINT fk_post_provincia FOREIGN KEY (provincia_id, pais_id) REFERENCES provincias (id, pais_id),
    CONSTRAINT fk_post_canton    FOREIGN KEY (canton_id, provincia_id) REFERENCES cantones (id, provincia_id),
    CONSTRAINT fk_post_pais      FOREIGN KEY (pais_id) REFERENCES paises (id),

    -- 1.3 Información adicional
    tiene_familiares_iess BOOLEAN NOT NULL DEFAULT FALSE,

    -- 1.4 Cargo al que postula y respaldo de organización
    cargo_id        INT REFERENCES cargos(id),
    respaldado_organizacion BOOLEAN,
    organizacion_nombre     VARCHAR(200),
    organizacion_documento  VARCHAR(255),      -- ruta/nombre del PDF (máx. 4 MB)
    CONSTRAINT ck_organizacion CHECK (
        respaldado_organizacion IS DISTINCT FROM TRUE
        OR (organizacion_nombre IS NOT NULL AND organizacion_documento IS NOT NULL)
    ),

    -- 1.10 Consentimiento y envío
    consentimiento_aceptado BOOLEAN NOT NULL DEFAULT FALSE,
    enviada_en      TIMESTAMPTZ,
    creado_en       TIMESTAMPTZ NOT NULL DEFAULT now(),
    actualizado_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_enviada CHECK (estado = 'BORRADOR' OR (consentimiento_aceptado AND enviada_en IS NOT NULL))
);

CREATE INDEX idx_postulaciones_cargo ON postulaciones(cargo_id);

-- 1.3.1 Familiares que trabajan en el IESS / BIESS
CREATE TABLE familiares_iess (
    id               BIGSERIAL PRIMARY KEY,
    postulacion_id   BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    nombres_apellidos VARCHAR(200) NOT NULL,
    parentesco       VARCHAR(20) NOT NULL CHECK (parentesco IN (
        'CONVIVIENTE','CONYUGE','CUNADO','HERMANO','HIJO_POLITICO','HIJO','MADRE','NIETO','NUERA',
        'PADRE','PRIMO_HERMANO','SOBRINO','SUEGRO','TIO','YERNO')),
    institucion      VARCHAR(5)  NOT NULL CHECK (institucion IN ('IESS','BIESS')),
    area             VARCHAR(120) NOT NULL,
    cargo            VARCHAR(120) NOT NULL
);

-- 1.5 Formación académica
CREATE TABLE formacion_academica (
    id                BIGSERIAL PRIMARY KEY,
    postulacion_id    BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    nivel_instruccion VARCHAR(30) NOT NULL CHECK (nivel_instruccion IN (
        'TERCER_NIVEL','CUARTO_ESPECIALIDAD','CUARTO_MAESTRIA','CUARTO_PHD')),
    institucion       VARCHAR(200) NOT NULL,
    tipo              VARCHAR(10)  NOT NULL CHECK (tipo IN ('NACIONAL','EXTRANJERA')),
    titulo            VARCHAR(200) NOT NULL,
    area_estudios     VARCHAR(40)  NOT NULL CHECK (area_estudios IN (
        'FINANZAS','MERCADO_CAPITALES','ADMINISTRACION','ECONOMIA','DERECHO','AFINES')),
    registro_senescyt VARCHAR(40)  NOT NULL,
    fecha_registro_senescyt DATE   NOT NULL,
    documento_respaldo VARCHAR(255) NOT NULL         -- PDF, máx. 2 MB
);

-- 1.6 Experiencia profesional
CREATE TABLE experiencia_profesional (
    id                  BIGSERIAL PRIMARY KEY,
    postulacion_id      BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    tipo                VARCHAR(10) NOT NULL CHECK (tipo IN ('GENERAL','DIRECCION')),
    cargo               VARCHAR(150) NOT NULL,
    institucion         VARCHAR(200) NOT NULL,
    fecha_inicio        DATE NOT NULL,
    fecha_fin           DATE NOT NULL,
    actividades         VARCHAR(40) NOT NULL CHECK (actividades IN (
        'FINANZAS','MERCADO_CAPITALES','ADMINISTRACION','ECONOMIA','DERECHO','AFINES')),
    descripcion_funciones VARCHAR(1000) NOT NULL,
    documento_respaldo  VARCHAR(255) NOT NULL,       -- PDF, máx. 2 MB
    CONSTRAINT ck_fechas_experiencia CHECK (fecha_fin >= fecha_inicio)
);

-- 1.7 Publicaciones e investigaciones
CREATE TABLE publicaciones (
    id                 BIGSERIAL PRIMARY KEY,
    postulacion_id     BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    titulo             VARCHAR(250) NOT NULL,
    tipo               VARCHAR(25)  NOT NULL CHECK (tipo IN ('PUBLICACION','TRABAJO_INVESTIGACION')),
    medio_publicacion  VARCHAR(200) NOT NULL,
    fecha_publicacion  DATE NOT NULL,
    relacionada_formacion BOOLEAN NOT NULL,
    documento_respaldo VARCHAR(255) NOT NULL
);

-- 1.8 Capacitación
CREATE TABLE capacitaciones (
    id                 BIGSERIAL PRIMARY KEY,
    postulacion_id     BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    tipo_evento        VARCHAR(10) NOT NULL CHECK (tipo_evento IN ('CURSO','SEMINARIO','TALLER')),
    nombre_evento      VARCHAR(200) NOT NULL,
    institucion        VARCHAR(200) NOT NULL,
    fecha_inicio       DATE NOT NULL,
    fecha_fin          DATE NOT NULL,
    horas              INT  NOT NULL CHECK (horas > 0),
    documento_respaldo VARCHAR(255) NOT NULL,
    CONSTRAINT ck_fechas_capacitacion CHECK (fecha_fin >= fecha_inicio)
);

-- 1.9 Respuestas de inhabilidades (una fila por pregunta)
CREATE TABLE postulacion_inhabilidades (
    postulacion_id  BIGINT NOT NULL REFERENCES postulaciones(id) ON DELETE CASCADE,
    pregunta_id     INT    NOT NULL REFERENCES inhabilidad_preguntas(id),
    respuesta       BOOLEAN NOT NULL,
    -- Solo para la pregunta 1.9.7 (removido por organismo de control)
    remocion_institucion VARCHAR(200),
    remocion_cargo       VARCHAR(120),
    remocion_fecha_fin   DATE,
    remocion_motivo      VARCHAR(250),
    PRIMARY KEY (postulacion_id, pregunta_id)
);

CREATE INDEX idx_familiares_postulacion  ON familiares_iess(postulacion_id);
CREATE INDEX idx_formacion_postulacion   ON formacion_academica(postulacion_id);
CREATE INDEX idx_experiencia_postulacion ON experiencia_profesional(postulacion_id);
CREATE INDEX idx_publicaciones_postulacion ON publicaciones(postulacion_id);
CREATE INDEX idx_capacitaciones_postulacion ON capacitaciones(postulacion_id);

COMMIT;
