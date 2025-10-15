package auth

import (
	"medical-system/domain/entities"
	"medical-system/domain/repositories"
	"medical-system/domain/services"
	"medical-system/infrastructure/auth"
)

type AuthApplicationService struct {
	userRepo     repositories.UserRepository
	authService  services.AuthService
	tokenGen     auth.TokenGenerator
	rbacEnforcer *auth.CasbinEnforcer
}

func NewAuthApplicationService(
	userRepo repositories.UserRepository,
	authService services.AuthService,
	tokenGen auth.TokenGenerator,
	rbacEnforcer *auth.CasbinEnforcer,
) *AuthApplicationService {
	return &AuthApplicationService{
		userRepo:     userRepo,
		authService:  authService,
		tokenGen:     tokenGen,
		rbacEnforcer: rbacEnforcer,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TenantID string `json:"tenant_id"`
}

type LoginResponse struct {
	User        *entities.User `json:"user"`
	AccessToken string         `json:"access_token"`
	Permissions []string       `json:"permissions"`
}

func (s *AuthApplicationService) Login(req LoginRequest) (*LoginResponse, error) {
	user, err := s.authService.VerifyCredentials(req.Email, req.Password, req.TenantID)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenGen.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	s.rbacEnforcer.AddRoleForUser(user.ID, user.Role, user.TenantID)

	return &LoginResponse{
		User:        user,
		AccessToken: token,
		Permissions: []string{},
	}, nil
}
