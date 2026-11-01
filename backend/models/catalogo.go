// Package models contiene las estructuras de datos compartidas entre capas (sin lógica de negocio).
package models

// Pais es un elemento del catálogo de países.
type Pais struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

// Provincia pertenece a un país (el catálogo cargado corresponde a las provincias de Ecuador).
type Provincia struct {
	ID     int    `json:"id"`
	PaisID int    `json:"pais_id"`
	Nombre string `json:"nombre"`
}

// Canton pertenece a una provincia. En el documento funcional se le llama "Ciudad".
type Canton struct {
	ID          int    `json:"id"`
	ProvinciaID int    `json:"provincia_id"`
	Nombre      string `json:"nombre"`
}
