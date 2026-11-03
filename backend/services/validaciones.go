package services

import (
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"

	"biess-backend/models"
)

var reDactilar = regexp.MustCompile(`^[A-Za-z0-9]{10}$`)

// CedulaValida verifica una cédula ecuatoriana: 10 dígitos, provincia 01-24 (o 30), tercer dígito < 6
// y dígito verificador por el algoritmo módulo 10.
func CedulaValida(c string) bool {
	if len(c) != 10 {
		return false
	}
	for _, r := range c {
		if r < '0' || r > '9' {
			return false
		}
	}
	provincia := int(c[0]-'0')*10 + int(c[1]-'0')
	if provincia < 1 || (provincia > 24 && provincia != 30) {
		return false
	}
	if c[2] > '5' {
		return false
	}
	suma := 0
	for i := 0; i < 9; i++ {
		v := int(c[i] - '0')
		if i%2 == 0 {
			v *= 2
			if v > 9 {
				v -= 9
			}
		}
		suma += v
	}
	verificador := (10 - suma%10) % 10
	return verificador == int(c[9]-'0')
}

// parseFecha acepta DD/MM/AAAA (formato del documento funcional) y AAAA-MM-DD (input date del navegador).
func parseFecha(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"02/01/2006", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("fecha inválida")
}

func correoValido(c string) bool {
	if len(c) == 0 || len(c) > 150 {
		return false
	}
	a, err := mail.ParseAddress(c)
	return err == nil && a.Address == c
}

// validarPassword exige 8-72 caracteres (72 es el límite de bcrypt) con al menos una letra y un número.
func validarPassword(p string) *models.AppError {
	if len(p) < 8 || len(p) > 72 {
		return models.ErrSolicitudInvalida("La contraseña debe tener entre 8 y 72 caracteres")
	}
	var letra, numero bool
	for _, r := range p {
		switch {
		case unicode.IsLetter(r):
			letra = true
		case unicode.IsDigit(r):
			numero = true
		}
	}
	if !letra || !numero {
		return models.ErrSolicitudInvalida("La contraseña debe incluir al menos una letra y un número")
	}
	return nil
}
