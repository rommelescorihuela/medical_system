package tenants

import (
	"medical-system/domain/entities"
	"medical-system/domain/services"
)

type TenantApplicationService struct {
	tenantService services.TenantService
}

type RegisterTenantRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Slug  string `json:"slug" validate:"required"`
	Plan  string `json:"plan" validate:"required,oneof=basic professional enterprise"`
}

type RegisterTenantResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Slug  string `json:"slug"`
	Plan  string `json:"plan"`
}

type TenantSettingsResponse struct {
	TenantID              string `json:"tenant_id"`
	AllowUserRegistration bool   `json:"allow_user_registration"`
	MaxUsers              int    `json:"max_users"`
	Timezone              string `json:"timezone"`
	Language              string `json:"language"`
}

func NewTenantApplicationService(tenantService services.TenantService) *TenantApplicationService {
	return &TenantApplicationService{
		tenantService: tenantService,
	}
}

func (s *TenantApplicationService) RegisterTenant(req RegisterTenantRequest) (*RegisterTenantResponse, error) {
	plan := entities.SubscriptionPlan(req.Plan)

	tenant, err := s.tenantService.CreateTenant(req.Name, req.Email, req.Slug, plan)
	if err != nil {
		return nil, err
	}

	return &RegisterTenantResponse{
		ID:    tenant.ID,
		Name:  tenant.Name,
		Email: tenant.Email,
		Slug:  tenant.Slug,
		Plan:  string(tenant.Plan),
	}, nil
}

func (s *TenantApplicationService) GetTenantBySlug(slug string) (*entities.Tenant, error) {
	return s.tenantService.GetTenantBySlug(slug)
}

func (s *TenantApplicationService) GetTenantSettings(tenantID string) (*TenantSettingsResponse, error) {
	settings, err := s.tenantService.GetTenantSettings(tenantID)
	if err != nil {
		return nil, err
	}

	return &TenantSettingsResponse{
		TenantID:              settings.TenantID,
		AllowUserRegistration: settings.AllowUserRegistration,
		MaxUsers:              settings.MaxUsers,
		Timezone:              settings.Timezone,
		Language:              settings.Language,
	}, nil
}

func (s *TenantApplicationService) UpdateTenantSettings(tenantID string, allowRegistration bool, maxUsers int, timezone, language string) error {
	settings, err := s.tenantService.GetTenantSettings(tenantID)
	if err != nil {
		return err
	}

	settings.AllowUserRegistration = allowRegistration
	settings.MaxUsers = maxUsers
	settings.Timezone = timezone
	settings.Language = language

	return s.tenantService.UpdateTenantSettings(settings)
}

func (s *TenantApplicationService) ValidateTenantForUserRegistration(tenantID string) error {
	return s.tenantService.ValidateTenantLimits(tenantID)
}

// Admin functions for tenant management
func (s *TenantApplicationService) ListActiveTenants() ([]*entities.Tenant, error) {
	return s.tenantService.ListActiveTenants()
}

func (s *TenantApplicationService) GetTenantByID(id string) (*entities.Tenant, error) {
	return s.tenantService.GetTenantByID(id)
}

func (s *TenantApplicationService) UpdateTenant(tenant *entities.Tenant) error {
	return s.tenantService.UpdateTenant(tenant)
}

func (s *TenantApplicationService) DeleteTenant(id string) error {
	return s.tenantService.DeleteTenant(id)
}
