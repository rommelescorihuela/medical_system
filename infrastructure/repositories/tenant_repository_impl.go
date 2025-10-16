package repositories

import (
	"medical-system/domain/entities"
	"medical-system/domain/repositories"

	"gorm.io/gorm"
)

type TenantRepositoryImpl struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) repositories.TenantRepository {
	return &TenantRepositoryImpl{db: db}
}

func (r *TenantRepositoryImpl) Create(tenant *entities.Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *TenantRepositoryImpl) FindByID(id string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	err := r.db.First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepositoryImpl) FindBySlug(slug string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	err := r.db.Where("slug = ?", slug).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepositoryImpl) FindByEmail(email string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	err := r.db.Where("email = ?", email).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *TenantRepositoryImpl) Update(tenant *entities.Tenant) error {
	return r.db.Save(tenant).Error
}

func (r *TenantRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&entities.Tenant{}, "id = ?", id).Error
}

func (r *TenantRepositoryImpl) ListActive() ([]*entities.Tenant, error) {
	var tenants []*entities.Tenant
	err := r.db.Where("is_active = ?", true).Find(&tenants).Error
	return tenants, err
}

func (r *TenantRepositoryImpl) CountUsersByTenant(tenantID string) (int64, error) {
	var count int64
	err := r.db.Model(&entities.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

type TenantSettingsRepositoryImpl struct {
	db *gorm.DB
}

func NewTenantSettingsRepository(db *gorm.DB) repositories.TenantSettingsRepository {
	return &TenantSettingsRepositoryImpl{db: db}
}

func (r *TenantSettingsRepositoryImpl) Create(settings *entities.TenantSettings) error {
	return r.db.Create(settings).Error
}

func (r *TenantSettingsRepositoryImpl) FindByTenantID(tenantID string) (*entities.TenantSettings, error) {
	var settings entities.TenantSettings
	err := r.db.Where("tenant_id = ?", tenantID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *TenantSettingsRepositoryImpl) Update(settings *entities.TenantSettings) error {
	return r.db.Save(settings).Error
}

func (r *TenantSettingsRepositoryImpl) Delete(tenantID string) error {
	return r.db.Where("tenant_id = ?", tenantID).Delete(&entities.TenantSettings{}).Error
}
