package repositories

import (
	"medical-system/domain/entities"
	"medical-system/domain/repositories"

	"gorm.io/gorm"
)

type UserRepositoryImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepositoryImpl) FindByEmailAndTenant(email, tenantID string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("email = ? AND tenant_id = ?", email, tenantID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) FindByID(id string) (*entities.User, error) {
	var user entities.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepositoryImpl) Update(user *entities.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepositoryImpl) Delete(id string) error {
	return r.db.Delete(&entities.User{}, "id = ?", id).Error
}
