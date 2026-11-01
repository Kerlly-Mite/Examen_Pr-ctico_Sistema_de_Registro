-- =====================================================================
-- 02_seed_catalogos.sql  |  Datos semilla: división política y catálogos
-- Fuente: Requerimiento Funcional IESS v1.4, EF-12 (1.2.1 a 1.2.3, 1.4, 1.9)
-- =====================================================================
-- Total: 194 países, 24 provincias, 199 cantones

BEGIN;

-- Países (194)
INSERT INTO paises (nombre) VALUES
  ('Afganistán'),
  ('Albania'),
  ('Alemania'),
  ('Andorra'),
  ('Angola'),
  ('Antigua y Barbuda'),
  ('Arabia Saudita'),
  ('Argelia'),
  ('Argentina'),
  ('Armenia'),
  ('Australia'),
  ('Austria'),
  ('Azerbaiyán'),
  ('Bahamas'),
  ('Bangladés'),
  ('Barbados'),
  ('Baréin'),
  ('Bélgica'),
  ('Belice'),
  ('Benín'),
  ('Bielorrusia'),
  ('Birmania'),
  ('Bolivia'),
  ('Bosnia y Herzegovina'),
  ('Botsuana'),
  ('Brasil'),
  ('Brunéi'),
  ('Bulgaria'),
  ('Burkina Faso'),
  ('Burundi'),
  ('Bután'),
  ('Cabo Verde'),
  ('Camboya'),
  ('Camerún'),
  ('Canadá'),
  ('Catar'),
  ('Chad'),
  ('Chile'),
  ('China'),
  ('Chipre'),
  ('Ciudad del Vaticano'),
  ('Colombia'),
  ('Comoras'),
  ('Corea del Norte'),
  ('Corea del Sur'),
  ('Costa de Marfil'),
  ('Costa Rica'),
  ('Croacia'),
  ('Cuba'),
  ('Dinamarca'),
  ('Dominica'),
  ('Ecuador'),
  ('Egipto'),
  ('El Salvador'),
  ('Emiratos Árabes Unidos'),
  ('Eritrea'),
  ('Eslovaquia'),
  ('Eslovenia'),
  ('España'),
  ('Estados Unidos'),
  ('Estonia'),
  ('Etiopía'),
  ('Filipinas'),
  ('Finlandia'),
  ('Fiyi'),
  ('Francia'),
  ('Gabón'),
  ('Gambia'),
  ('Georgia'),
  ('Ghana'),
  ('Granada'),
  ('Grecia'),
  ('Guatemala'),
  ('Guyana'),
  ('Guinea'),
  ('Guinea-Bisáu'),
  ('Guinea Ecuatorial'),
  ('Haití'),
  ('Honduras'),
  ('Hungría'),
  ('India'),
  ('Indonesia'),
  ('Irak'),
  ('Irán'),
  ('Irlanda'),
  ('Islandia'),
  ('Islas Marshall'),
  ('Islas Salomón'),
  ('Israel'),
  ('Italia'),
  ('Jamaica'),
  ('Japón'),
  ('Jordania'),
  ('Kazajistán'),
  ('Kenia'),
  ('Kirguistán'),
  ('Kiribati'),
  ('Kuwait'),
  ('Laos'),
  ('Lesoto'),
  ('Letonia'),
  ('Líbano'),
  ('Liberia'),
  ('Libia'),
  ('Liechtenstein'),
  ('Lituania'),
  ('Luxemburgo'),
  ('Macedonia del Norte'),
  ('Madagascar'),
  ('Malasia'),
  ('Malaui'),
  ('Maldivas'),
  ('Malí'),
  ('Malta'),
  ('Marruecos'),
  ('Mauricio'),
  ('Mauritania'),
  ('México'),
  ('Micronesia'),
  ('Moldavia'),
  ('Mónaco'),
  ('Mongolia'),
  ('Montenegro'),
  ('Mozambique'),
  ('Namibia'),
  ('Nauru'),
  ('Nepal'),
  ('Nicaragua'),
  ('Níger'),
  ('Nigeria'),
  ('Noruega'),
  ('Nueva Zelanda'),
  ('Omán'),
  ('Países Bajos'),
  ('Pakistán'),
  ('Palaos'),
  ('Panamá'),
  ('Papúa Nueva Guinea'),
  ('Paraguay'),
  ('Perú'),
  ('Polonia'),
  ('Portugal'),
  ('Reino Unido'),
  ('República Centroafricana'),
  ('República Checa'),
  ('República del Congo'),
  ('República Democrática del Congo'),
  ('República Dominicana'),
  ('Ruanda'),
  ('Rumanía'),
  ('Rusia'),
  ('Samoa'),
  ('San Cristóbal y Nieves'),
  ('San Marino'),
  ('San Vicente y las Granadinas'),
  ('Santa Lucía'),
  ('Santo Tomé y Príncipe'),
  ('Senegal'),
  ('Serbia'),
  ('Seychelles'),
  ('Sierra Leona'),
  ('Singapur'),
  ('Siria'),
  ('Somalia'),
  ('Sri Lanka'),
  ('Suazilandia'),
  ('Sudáfrica'),
  ('Sudán'),
  ('Sudán del Sur'),
  ('Suecia'),
  ('Suiza'),
  ('Surinam'),
  ('Tailandia'),
  ('Tanzania'),
  ('Tayikistán'),
  ('Timor Oriental'),
  ('Togo'),
  ('Tonga'),
  ('Trinidad y Tobago'),
  ('Túnez'),
  ('Turkmenistán'),
  ('Turquía'),
  ('Tuvalu'),
  ('Ucrania'),
  ('Uganda'),
  ('Uruguay'),
  ('Uzbekistán'),
  ('Vanuatu'),
  ('Venezuela'),
  ('Vietnam'),
  ('Yemen'),
  ('Yibuti'),
  ('Zambia'),
  ('Zimbabue');

