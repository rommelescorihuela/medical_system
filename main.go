package main

import (
	"log"
	"medical-system/container"
	authmiddleware "medical-system/middleware"
	"medical-system/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize dependency container
	container := container.NewContainer()

	// Initialize Echo server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup routes
	routes.SetupAuthRoutes(e, container)

	// Initialize auth middleware
	var authMiddleware *authmiddleware.AuthMiddleware
	container.DigContainer().Invoke(func(am *authmiddleware.AuthMiddleware) {
		authMiddleware = am
	})

	// Protected routes with JWT
	api := e.Group("/api/protected")
	api.Use(authMiddleware.JWTMiddleware())
	api.GET("/profile", func(c echo.Context) error {
		userID := c.Get("user_id").(string)
		role := c.Get("role").(string)
		tenantID := c.Get("tenant_id").(string)
		return c.JSON(200, map[string]interface{}{
			"message":   "Profile accessed successfully with JWT",
			"user_id":   userID,
			"role":      role,
			"tenant_id": tenantID,
		})
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Start server
	log.Println("🚀 Server starting on port 8080")
	if err := e.Start(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
