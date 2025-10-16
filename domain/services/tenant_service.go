package services

import (
	"errors"
	"medical-system/domain/entities"
	"medical-system/domain/repositories"
	"strings"
)

type TenantService interface {
	CreateTenant(name, email, slug string, plan entities.SubscriptionPlan) (*entities.Tenant, error)
	GetTenantByID(id string) (*entities.Tenant, error)
	GetTenantBySlug(slug string) (*entities.Tenant, error)
	UpdateTenant(tenant *entities.Tenant) error
	DeleteTenant(id string) error
	ListActiveTenants() ([]*entities.Tenant, error)
	ValidateTenantLimits(tenantID string) error
	GetTenantSettings(tenantID string) (*entities.TenantSettings, error)
	UpdateTenantSettings(settings *entities.TenantSettings) error
}

type TenantServiceImpl struct {
	tenantRepo         repositories.TenantRepository
	tenantSettingsRepo repositories.TenantSettingsRepository
}

func NewTenantService(tenantRepo repositories.TenantRepository, tenantSettingsRepo repositories.TenantSettingsRepository) TenantService {
	return &TenantServiceImpl{
		tenantRepo:         tenantRepo,
		tenantSettingsRepo: tenantSettingsRepo,
	}
}

func (s *TenantServiceImpl) CreateTenant(name, email, slug string, plan entities.SubscriptionPlan) (*entities.Tenant, error) {
	// Validate inputs
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("tenant name cannot be empty")
	}
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("tenant email cannot be empty")
	}
	if strings.TrimSpace(slug) == "" {
		return nil, errors.New("tenant slug cannot be empty")
	}

	// Check if email or slug already exists
	if _, err := s.tenantRepo.FindByEmail(email); err == nil {
		return nil, errors.New("tenant with this email already exists")
	}
	if _, err := s.tenantRepo.FindBySlug(slug); err == nil {
		return nil, errors.New("tenant with this slug already exists")
	}

	// Create tenant
	tenant := &entities.Tenant{
		Name:     name,
		Email:    email,
		Slug:     slug,
		Plan:     plan,
		IsActive: true,
	}

	if err := s.tenantRepo.Create(tenant); err != nil {
		return nil, err
	}

	// Create default settings
	settings := &entities.TenantSettings{
		TenantID:              tenant.ID,
		AllowUserRegistration: true,
		MaxUsers:              plan.GetUserLimit(),
		Timezone:              "UTC",
		Language:              "en",
	}

	if err := s.tenantSettingsRepo.Create(settings); err != nil {
		// If settings creation fails, we should probably delete the tenant
		// But for simplicity, we'll just return the error
		return nil, err
	}

	return tenant, nil
}

func (s *TenantServiceImpl) GetTenantByID(id string) (*entities.Tenant, error) {
	return s.tenantRepo.FindByID(id)
}

func (s *TenantServiceImpl) GetTenantBySlug(slug string) (*entities.Tenant, error) {
	return s.tenantRepo.FindBySlug(slug)
}

func (s *TenantServiceImpl) UpdateTenant(tenant *entities.Tenant) error {
	return s.tenantRepo.Update(tenant)
}

func (s *TenantServiceImpl) DeleteTenant(id string) error {
	// First delete settings, then tenant
	if err := s.tenantSettingsRepo.Delete(id); err != nil {
		return err
	}
	return s.tenantRepo.Delete(id)
}

func (s *TenantServiceImpl) ListActiveTenants() ([]*entities.Tenant, error) {
	return s.tenantRepo.ListActive()
}

func (s *TenantServiceImpl) ValidateTenantLimits(tenantID string) error {
	tenant, err := s.tenantRepo.FindByID(tenantID)
	if err != nil {
		return err
	}

	settings, err := s.tenantSettingsRepo.FindByTenantID(tenantID)
	if err != nil {
		return err
	}

	userCount, err := s.tenantRepo.CountUsersByTenant(tenantID)
	if err != nil {
		return err
	}

	maxUsers := settings.MaxUsers
	if maxUsers == 0 {
		maxUsers = tenant.Plan.GetUserLimit()
	}

	if int(userCount) >= maxUsers {
		return errors.New("tenant has reached maximum user limit")
	}

	return nil
}

func (s *TenantServiceImpl) GetTenantSettings(tenantID string) (*entities.TenantSettings, error) {
	return s.tenantSettingsRepo.FindByTenantID(tenantID)
}

func (s *TenantServiceImpl) UpdateTenantSettings(settings *entities.TenantSettings) error {
	return s.tenantSettingsRepo.Update(settings)
}
