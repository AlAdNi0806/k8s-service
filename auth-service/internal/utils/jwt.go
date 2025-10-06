package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func InitJWT(secret string) {
	jwtKey = []byte(secret)
}

func GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateToken(tokenString string) (int64, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}

	if userID, ok := claims["user_id"].(float64); ok {
		return int64(userID), nil
	}
	return 0, fmt.Errorf("invalid user_id in token")
}
