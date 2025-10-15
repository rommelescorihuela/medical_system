package auth

import (
	"medical-system/domain/entities"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenGenerator interface {
	GenerateToken(user *entities.User) (string, error)
	ValidateToken(tokenString string) (*jwt.MapClaims, error)
}

type JWTGenerator struct {
	secretKey string
}

func NewJWTGenerator(secretKey string) TokenGenerator {
	return &JWTGenerator{secretKey: secretKey}
}

func (j *JWTGenerator) GenerateToken(user *entities.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"email":     user.Email,
		"role":      user.Role,
		"tenant_id": user.TenantID,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTGenerator) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
