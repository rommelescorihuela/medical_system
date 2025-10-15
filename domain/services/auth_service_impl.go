package services

import (
	"errors"

	"medical-system/domain/entities"
	"medical-system/domain/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	userRepo repositories.UserRepository
}

func NewAuthService(userRepo repositories.UserRepository) AuthService {
	return &AuthServiceImpl{userRepo: userRepo}
}

func (s *AuthServiceImpl) RegisterUser(user *entities.User, password string) error {
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
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
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
