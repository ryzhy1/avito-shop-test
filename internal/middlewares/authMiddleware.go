package middlewares

import (
	"avito-shop/internal/lib/jwt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	jwtGen *jwt.Generator
}

func NewAuthMiddleware(jwtGen *jwt.Generator) *AuthMiddleware {
	return &AuthMiddleware{
		jwtGen: jwtGen,
	}
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header"})
			return
		}
		tokenString := parts[1]

		id, err := m.jwtGen.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("user_id", id)
		c.Next()
	}
}
