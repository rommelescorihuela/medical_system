package routes

import (
	"medical-system/application/auth"
	"medical-system/container"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(e *echo.Echo, container *container.Container) {
	authService, err := container.GetAuthService()
	if err != nil {
		panic("Failed to get auth service: " + err.Error())
	}

	handler := NewAuthHandler(authService)

	// Public routes
	e.POST("/api/auth/login", handler.Login)
	e.POST("/api/auth/register", handler.Register)
}

type AuthHandler struct {
	authService *auth.AuthApplicationService
}

func NewAuthHandler(authService *auth.AuthApplicationService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req auth.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}

	response, err := h.authService.Login(req)
	if err != nil {
		return c.JSON(401, map[string]string{"error": "Invalid credentials"})
	}

	return c.JSON(200, response)
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req auth.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request"})
	}

	response, err := h.authService.Register(req)
	if err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}

	return c.JSON(201, response)
}
