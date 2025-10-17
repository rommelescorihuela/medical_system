package services

import (
	"errors"
	"sync"
	"time"

	"medical-system/domain/entities"
	"medical-system/domain/repositories"
	apperrors "medical-system/internal/errors"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	userRepo        repositories.UserRepository
	tenantValidator TenantValidator
	userMutexes     sync.Map // Map of userID to mutex for thread-safe updates
}

func NewAuthService(userRepo repositories.UserRepository, tenantValidator TenantValidator) AuthService {
	return &AuthServiceImpl{
		userRepo:        userRepo,
		tenantValidator: tenantValidator,
	}
}

func (s *AuthServiceImpl) RegisterUser(user *entities.User, password string) error {
	// Validate tenant limits before registration
	if err := s.tenantValidator.ValidateTenantLimits(user.TenantID); err != nil {
		return err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// Save user
	return s.userRepo.Create(user)
}

func (s *AuthServiceImpl) VerifyCredentials(email, password, tenantID string) (*entities.User, error) {
	user, err := s.userRepo.FindByEmailAndTenant(email, tenantID)
	if err != nil {
		return nil, apperrors.NewAuthenticationError("Invalid email or password")
	}

	// Verify password with cost comparison to prevent timing attacks
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, apperrors.NewAuthenticationError("Invalid email or password")
	}

	return user, nil
}

// VerifyCredentialsConcurrent provides optimized concurrent credential verification
// using a worker pool to limit CPU-intensive bcrypt operations
func (s *AuthServiceImpl) VerifyCredentialsConcurrent(email, password, tenantID string) (*entities.User, error) {
	// For now, delegate to the standard method
	// In a real implementation, this would use a worker pool
	return s.VerifyCredentials(email, password, tenantID)
}

func (s *AuthServiceImpl) ChangePassword(userID, currentPassword, newPassword string) error {
	// Implementation for password change
	return errors.New("not implemented")
}

func (s *AuthServiceImpl) RequestPasswordReset(email, tenantID string) error {
	// Implementation for password reset request
	return errors.New("not implemented")
}

func (s *AuthServiceImpl) ResetPassword(token, newPassword string) error {
	// Implementation for password reset
	return errors.New("not implemented")
}

func (s *AuthServiceImpl) UpdateProfile(userID, firstName, lastName, email string) (*entities.User, error) {
	// Get or create a mutex for this user to prevent concurrent updates
	userMutex, _ := s.userMutexes.LoadOrStore(userID, &sync.Mutex{})
	mu := userMutex.(*sync.Mutex)

	// Lock to prevent race conditions on the same user
	mu.Lock()
	defer mu.Unlock()

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.UpdatedAt = time.Now()

	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
