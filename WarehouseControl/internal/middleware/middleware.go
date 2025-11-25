package middleware

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/ginext"
)

func AuthMiddleware(secret string) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, ginext.H{"error": "authorization header missing"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, ginext.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, ginext.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}
