package repositories

import "medical-system/domain/entities"

type TenantRepository interface {
	Create(tenant *entities.Tenant) error
	FindByID(id string) (*entities.Tenant, error)
	FindBySlug(slug string) (*entities.Tenant, error)
	FindByEmail(email string) (*entities.Tenant, error)
	Update(tenant *entities.Tenant) error
	Delete(id string) error
	ListActive() ([]*entities.Tenant, error)
	CountUsersByTenant(tenantID string) (int64, error)
}

type TenantSettingsRepository interface {
	Create(settings *entities.TenantSettings) error
	FindByTenantID(tenantID string) (*entities.TenantSettings, error)
	Update(settings *entities.TenantSettings) error
	Delete(tenantID string) error
}
