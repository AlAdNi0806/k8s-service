// internal/utils/jwt.go
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

func ValidateToken(tokenString string) (int64, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return 0, err
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return 0, fmt.Errorf("token expired")
		}
	}

	if userID, ok := claims["user_id"].(float64); ok {
		return int64(userID), nil
	}
	return 0, fmt.Errorf("invalid user_id in token")
}
