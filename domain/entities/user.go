package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           string `json:"id" gorm:"primaryKey"`
	Email        string `json:"email" gorm:"uniqueIndex:idx_user_email_tenant"`
	TenantID     string `json:"tenant_id" gorm:"uniqueIndex:idx_user_email_tenant"`
	PasswordHash string `json:"-" gorm:"column:password_hash"`
	Role         string `json:"role" gorm:"default:user"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}
