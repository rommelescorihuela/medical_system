package middleware

import (
	"medical-system/application/tenants"
	"medical-system/domain/entities"
	"strings"

	"github.com/labstack/echo/v4"
)

type TenantMiddleware struct {
	tenantService *tenants.TenantApplicationService
}

func NewTenantMiddleware(tenantService *tenants.TenantApplicationService) *TenantMiddleware {
	return &TenantMiddleware{
		tenantService: tenantService,
	}
}

// TenantIdentifier middleware extracts tenant information from various sources
func (m *TenantMiddleware) TenantIdentifier() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := m.extractTenantID(c)

			if tenantID != "" {
				// Set tenant ID in context for later use
				c.Set("tenant_id", tenantID)

				// Try to load tenant information if available
				if tenant, err := m.tenantService.GetTenantBySlug(tenantID); err == nil {
					c.Set("tenant", tenant)
				}
			}

			return next(c)
		}
	}
}

// extractTenantID extracts tenant identifier from various sources
func (m *TenantMiddleware) extractTenantID(c echo.Context) string {
	// Priority order for tenant identification:
	// 1. JWT token claims (already set by auth middleware)
	// 2. Subdomain (e.g., clinic1.medical-system.com -> clinic1)
	// 3. Header (X-Tenant-ID)
	// 4. Query parameter (tenant_id)
	// 5. Path parameter (tenant_slug)

	// 1. From JWT claims (if auth middleware already ran)
	if tenantID := c.Get("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(string); ok && tid != "" {
			return tid
		}
	}

	// 2. From subdomain
	if tenantID := m.extractFromSubdomain(c); tenantID != "" {
		return tenantID
	}

	// 3. From header
	if tenantID := c.Request().Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}

	// 4. From query parameter
	if tenantID := c.QueryParam("tenant_id"); tenantID != "" {
		return tenantID
	}

	// 5. From path parameter (for tenant-specific routes)
	if tenantID := c.Param("tenant_slug"); tenantID != "" {
		return tenantID
	}

	return ""
}

// extractFromSubdomain extracts tenant ID from subdomain
func (m *TenantMiddleware) extractFromSubdomain(c echo.Context) string {
	host := c.Request().Host

	// Remove port if present
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Split by dots
	parts := strings.Split(host, ".")

	// If we have more than 2 parts, the first part might be the subdomain
	if len(parts) > 2 {
		subdomain := parts[0]

		// Skip common subdomains like www, api, etc.
		commonSubdomains := map[string]bool{
			"www":     true,
			"api":     true,
			"app":     true,
			"admin":   true,
			"dev":     true,
			"staging": true,
			"test":    true,
		}

		if !commonSubdomains[subdomain] {
			return subdomain
		}
	}

	return ""
}

// TenantValidator middleware ensures tenant exists and is active
func (m *TenantMiddleware) TenantValidator() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenantID := c.Get("tenant_id")
			if tenantID == nil {
				return c.JSON(400, map[string]string{"error": "Tenant ID is required"})
			}

			tid, ok := tenantID.(string)
			if !ok || tid == "" {
				return c.JSON(400, map[string]string{"error": "Invalid tenant ID"})
			}

			// Check if tenant exists and is active
			tenant, err := m.tenantService.GetTenantBySlug(tid)
			if err != nil {
				return c.JSON(404, map[string]string{"error": "Tenant not found"})
			}

			if !tenant.IsActive {
				return c.JSON(403, map[string]string{"error": "Tenant is not active"})
			}

			// Set validated tenant in context
			c.Set("tenant", tenant)

			return next(c)
		}
	}
}

// GetTenantFromContext helper function to get tenant from echo context
func GetTenantFromContext(c echo.Context) (*entities.Tenant, bool) {
	tenant := c.Get("tenant")
	if tenant == nil {
		return nil, false
	}

	t, ok := tenant.(*entities.Tenant)
	return t, ok
}

// GetTenantIDFromContext helper function to get tenant ID from echo context
func GetTenantIDFromContext(c echo.Context) (string, bool) {
	tenantID := c.Get("tenant_id")
	if tenantID == nil {
		return "", false
	}

	tid, ok := tenantID.(string)
	return tid, ok
}
