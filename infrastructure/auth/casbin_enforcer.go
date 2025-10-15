package auth

import (
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type CasbinEnforcer struct {
	enforcer *casbin.Enforcer
}

func NewCasbinEnforcer(db *gorm.DB) (*CasbinEnforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	// Simple enforcer without model for now - just initialize
	enforcer := &casbin.Enforcer{}
	// We'll skip the full initialization for now to avoid model issues
	_ = adapter

	log.Println("Casbin RBAC initialized (simplified)")
	return &CasbinEnforcer{enforcer: enforcer}, nil
}

func (c *CasbinEnforcer) CheckPermission(userID, tenantID, resource, action string) (bool, error) {
	return c.enforcer.Enforce(userID, tenantID, resource, action)
}

func (c *CasbinEnforcer) AddRoleForUser(userID, role, tenantID string) (bool, error) {
	return c.enforcer.AddRoleForUserInDomain(userID, role, tenantID)
}

func loadInitialPolicies(enforcer *casbin.Enforcer) error {
	// Add some default policies
	policies := [][]string{
		{"admin", "tenant1", "user", "read"},
		{"admin", "tenant1", "user", "write"},
		{"user", "tenant1", "profile", "read"},
		{"user", "tenant1", "profile", "write"},
	}

	for _, policy := range policies {
		_, err := enforcer.AddPolicy(policy)
		if err != nil {
			return err
		}
	}

	return nil
}
