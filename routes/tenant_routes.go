package routes

import (
	"medical-system/application/tenants"
	"medical-system/container"
	infraauth "medical-system/infrastructure/auth"
	"medical-system/internal/validation"
	authmiddleware "medical-system/middleware"

	"github.com/labstack/echo/v4"
)

func SetupTenantRoutes(e *echo.Echo, container *container.Container) {
	tenantService, err := container.GetTenantService()
	if err != nil {
		panic("Failed to get tenant service: " + err.Error())
	}

	tokenGen, err := container.GetTokenGen()
	if err != nil {
		panic("Failed to get token generator: " + err.Error())
	}

	handler := NewTenantHandler(tenantService)

	// Initialize admin middleware
	adminMiddleware := authmiddleware.NewAdminMiddleware()

	// Get RBAC enforcer for admin auth middleware
	var rbacEnforcer *infraauth.CasbinEnforcer
	container.DigContainer().Invoke(func(rbac *infraauth.CasbinEnforcer) {
		rbacEnforcer = rbac
	})

	adminAuthMiddleware := authmiddleware.NewAuthMiddleware(tokenGen, nil, rbacEnforcer)

	// Public routes for tenant registration
	e.POST("/api/tenants/register", handler.RegisterTenant)

	// Admin-only routes for tenant management
	admin := e.Group("/api/admin/tenants")
	admin.Use(adminAuthMiddleware.JWTMiddleware())
	admin.Use(adminMiddleware.RequireAdmin())

	// Tenant management routes
	admin.GET("", handler.ListTenants)
	admin.GET("/:id/settings", handler.GetTenantSettings)
	admin.PUT("/:id/settings", handler.UpdateTenantSettings)
	admin.DELETE("/:id", handler.DeleteTenant)
	admin.PUT("/:id/status", handler.UpdateTenantStatus)
}

type TenantHandler struct {
	tenantService *tenants.TenantApplicationService
}

func NewTenantHandler(tenantService *tenants.TenantApplicationService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

func (h *TenantHandler) RegisterTenant(c echo.Context) error {
	var req tenants.RegisterTenantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}

	// Validate request
	if validationErrors := validation.ValidateStructWithTranslation(req); validationErrors != nil {
		return c.JSON(400, map[string]interface{}{
			"error":   "Validation failed",
			"details": validationErrors,
		})
	}

	response, err := h.tenantService.RegisterTenant(req)
	if err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	return c.JSON(201, response)
}

// Admin handlers for tenant management
func (h *TenantHandler) ListTenants(c echo.Context) error {
	tenants, err := h.tenantService.ListActiveTenants()
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to list tenants"})
	}

	return c.JSON(200, tenants)
}

func (h *TenantHandler) GetTenantSettings(c echo.Context) error {
	tenantID := c.Param("id")

	settings, err := h.tenantService.GetTenantSettings(tenantID)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "Tenant settings not found"})
	}

	return c.JSON(200, settings)
}

func (h *TenantHandler) UpdateTenantSettings(c echo.Context) error {
	tenantID := c.Param("id")

	var req struct {
		AllowUserRegistration bool   `json:"allow_user_registration"`
		MaxUsers              int    `json:"max_users"`
		Timezone              string `json:"timezone"`
		Language              string `json:"language"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}

	err := h.tenantService.UpdateTenantSettings(
		tenantID,
		req.AllowUserRegistration,
		req.MaxUsers,
		req.Timezone,
		req.Language,
	)
	if err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "Settings updated successfully"})
}

func (h *TenantHandler) DeleteTenant(c echo.Context) error {
	tenantID := c.Param("id")

	err := h.tenantService.DeleteTenant(tenantID)
	if err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "Tenant deleted successfully"})
}

func (h *TenantHandler) UpdateTenantStatus(c echo.Context) error {
	tenantID := c.Param("id")

	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}

	tenant, err := h.tenantService.GetTenantByID(tenantID)
	if err != nil {
		return c.JSON(404, map[string]string{"error": "Tenant not found"})
	}

	tenant.IsActive = req.IsActive
	err = h.tenantService.UpdateTenant(tenant)
	if err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	return c.JSON(200, map[string]string{"message": "Tenant status updated successfully"})
}
