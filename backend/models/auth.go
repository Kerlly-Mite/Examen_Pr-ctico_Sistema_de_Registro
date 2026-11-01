package models

import "time"

// RegistroCivil representa la fuente SIMULADA de identidad (tabla registro_civil_simulado).
type RegistroCivil struct {
	Cedula          string
	Nombres         string
	Apellidos       string
	Genero          string
	FechaNacimiento time.Time
	Nacionalidad    string
	EstadoCivil     string
	FechaExpedicion time.Time
	CodigoDactilar  string
}

// Postulante es la cuenta de acceso (tabla postulantes). PasswordHash nunca se serializa.
type Postulante struct {
	ID           int64
	Cedula       string
	Correo       string
	PasswordHash string
	CreadoEn     time.Time
}

// PerfilPostulante es lo que se devuelve al frontend sobre la persona autenticada.
type PerfilPostulante struct {
	ID        int64  `json:"id"`
	Cedula    string `json:"cedula"`
	Correo    string `json:"correo"`
	Nombres   string `json:"nombres"`
	Apellidos string `json:"apellidos"`
}

// VerificarIdentidadReq se usa en el registro (paso 1) y en la recuperación de contraseña.
// Se debe enviar fecha_expedicion (DD/MM/AAAA o AAAA-MM-DD) o codigo_dactilar (10 dígitos).
type VerificarIdentidadReq struct {
	Cedula          string `json:"cedula"`
	FechaExpedicion string `json:"fecha_expedicion"`
	CodigoDactilar  string `json:"codigo_dactilar"`
}

type VerificarIdentidadResp struct {
	TokenRegistro    string `json:"token_registro"`
	Nombres          string `json:"nombres"`
	Apellidos        string `json:"apellidos"`
	ExpiraEnSegundos int    `json:"expira_en_segundos"`
}

type RegistroReq struct {
	TokenRegistro string `json:"token_registro"`
	Correo        string `json:"correo"`
	Password      string `json:"password"`
}

type LoginReq struct {
	Cedula   string `json:"cedula"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token      string           `json:"token"`
	Postulante PerfilPostulante `json:"postulante"`
}

type RestablecerReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}
