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

COMMIT;
