package handler

import (
	"LAB1/internal/app/ds"
	"LAB1/internal/app/role"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WithAuthCheck возвращает middleware, которое проверяет роль пользователя
func (h *Handler) WithAuthCheck(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claimsVal, exists := ctx.Get("claims")
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "claims not found"})
			return
		}

		claims, ok := claimsVal.(*ds.JWTClaims)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid claims"})
			return
		}

		// Проверяем, есть ли роль пользователя в списке разрешённых
		for _, r := range allowedRoles {
			if claims.Role == r {
				ctx.Next() // роль разрешена, продолжаем выполнение
				return
			}
		}

		// если роль не найдена в allowedRoles
		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
	}
}
