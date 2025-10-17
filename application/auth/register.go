package auth

import (
	"medical-system/domain/entities"
)

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	TenantID  string `json:"tenant_id" validate:"required,uuid4"`
	Role      string `json:"role" validate:"required,oneof=admin user"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

type RegisterResponse struct {
	User    *entities.User `json:"user"`
	Message string         `json:"message"`
}

type UpdateProfileRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

type UpdateProfileResponse struct {
	User    *entities.User `json:"user"`
	Message string         `json:"message"`
}

func (s *AuthApplicationService) Register(req RegisterRequest) (*RegisterResponse, error) {
	user := &entities.User{
		Email:    req.Email,
		TenantID: req.TenantID,
		Role:     req.Role,
		IsActive: true,
	}

	err := s.authService.RegisterUser(user, req.Password)
	if err != nil {
		return nil, err
	}

	return &RegisterResponse{
		User:    user,
		Message: "User registered successfully",
	}, nil
}

func (s *AuthApplicationService) UpdateProfile(userID string, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
	user, err := s.authService.UpdateProfile(userID, req.FirstName, req.LastName, req.Email)
	if err != nil {
		return nil, err
	}

	return &UpdateProfileResponse{
		User:    user,
		Message: "User updated successfully",
	}, nil
}
