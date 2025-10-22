package handler

import (
	"LAB1/internal/app/ds"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtPrefix = "Bearer "

func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, jwtPrefix) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "authorization header missing or invalid"})
			return
		}

		tokenStr := authHeader[len(jwtPrefix):]

		token, err := jwt.ParseWithClaims(tokenStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid token claims"})
			return
		}

		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
