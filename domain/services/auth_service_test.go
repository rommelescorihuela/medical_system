package services

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"medical-system/domain/entities"
	apperrors "medical-system/internal/errors"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmailAndTenant(email, tenantID string) (*entities.User, error) {
	args := m.Called(email, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id string) (*entities.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entities.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockTenantValidator is a mock implementation of TenantValidator
type MockTenantValidator struct {
	mock.Mock
}

func (m *MockTenantValidator) ValidateTenantLimits(tenantID string) error {
	args := m.Called(tenantID)
	return args.Error(0)
}

func TestAuthServiceImpl_RegisterUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		Email:    "newuser@example.com",
		TenantID: "1",
	}

	mockValidator.On("ValidateTenantLimits", "1").Return(nil)
	mockRepo.On("Create", mock.AnythingOfType("*entities.User")).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(0).(*entities.User)
		assert.Equal(t, "newuser@example.com", u.Email)
		assert.Equal(t, "1", u.TenantID)
		assert.NotEmpty(t, u.PasswordHash) // Should be hashed
	})

	err := authService.RegisterUser(user, "password123")

	assert.NoError(t, err)
	mockValidator.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthServiceImpl_RegisterUser_TenantLimitExceeded(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		Email:    "newuser@example.com",
		TenantID: "1",
	}

	mockValidator.On("ValidateTenantLimits", "1").Return(apperrors.NewConflictError("tenant has reached maximum user limit"))

	err := authService.RegisterUser(user, "password123")

	assert.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, apperrors.ErrorTypeConflict, appErr.Type)
	mockValidator.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestAuthServiceImpl_VerifyCredentials_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // bcrypt hash for "password"
		TenantID:     "1",
	}

	mockRepo.On("FindByEmailAndTenant", "test@example.com", "1").Return(user, nil)

	result, err := authService.VerifyCredentials("test@example.com", "password", "1")

	assert.NoError(t, err)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.TenantID, result.TenantID)
	mockRepo.AssertExpectations(t)
}

func TestAuthServiceImpl_VerifyCredentials_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	mockRepo.On("FindByEmailAndTenant", "nonexistent@example.com", "1").Return(nil, errors.New("user not found"))

	_, err := authService.VerifyCredentials("nonexistent@example.com", "password", "1")

	assert.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, apperrors.ErrorTypeAuthentication, appErr.Type)
	mockRepo.AssertExpectations(t)
}

func TestAuthServiceImpl_VerifyCredentials_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		Email:        "test@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // bcrypt hash for "password"
		TenantID:     "1",
	}

	mockRepo.On("FindByEmailAndTenant", "test@example.com", "1").Return(user, nil)

	_, err := authService.VerifyCredentials("test@example.com", "wrongpassword", "1")

	assert.Error(t, err)
	assert.IsType(t, &apperrors.AppError{}, err)
	appErr := err.(*apperrors.AppError)
	assert.Equal(t, apperrors.ErrorTypeAuthentication, appErr.Type)
	mockRepo.AssertExpectations(t)
}

func TestAuthServiceImpl_UpdateProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	user := &entities.User{
		ID:        "1",
		Email:     "old@example.com",
		FirstName: "Old",
		LastName:  "Name",
	}

	mockRepo.On("FindByID", "1").Return(user, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entities.User")).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(0).(*entities.User)
		assert.Equal(t, "New", u.FirstName)
		assert.Equal(t, "Name", u.LastName)
		assert.Equal(t, "new@example.com", u.Email)
	})

	result, err := authService.UpdateProfile("1", "New", "Name", "new@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "New", result.FirstName)
	assert.Equal(t, "Name", result.LastName)
	assert.Equal(t, "new@example.com", result.Email)
	mockRepo.AssertExpectations(t)
}

func TestAuthServiceImpl_UpdateProfile_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockValidator := new(MockTenantValidator)
	authService := &AuthServiceImpl{
		userRepo:        mockRepo,
		tenantValidator: mockValidator,
	}

	mockRepo.On("FindByID", "999").Return(nil, errors.New("user not found"))

	_, err := authService.UpdateProfile("999", "New", "Name", "new@example.com")

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
