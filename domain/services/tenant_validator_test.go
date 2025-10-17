package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"medical-system/domain/entities"
	apperrors "medical-system/internal/errors"
)

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *entities.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) FindByID(id string) (*entities.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindBySlug(slug string) (*entities.Tenant, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) FindByEmail(email string) (*entities.Tenant, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) ListActive() ([]*entities.Tenant, error) {
	args := m.Called()
	return args.Get(0).([]*entities.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(tenant *entities.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTenantRepository) CountUsersByTenant(tenantID string) (int64, error) {
	args := m.Called(tenantID)
	return args.Get(0).(int64), args.Error(1)
}

// MockTenantSettingsRepository is a mock implementation of TenantSettingsRepository
type MockTenantSettingsRepository struct {
	mock.Mock
}

func (m *MockTenantSettingsRepository) Create(settings *entities.TenantSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockTenantSettingsRepository) FindByTenantID(tenantID string) (*entities.TenantSettings, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TenantSettings), args.Error(1)
}

func (m *MockTenantSettingsRepository) Update(settings *entities.TenantSettings) error {
	args := m.Called(settings)
	return args.Error(0)
}

func (m *MockTenantSettingsRepository) Delete(tenantID string) error {
	args := m.Called(tenantID)
	return args.Error(0)
}

func TestTenantValidatorImpl_ValidateTenantLimits_Success(t *testing.T) {
	mockTenantRepo := new(MockTenantRepository)
	mockSettingsRepo := new(MockTenantSettingsRepository)
	validator := &TenantValidatorImpl{
		tenantRepo:         mockTenantRepo,
		tenantSettingsRepo: mockSettingsRepo,
	}

	tenant := &entities.Tenant{
		ID:   "1",
		Name: "Test Tenant",
		Plan: entities.PlanBasic,
	}

	settings := &entities.TenantSettings{
		TenantID: "1",
		MaxUsers: 5,
	}

	mockTenantRepo.On("FindByID", "1").Return(tenant, nil)
	mockSettingsRepo.On("FindByTenantID", "1").Return(settings, nil)
	mockTenantRepo.On("CountUsersByTenant", "1").Return(int64(3), nil)

	err := validator.ValidateTenantLimits("1")

	assert.NoError(t, err)
	mockTenantRepo.AssertExpectations(t)
	mockSettingsRepo.AssertExpectations(t)
}

func TestTenantValidatorImpl_ValidateTenantLimits_TenantNotFound(t *testing.T) {
	mockTenantRepo := new(MockTenantRepository)
	mockSettingsRepo := new(MockTenantSettingsRepository)
	validator := &TenantValidatorImpl{
		tenantRepo:         mockTenantRepo,
		tenantSettingsRepo: mockSettingsRepo,
	}

	mockTenantRepo.On("FindByID", "999").Return(nil, errors.New("tenant not found"))

	err := validator.ValidateTenantLimits("999")

	assert.Error(t, err)
	mockTenantRepo.AssertExpectations(t)
	mockSettingsRepo.AssertNotCalled(t, "FindByTenantID")
	mockTenantRepo.AssertNotCalled(t, "CountUsersByTenant")
}

func TestTenantValidatorImpl_ValidateTenantLimits_SettingsNotFound(t *testing.T) {
	mockTenantRepo := new(MockTenantRepository)
	mockSettingsRepo := new(MockTenantSettingsRepository)
	validator := &TenantValidatorImpl{
		tenantRepo:         mockTenantRepo,
		tenantSettingsRepo: mockSettingsRepo,
	}

	tenant := &entities.Tenant{
		ID:   "1",
		Name: "Test Tenant",
		Plan: entities.PlanBasic,
	}

	mockTenantRepo.On("FindByID", "1").Return(tenant, nil)
	mockSettingsRepo.On("FindByTenantID", "1").Return(nil, errors.New("settings not found"))

	err := validator.ValidateTenantLimits("1")

	assert.Error(t, err)
	mockTenantRepo.AssertExpectations(t)
	mockSettingsRepo.AssertExpectations(t)
	mockTenantRepo.AssertNotCalled(t, "CountUsersByTenant")
}

func TestTenantValidatorImpl_ValidateTenantLimits_LimitExceeded(t *testing.T) {
	mockTenantRepo := new(MockTenantRepository)
	mockSettingsRepo := new(MockTenantSettingsRepository)
	validator := &TenantValidatorImpl{
		tenantRepo:         mockTenantRepo,
		tenantSettingsRepo: mockSettingsRepo,
	}

	tenant := &entities.Tenant{
		ID:   "1",
		Name: "Test Tenant",
		Plan: entities.PlanBasic,
	}

	settings := &entities.TenantSettings{
		TenantID: "1",
		MaxUsers: 5,
	}

	mockTenantRepo.On("FindByID", "1").Return(tenant, nil)
	mockSettingsRepo.On("FindByTenantID", "1").Return(settings, nil)
	mockTenantRepo.On("CountUsersByTenant", "1").Return(int64(5), nil)

	err := validator.ValidateTenantLimits("1")

	assert.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, apperrors.ErrorTypeConflict, appErr.Type)
	assert.Contains(t, appErr.Message, "tenant has reached maximum user limit")
	mockTenantRepo.AssertExpectations(t)
	mockSettingsRepo.AssertExpectations(t)
}

func TestTenantValidatorImpl_ValidateTenantLimits_UsePlanLimit(t *testing.T) {
	mockTenantRepo := new(MockTenantRepository)
	mockSettingsRepo := new(MockTenantSettingsRepository)
	validator := &TenantValidatorImpl{
		tenantRepo:         mockTenantRepo,
		tenantSettingsRepo: mockSettingsRepo,
	}

	tenant := &entities.Tenant{
		ID:   "1",
		Name: "Test Tenant",
		Plan: entities.PlanBasic,
	}

	settings := &entities.TenantSettings{
		TenantID: "1",
		MaxUsers: 0, // Use plan limit
	}

	mockTenantRepo.On("FindByID", "1").Return(tenant, nil)
	mockSettingsRepo.On("FindByTenantID", "1").Return(settings, nil)
	mockTenantRepo.On("CountUsersByTenant", "1").Return(int64(4), nil)

	err := validator.ValidateTenantLimits("1")

	assert.NoError(t, err)
	mockTenantRepo.AssertExpectations(t)
	mockSettingsRepo.AssertExpectations(t)
}
