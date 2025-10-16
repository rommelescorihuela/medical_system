package middleware

import (
	"github.com/labstack/echo/v4"
)

// AdminMiddleware validates that the authenticated user has admin role
type AdminMiddleware struct{}

// NewAdminMiddleware creates a new admin middleware instance
func NewAdminMiddleware() *AdminMiddleware {
	return &AdminMiddleware{}
}

// RequireAdmin middleware ensures the user has admin role
func (m *AdminMiddleware) RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if user is authenticated (this should be done by JWTMiddleware first)
			userID := c.Get("user_id")
			if userID == nil {
				return c.JSON(401, map[string]string{"error": "Authentication required"})
			}

			// Check if user has admin role
			role := c.Get("role")
			if role == nil {
				return c.JSON(403, map[string]string{"error": "Role information missing"})
			}

			roleStr, ok := role.(string)
			if !ok {
				return c.JSON(403, map[string]string{"error": "Invalid role format"})
			}

			if roleStr != "admin" {
				return c.JSON(403, map[string]string{
					"error":        "Admin access required",
					"current_role": roleStr,
				})
			}

			return next(c)
		}
	}
}

// RequireSuperAdmin middleware for system-wide admin operations
// This could be used for operations that affect multiple tenants
func (m *AdminMiddleware) RequireSuperAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First check if user is admin
			role := c.Get("role")
			if role == nil {
				return c.JSON(403, map[string]string{"error": "Role information missing"})
			}

			roleStr, ok := role.(string)
			if !ok {
				return c.JSON(403, map[string]string{"error": "Invalid role format"})
			}

			// For now, admin role is sufficient. In the future, you might want
			// a separate super_admin role for system-wide operations
			if roleStr != "admin" {
				return c.JSON(403, map[string]string{
					"error":        "Super admin access required",
					"current_role": roleStr,
				})
			}

			return next(c)
		}
	}
}

// RequireTenantAdmin middleware ensures user is admin of the current tenant
func (m *AdminMiddleware) RequireTenantAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check authentication
			userID := c.Get("user_id")
			if userID == nil {
				return c.JSON(401, map[string]string{"error": "Authentication required"})
			}

			// Check role
			role := c.Get("role")
			if role == nil {
				return c.JSON(403, map[string]string{"error": "Role information missing"})
			}

			roleStr, ok := role.(string)
			if !ok {
				return c.JSON(403, map[string]string{"error": "Invalid role format"})
			}

			if roleStr != "admin" {
				return c.JSON(403, map[string]string{
					"error":        "Tenant admin access required",
					"current_role": roleStr,
				})
			}

			// Verify tenant context exists
			tenantID, exists := GetTenantIDFromContext(c)
			if !exists {
				return c.JSON(400, map[string]string{"error": "Tenant context required"})
			}

			// Additional validation could be added here:
			// - Check if user belongs to the tenant
			// - Verify tenant is active
			// - Check specific permissions within the tenant

			c.Set("admin_tenant_id", tenantID)
			return next(c)
		}
	}
}

// Helper functions

// IsAdmin checks if the current user has admin role
func IsAdmin(c echo.Context) bool {
	role := c.Get("role")
	if role == nil {
		return false
	}

	roleStr, ok := role.(string)
	return ok && roleStr == "admin"
}

// GetCurrentUserRole returns the current user's role
func GetCurrentUserRole(c echo.Context) (string, bool) {
	role := c.Get("role")
	if role == nil {
		return "", false
	}

	roleStr, ok := role.(string)
	return roleStr, ok
}

// GetCurrentUserID returns the current user's ID
func GetCurrentUserID(c echo.Context) (string, bool) {
	userID := c.Get("user_id")
	if userID == nil {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}
