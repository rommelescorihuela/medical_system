package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SubscriptionPlan represents the different subscription tiers
type SubscriptionPlan string

const (
	PlanBasic        SubscriptionPlan = "basic"
	PlanProfessional SubscriptionPlan = "professional"
	PlanEnterprise   SubscriptionPlan = "enterprise"
)

// GetUserLimit returns the maximum number of users allowed for the plan
func (p SubscriptionPlan) GetUserLimit() int {
	switch p {
	case PlanBasic:
		return 5
	case PlanProfessional:
		return 50
	case PlanEnterprise:
		return 1000
	default:
		return 0
	}
}

type Tenant struct {
	ID        string           `json:"id" gorm:"primaryKey"`
	Name      string           `json:"name" gorm:"uniqueIndex;not null"`
	Slug      string           `json:"slug" gorm:"uniqueIndex;not null"`
	Email     string           `json:"email" gorm:"uniqueIndex;not null"`
	Plan      SubscriptionPlan `json:"plan" gorm:"default:basic"`
	IsActive  bool             `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type TenantSettings struct {
	ID       string `json:"id" gorm:"primaryKey"`
	TenantID string `json:"tenant_id" gorm:"uniqueIndex;not null"`
	Tenant   Tenant `json:"-" gorm:"foreignKey:TenantID;references:ID;constraint:OnDelete:CASCADE"`

	// Configuration settings
	AllowUserRegistration bool   `json:"allow_user_registration" gorm:"default:true"`
	MaxUsers              int    `json:"max_users"`
	Timezone              string `json:"timezone" gorm:"default:UTC"`
	Language              string `json:"language" gorm:"default:en"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ts *TenantSettings) BeforeCreate(tx *gorm.DB) error {
	if ts.ID == "" {
		ts.ID = uuid.New().String()
	}
	// Set max users based on plan if not explicitly set
	if ts.MaxUsers == 0 && ts.Tenant.Plan != "" {
		ts.MaxUsers = ts.Tenant.Plan.GetUserLimit()
	}
	return nil
}
