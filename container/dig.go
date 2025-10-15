package container

import (
	appauth "medical-system/application/auth"
	"medical-system/domain/services"
	infraauth "medical-system/infrastructure/auth"
	"medical-system/infrastructure/database"
	"medical-system/infrastructure/repositories"
	authmiddleware "medical-system/middleware"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/dig"
)

type Container struct {
	dig *dig.Container
}

func (c *Container) DigContainer() *dig.Container {
	return c.dig
}

func NewContainer() *Container {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		// If .env file doesn't exist, continue without it
		// This allows the app to run with environment variables set elsewhere
	}

	c := &Container{dig: dig.New()}
	c.registerDependencies()
	return c
}

func (c *Container) registerDependencies() {
	// Database
	c.dig.Provide(database.NewConnection)

	// Token Generator
	c.dig.Provide(func() infraauth.TokenGenerator {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "your-super-secret-jwt-key-change-this-in-production" // fallback
		}
		return infraauth.NewJWTGenerator(jwtSecret)
	})

	// Casbin RBAC
	c.dig.Provide(infraauth.NewCasbinEnforcer)

	// Repositories
	c.dig.Provide(repositories.NewUserRepository)

	// Domain Services
	c.dig.Provide(services.NewAuthService)

	// Application Services
	c.dig.Provide(appauth.NewAuthApplicationService)

	// Middleware
	c.dig.Provide(func(tokenGen infraauth.TokenGenerator) *authmiddleware.AuthMiddleware {
		return &authmiddleware.AuthMiddleware{TokenGen: tokenGen}
	})
}

func (c *Container) GetAuthService() (*appauth.AuthApplicationService, error) {
	var service *appauth.AuthApplicationService
	err := c.dig.Invoke(func(s *appauth.AuthApplicationService) {
		service = s
	})
	return service, err
}