-- Provincias de Ecuador (24)
INSERT INTO provincias (pais_id, nombre)
SELECT (SELECT id FROM paises WHERE nombre = 'Ecuador'), v.nombre FROM (VALUES
  ('Azuay'),
  ('Bolívar'),
  ('Cañar'),
  ('Carchi'),
  ('Chimborazo'),
  ('Cotopaxi'),
  ('El Oro'),
  ('Esmeraldas'),
  ('Galápagos'),
  ('Guayas'),
  ('Imbabura'),
  ('Loja'),
  ('Los Ríos'),
  ('Manabí'),
  ('Morona Santiago'),
  ('Napo'),
  ('Orellana'),
  ('Pastaza'),
  ('Pichincha'),
  ('Santa Elena'),
  ('Santo Domingo de los Tsáchilas'),
  ('Sucumbíos'),
  ('Tungurahua'),
  ('Zamora Chinchipe')
) AS v(nombre);

-- Cantones por provincia
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Cuenca'), ('Gualaceo'), ('Chordeleg'), ('Sígsig'), ('Girón'), ('Santa Isabel'), ('Pucará'), ('Camilo Ponce Enríquez')) AS v(nombre) WHERE p.nombre = 'Azuay';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Guaranda'), ('Chillanes'), ('Chimbo'), ('Echeandía'), ('San Miguel'), ('Caluma'), ('Las Naves')) AS v(nombre) WHERE p.nombre = 'Bolívar';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Azogues'), ('Biblián'), ('Cañar'), ('Déleg'), ('El Tambo'), ('Suscal'), ('La Troncal')) AS v(nombre) WHERE p.nombre = 'Cañar';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Tulcán'), ('Bolívar'), ('Espejo'), ('Mira'), ('Montúfar'), ('San Pedro de Huaca')) AS v(nombre) WHERE p.nombre = 'Carchi';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Riobamba'), ('Alausí'), ('Colta'), ('Chunchi'), ('Guamote'), ('Guano'), ('Pallatanga'), ('Penipe'), ('Cumandá')) AS v(nombre) WHERE p.nombre = 'Chimborazo';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Latacunga'), ('La Maná'), ('Pangua'), ('Pujilí'), ('Salcedo'), ('Saquisilí'), ('Sigchos')) AS v(nombre) WHERE p.nombre = 'Cotopaxi';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Machala'), ('Arenillas'), ('Atahualpa'), ('Balsas'), ('Chilla'), ('El Guabo'), ('Huaquillas'), ('Las Lajas'), ('Marcabelí'), ('Pasaje'), ('Piñas'), ('Portovelo'), ('Santa Rosa'), ('Zaruma')) AS v(nombre) WHERE p.nombre = 'El Oro';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Esmeraldas'), ('Atacames'), ('Eloy Alfaro'), ('Muisne'), ('Quinindé'), ('Rioverde'), ('San Lorenzo')) AS v(nombre) WHERE p.nombre = 'Esmeraldas';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('San Cristóbal'), ('Isabela'), ('Santa Cruz')) AS v(nombre) WHERE p.nombre = 'Galápagos';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Guayaquil'), ('Durán'), ('Milagro'), ('Samborondón'), ('Daule'), ('Playas'), ('Balzar'), ('Colimes'), ('El Empalme'), ('El Triunfo'), ('Isidro Ayora'), ('Lomas de Sargentillo'), ('Naranjal'), ('Naranjito'), ('Nobol'), ('Pedro Carbo'), ('Salitre'), ('Simón Bolívar'), ('Yaguachi')) AS v(nombre) WHERE p.nombre = 'Guayas';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Ibarra'), ('Antonio Ante'), ('Cotacachi'), ('Otavalo'), ('Pimampiro'), ('San Miguel de Urcuquí')) AS v(nombre) WHERE p.nombre = 'Imbabura';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Loja'), ('Calvas'), ('Catamayo'), ('Celica'), ('Chaguarpamba'), ('Espíndola'), ('Gonzanamá'), ('Macará'), ('Paltas'), ('Puyango'), ('Saraguro'), ('Sozoranga'), ('Zapotillo')) AS v(nombre) WHERE p.nombre = 'Loja';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Babahoyo'), ('Baba'), ('Buena Fe'), ('Mocache'), ('Montalvo'), ('Palenque'), ('Puebloviejo'), ('Quevedo'), ('Urdaneta'), ('Ventanas'), ('Vinces'), ('Valencia')) AS v(nombre) WHERE p.nombre = 'Los Ríos';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Portoviejo'), ('Manta'), ('Chone'), ('Bolívar'), ('El Carmen'), ('Flavio Alfaro'), ('Jipijapa'), ('Junín'), ('Montecristi'), ('Paján'), ('Pichincha'), ('Rocafuerte'), ('Santa Ana'), ('Sucre'), ('Tosagua'), ('24 de Mayo'), ('Pedernales'), ('Olmedo'), ('Puerto López'), ('Jama'), ('Jaramijó'), ('San Vicente')) AS v(nombre) WHERE p.nombre = 'Manabí';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Morona'), ('Gualaquiza'), ('Limón Indanza'), ('Palora'), ('San Juan Bosco'), ('Santiago'), ('Sucúa'), ('Taisha'), ('Tiwintza')) AS v(nombre) WHERE p.nombre = 'Morona Santiago';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Tena'), ('Archidona'), ('Carlos Julio Arosemena Tola'), ('El Chaco'), ('Quijos')) AS v(nombre) WHERE p.nombre = 'Napo';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Francisco de Orellana'), ('Aguarico'), ('La Joya de los Sachas'), ('Loreto')) AS v(nombre) WHERE p.nombre = 'Orellana';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Pastaza'), ('Arajuno'), ('Mera'), ('Santa Clara')) AS v(nombre) WHERE p.nombre = 'Pastaza';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Quito'), ('Cayambe'), ('Mejía'), ('Pedro Moncayo'), ('Pedro Vicente Maldonado'), ('Puerto Quito'), ('Rumiñahui'), ('San Miguel de los Bancos')) AS v(nombre) WHERE p.nombre = 'Pichincha';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Santa Elena'), ('La Libertad')) AS v(nombre) WHERE p.nombre = 'Santa Elena';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('La Concordia'), ('Santo Domingo')) AS v(nombre) WHERE p.nombre = 'Santo Domingo de los Tsáchilas';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Cascales'), ('Cuyabeno'), ('Gonzalo Pizarro'), ('Lago Agrio'), ('Putumayo'), ('Shushufindi'), ('Sucumbíos')) AS v(nombre) WHERE p.nombre = 'Sucumbíos';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Ambato'), ('Baños de Agua Santa'), ('Cevallos'), ('Mocha'), ('Patate'), ('Pelileo'), ('Píllaro'), ('Quero'), ('Tisaleo')) AS v(nombre) WHERE p.nombre = 'Tungurahua';
INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.nombre FROM provincias p, (VALUES ('Centinela del Cóndor'), ('Chinchipe'), ('El Pangui'), ('Nangaritza'), ('Palanda'), ('Paquisha'), ('Yacuambi'), ('Yantzaza'), ('Zamora')) AS v(nombre) WHERE p.nombre = 'Zamora Chinchipe';

