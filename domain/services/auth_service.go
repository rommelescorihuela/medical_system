package services

import "medical-system/domain/entities"

type AuthService interface {
	RegisterUser(user *entities.User, password string) error
	VerifyCredentials(email, password, tenantID string) (*entities.User, error)
	ChangePassword(userID, currentPassword, newPassword string) error
	RequestPasswordReset(email, tenantID string) error
	ResetPassword(token, newPassword string) error
}
