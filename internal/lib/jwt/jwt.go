package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
)

type Generator struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewGenerator(secret string, accessTTL time.Duration, refreshTTL time.Duration) *Generator {
	return &Generator{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (g *Generator) GeneratePair(id string) (accessToken string, refreshToken string, err error) {
	now := time.Now().Unix()

	jtiAccess := uuid.NewString()
	jtiRefresh := uuid.NewString()

	accessClaims := jwt.MapClaims{
		"sub": id,
		"iat": now,
		"exp": time.Now().Add(g.accessTTL).Unix(),
		"jti": jtiAccess,
		"typ": "access",
	}

	refreshClaims := jwt.MapClaims{
		"sub": id,
		"iat": now,
		"exp": time.Now().Add(g.refreshTTL).Unix(),
		"jti": jtiRefresh,
		"typ": "refresh",
	}

	aToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessClaims)
	accessToken, err = aToken.SignedString(g.secret)
	if err != nil {
		return "", "", err
	}

	rToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshClaims)
	refreshToken, err = rToken.SignedString(g.secret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (g *Generator) ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return g.secret, nil
	})

	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	id, ok := claims["sub"]
	if !ok {
		return "", errors.New("invalid user_id in token")
	}

	return id.(string), nil
}
