package services

import (
	"medical-system/domain/repositories"
	apperrors "medical-system/internal/errors"
)

type TenantValidatorImpl struct {
	tenantRepo         repositories.TenantRepository
	tenantSettingsRepo repositories.TenantSettingsRepository
}

func NewTenantValidator(tenantRepo repositories.TenantRepository, tenantSettingsRepo repositories.TenantSettingsRepository) TenantValidator {
	return &TenantValidatorImpl{
		tenantRepo:         tenantRepo,
		tenantSettingsRepo: tenantSettingsRepo,
	}
}

func (v *TenantValidatorImpl) ValidateTenantLimits(tenantID string) error {
	tenant, err := v.tenantRepo.FindByID(tenantID)
	if err != nil {
		return err
	}

	settings, err := v.tenantSettingsRepo.FindByTenantID(tenantID)
	if err != nil {
		return err
	}

	userCount, err := v.tenantRepo.CountUsersByTenant(tenantID)
	if err != nil {
		return err
	}

	maxUsers := settings.MaxUsers
	if maxUsers == 0 {
		maxUsers = tenant.Plan.GetUserLimit()
	}

	if int(userCount) >= maxUsers {
		return apperrors.NewConflictError("tenant has reached maximum user limit")
	}

	return nil
}
