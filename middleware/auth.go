package middleware

import (
	"medical-system/infrastructure/auth"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	TokenGen auth.TokenGenerator
}

func NewAuthMiddleware(tokenGen auth.TokenGenerator) *AuthMiddleware {
	return &AuthMiddleware{TokenGen: tokenGen}
}

func (m *AuthMiddleware) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(401, map[string]string{"error": "Missing authorization header"})
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(401, map[string]string{"error": "Invalid authorization header format"})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := m.TokenGen.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(401, map[string]string{"error": "Invalid or expired token"})
			}

			// Set user info in context
			if userID, ok := (*claims)["user_id"].(string); ok {
				c.Set("user_id", userID)
			}
			if tenantID, ok := (*claims)["tenant_id"].(string); ok {
				c.Set("tenant_id", tenantID)
			}
			if role, ok := (*claims)["role"].(string); ok {
				c.Set("role", role)
			}

			return next(c)
		}
	}
}

func (m *AuthMiddleware) RBACMiddleware(resource, action string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("user_id")
			if userID == nil {
				return c.JSON(403, map[string]string{"error": "Access denied - not authenticated"})
			}

			// For now, just check if user is authenticated
			// TODO: Implement full RBAC with Casbin
			return next(c)
		}
	}
}
