-- =====================================================================
-- 03_seed_simulados.sql  |  Datos SIMULADOS (personas ficticias) para pruebas
-- Sustituyen la interoperabilidad con el Registro Civil y el Historial Laboral del IESS.
-- Todas las cédulas cumplen el algoritmo de validación (módulo 10); los nombres son inventados.
-- Casos de prueba:  1 = afiliada | 2 = empleador (NO afiliado) | 3 = jubilada |
--                   4 = afiliado | 5 = extranjero (debe cerrarse la postulación)
-- =====================================================================

BEGIN;

INSERT INTO registro_civil_simulado
  (cedula, nombres, apellidos, genero, fecha_nacimiento, nacionalidad, estado_civil, fecha_expedicion, codigo_dactilar) VALUES
  ('0901234567', 'MARIA FERNANDA', 'TORRES LEON', 'FEMENINO', '1985-03-14', 'ECUATORIANA', 'SOLTERO', '2018-06-21', '1000000001'),
  ('1702345677', 'CARLOS ANDRES', 'PAREDES VERA', 'MASCULINO', '1978-11-02', 'ECUATORIANA', 'CASADO', '2019-02-11', '1000000002'),
  ('2403456789', 'LUCIA ELENA', 'MORA CEDEÑO', 'FEMENINO', '1955-07-30', 'ECUATORIANA', 'VIUDO', '2017-09-05', '1000000003'),
  ('0104567896', 'JOSE LUIS', 'RAMIREZ SOLIS', 'MASCULINO', '1990-01-25', 'ECUATORIANA', 'UNION_DE_HECHO', '2020-12-15', '1000000004'),
  ('1705678900', 'JOHN MICHAEL', 'SMITH BROWN', 'MASCULINO', '1982-05-09', 'ESTADOUNIDENSE', 'SOLTERO', '2021-04-19', '1000000005');

INSERT INTO historial_laboral_simulado (cedula, es_afiliado, es_jubilado, es_empleador) VALUES
  ('0901234567', TRUE, FALSE, FALSE),
  ('1702345677', FALSE, FALSE, TRUE),
  ('2403456789', FALSE, TRUE, FALSE),
  ('0104567896', TRUE, FALSE, FALSE),
  ('1705678900', FALSE, FALSE, FALSE);

-- Personas ficticias adicionales (cedulas validas por modulo 10, dactilar alfanumerico)
-- 6 = afiliado | 7 = afiliada | 8 = jubilado | 9 = empleador (NO afiliado) | 10 = afiliada
INSERT INTO registro_civil_simulado
  (cedula, nombres, apellidos, genero, fecha_nacimiento, nacionalidad, estado_civil, fecha_expedicion, codigo_dactilar) VALUES
  ('0912345675', 'JORGE LUIS', 'PRUEBA MORA', 'MASCULINO', '1988-04-17', 'ECUATORIANA', 'CASADO', '2019-08-12', 'A1B2C3D4E5'),
  ('1709876542', 'ANA PATRICIA', 'PRUEBA VERA', 'FEMENINO', '1992-09-03', 'ECUATORIANA', 'SOLTERO', '2021-01-20', 'E3333I1222'),
  ('0102030400', 'LUIS DIEGO', 'PRUEBA CUEVA', 'MASCULINO', '1952-12-11', 'ECUATORIANA', 'VIUDO', '2016-05-30', 'V1234V5678'),
  ('2401234568', 'DIEGO ANDRES', 'PRUEBA SOTO', 'MASCULINO', '1975-02-28', 'ECUATORIANA', 'UNION_DE_HECHO', '2018-10-09', 'K9H8G7F6D5'),
  ('1304567892', 'MARIA JOSE', 'PRUEBA LEON', 'FEMENINO', '1996-06-22', 'ECUATORIANA', 'SOLTERO', '2022-03-15', 'M7N6B5V4C3');

INSERT INTO historial_laboral_simulado (cedula, es_afiliado, es_jubilado, es_empleador) VALUES
  ('0912345675', TRUE, FALSE, FALSE),
  ('1709876542', TRUE, FALSE, FALSE),
  ('0102030400', FALSE, TRUE, FALSE),
  ('2401234568', FALSE, FALSE, TRUE),
  ('1304567892', TRUE, FALSE, FALSE);

COMMIT;
