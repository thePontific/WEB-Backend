package ds

import (
	"LAB1/internal/app/role"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	jwt.RegisteredClaims           // вместо StandardClaims
	UserUUID             uuid.UUID `json:"user_uuid"`
	Scopes               []string  `json:"scopes"`
	Role                 role.Role `json:"role"`
}
