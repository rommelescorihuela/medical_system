package container

import (
	appauth "medical-system/application/auth"
	apptenants "medical-system/application/tenants"
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
	c.dig.Provide(repositories.NewTenantRepository)
	c.dig.Provide(repositories.NewTenantSettingsRepository)

	// Domain Services
	c.dig.Provide(services.NewAuthService)
	c.dig.Provide(services.NewTenantService)
	c.dig.Provide(services.NewTenantValidator)

	// Application Services
	c.dig.Provide(appauth.NewAuthApplicationService)
	c.dig.Provide(apptenants.NewTenantApplicationService)

	// Middleware
	c.dig.Provide(func(tokenGen infraauth.TokenGenerator, tenantService *apptenants.TenantApplicationService, rbacEnforcer *infraauth.CasbinEnforcer) *authmiddleware.AuthMiddleware {
		return authmiddleware.NewAuthMiddleware(tokenGen, tenantService, rbacEnforcer)
	})
	c.dig.Provide(func(tenantService *apptenants.TenantApplicationService) *authmiddleware.TenantMiddleware {
		return authmiddleware.NewTenantMiddleware(tenantService)
	})
	c.dig.Provide(authmiddleware.NewAdminMiddleware)
}

func (c *Container) GetAuthService() (*appauth.AuthApplicationService, error) {
	var service *appauth.AuthApplicationService
	err := c.dig.Invoke(func(s *appauth.AuthApplicationService) {
		service = s
	})
	return service, err
}

func (c *Container) GetTenantService() (*apptenants.TenantApplicationService, error) {
	var service *apptenants.TenantApplicationService
	err := c.dig.Invoke(func(s *apptenants.TenantApplicationService) {
		service = s
	})
	return service, err
}

func (c *Container) GetTokenGen() (infraauth.TokenGenerator, error) {
	var tokenGen infraauth.TokenGenerator
	err := c.dig.Invoke(func(tg infraauth.TokenGenerator) {
		tokenGen = tg
	})
	return tokenGen, err
}
