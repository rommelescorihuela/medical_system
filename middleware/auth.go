package middleware

import (
	"medical-system/application/tenants"
	"medical-system/infrastructure/auth"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	TokenGen      auth.TokenGenerator
	tenantService *tenants.TenantApplicationService
	rbacEnforcer  *auth.CasbinEnforcer
}

func NewAuthMiddleware(tokenGen auth.TokenGenerator, tenantService *tenants.TenantApplicationService, rbacEnforcer *auth.CasbinEnforcer) *AuthMiddleware {
	return &AuthMiddleware{
		TokenGen:      tokenGen,
		tenantService: tenantService,
		rbacEnforcer:  rbacEnforcer,
	}
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

				// Load tenant information if available (only if tenantService is not nil)
				if m.tenantService != nil {
					if tenant, err := m.tenantService.GetTenantBySlug(tenantID); err == nil {
						c.Set("tenant", tenant)
					}
				}
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

			tenantID := c.Get("tenant_id")
			if tenantID == nil {
				return c.JSON(403, map[string]string{"error": "Access denied - tenant context required"})
			}

			userIDStr, ok := userID.(string)
			if !ok {
				return c.JSON(403, map[string]string{"error": "Invalid user ID"})
			}

			tenantIDStr, ok := tenantID.(string)
			if !ok {
				return c.JSON(403, map[string]string{"error": "Invalid tenant ID"})
			}

			// Check permission using Casbin
			// Note: This assumes the enforcer is available through dependency injection
			// For now, we'll implement a basic check - this should be enhanced
			allowed, err := m.checkPermission(userIDStr, tenantIDStr, resource, action)
			if err != nil {
				return c.JSON(500, map[string]string{"error": "Permission check failed"})
			}

			if !allowed {
				return c.JSON(403, map[string]string{
					"error":    "Access denied - insufficient permissions",
					"resource": resource,
					"action":   action,
				})
			}

			return next(c)
		}
	}
}

// checkPermission performs RBAC permission check using Casbin
func (m *AuthMiddleware) checkPermission(userID, tenantID, resource, action string) (bool, error) {
	if m.rbacEnforcer == nil {
		// Fallback to basic checks if enforcer is not available
		return m.basicPermissionCheck(userID, tenantID, resource, action)
	}

	// Use Casbin for permission checking
	return m.rbacEnforcer.CheckPermission(userID, tenantID, resource, action)
}

// basicPermissionCheck provides fallback permission checking
func (m *AuthMiddleware) basicPermissionCheck(userID, tenantID, resource, action string) (bool, error) {
	// Get role from context (this should be set during JWT validation)
	role := "user" // Default fallback

	// Basic permission matrix for fallback
	switch role {
	case "admin", "super_admin":
		return true, nil // Admins can do everything
	case "user":
		// Users can only access their own profile
		if resource == "profile" && (action == "read" || action == "write") {
			return true, nil
		}
	}

	return false, nil
}
