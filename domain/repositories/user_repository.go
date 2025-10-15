package repositories

import "medical-system/domain/entities"

type UserRepository interface {
	Create(user *entities.User) error
	FindByEmailAndTenant(email, tenantID string) (*entities.User, error)
	FindByID(id string) (*entities.User, error)
	Update(user *entities.User) error
	Delete(id string) error
}
