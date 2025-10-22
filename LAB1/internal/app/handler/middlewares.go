package handler

import (
	"LAB1/internal/app/ds"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

const jwtPrefix = "Bearer "

func (h *Handler) JWTMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")

		// 1. Если заголовка вообще нет - 401
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		// 2. Если формат неправильный - 401
		if !strings.HasPrefix(authHeader, jwtPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header format invalid"})
			return
		}

		jwtStr := authHeader[len(jwtPrefix):]

		// 3. Проверяем Redis блеклист - 401
		err := h.Redis.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			return
		} else if err != redis.Nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
			return
		}

		// 4. Проверяем JWT - 401 если невалидный
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(h.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// 5. Добавляем логирование пользователя
		user, err := h.Repository.GetUserByUUID(claims.UserUUID)
		if err == nil {
			// Логируем в консоль для демонстрации
			println("=== JWT MIDDLEWARE ===")
			println("User:", user.Login)
			println("Role:", user.Role)
			println("UUID:", user.UUID.String())
			println("=====================")
		}

		ctx.Set("claims", claims)
		ctx.Set("user_uuid", claims.UserUUID)
		ctx.Set("user_role", claims.Role)
		ctx.Next()
	}
}
