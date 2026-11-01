package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims son los datos que viajan firmados dentro del token de sesión.
type Claims struct {
	PostulanteID int64
	Cedula       string
}

// GestorTokens genera y valida los JWT de sesión. Es una interfaz para poder inyectarla
// tanto en el AuthService (genera) como en el middleware (valida).
type GestorTokens interface {
	Generar(postulanteID int64, cedula string) (string, error)
	Validar(token string) (*Claims, error)
}

type jwtClaims struct {
	Cedula string `json:"ced"`
	jwt.RegisteredClaims
}

type gestorJWT struct {
	secreto  []byte
	vigencia time.Duration
}

// NewGestorTokens crea el gestor con HS256. El secreto sale de JWT_SECRET (mínimo 32 caracteres).
func NewGestorTokens(secreto string, vigencia time.Duration) GestorTokens {
	return &gestorJWT{secreto: []byte(secreto), vigencia: vigencia}
}

func (g *gestorJWT) Generar(postulanteID int64, cedula string) (string, error) {
	ahora := time.Now()
	claims := jwtClaims{
		Cedula: cedula,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(postulanteID, 10),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(g.vigencia)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(g.secreto)
	if err != nil {
		return "", fmt.Errorf("firmar JWT: %w", err)
	}
	return token, nil
}

func (g *gestorJWT) Validar(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &jwtClaims{}, func(_ *jwt.Token) (any, error) {
		return g.secreto, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}
	c, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("token inválido")
	}
	id, err := strconv.ParseInt(c.Subject, 10, 64)
	if err != nil {
		return nil, errors.New("token sin identificador válido")
	}
	return &Claims{PostulanteID: id, Cedula: c.Cedula}, nil
}