-- Cargos a los que se postula (EF-12, 1.4)
INSERT INTO cargos (codigo, nombre, sector) VALUES
  ('AFILIADOS',   'Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS', 'AFILIADOS'),
  ('JUBILADOS',   'Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS', 'JUBILADOS'),
  ('EMPLEADORES', 'Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS', 'EMPLEADORES');

-- Preguntas de inhabilidades (EF-12, 1.9). respuesta_habilitante = respuesta con la que SÍ puede continuar
INSERT INTO inhabilidad_preguntas (codigo, texto, respuesta_habilitante) VALUES
  ('1.9.1', '¿Es Ecuatoriano/a y está en pleno goce de los derechos de participación política?', TRUE),
  ('1.9.2', '¿Se encuentra inhabilitado para ejercer el comercio?', FALSE),
  ('1.9.3', '¿Está en mora, directa o indirectamente, en el pago de sus obligaciones en cualquiera de las instituciones del Estado o de las instituciones bajo control de la Superintendencia de Bancos, incluidas las off-shore?', FALSE),
  ('1.9.4', '¿Mantiene algún vínculo, por propiedad o administración, con las instituciones del sistema financiero y del sistema de seguros privados, con las bolsas de valores del país, Casas de Valores y Administradoras de Fondos y Fideicomisos, de acuerdo con las normas que la Junta de Política y Regulación Monetaria y Financiera emita para el efecto?', FALSE),
  ('1.9.5', '¿Es funcionario o empleado del IESS - BIESS?', FALSE),
  ('1.9.6', '¿Mantiene interés propio o representa a terceros en la propiedad, la dirección o la gestión de las compañías aseguradoras u otras personas que integran el sistema nacional de seguridad social?', FALSE),
  ('1.9.7', '¿En los últimos 5 años, ha sido removido por algún Organismo de Control?', FALSE),
  ('1.9.8', '¿Mantiene obligaciones pendientes con el SRI?', FALSE),
  ('1.9.9', '¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de cualquier institución financiera?', FALSE),
  ('1.9.10', '¿Actualmente se encuentra litigando en contra del IESS- BIESS?', FALSE),
  ('1.9.11', '¿Se encuentra procesado, enjuiciado o condenado por algún delito relacionado con corrupción o crimen organizado?', FALSE),
  ('1.9.12', '¿Mantiene en firme algún tipo de responsabilidades por la Contraloría General del Estado?', FALSE),
  ('1.9.13', '¿Se encuentra registrado en la base de datos de la Unidad de Análisis Financiero y Económico?', FALSE);

COMMIT;
