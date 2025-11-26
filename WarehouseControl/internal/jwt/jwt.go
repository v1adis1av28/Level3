package jwt

import (
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/v1adis1av28/Level3/WarehouseControl/internal/models"
)

func GenerateToken(secret string, req *models.LoginRequest) (string, error) {
	claims := jwt.MapClaims{
		"role":     strings.ToLower(req.Role),
		"username": strings.ToLower(req.Username),
		"exp":      time.Now().Add(time.Hour * 4).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ExtractPayloadFromClaims(tokenStr string, jwtSecretKey string) (*models.Payload, error) {
	var payload models.Payload

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, nil
		}
		return []byte(jwtSecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil
	}

	payload.Username = claims["username"].(string)
	payload.Role = claims["role"].(string)

	return &payload, nil
}
