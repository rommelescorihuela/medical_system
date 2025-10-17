package auth

import (
	"log"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
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

	// Define the RBAC model with domain support
	modelText := `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && r.obj == p.obj && r.act == p.act
`

	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// Load initial policies
	if err := loadInitialPolicies(enforcer); err != nil {
		log.Printf("Warning: Failed to load initial policies: %v", err)
	}

	log.Println("Casbin RBAC initialized with domain support")
	return &CasbinEnforcer{enforcer: enforcer}, nil
}

func (c *CasbinEnforcer) CheckPermission(userID, tenantID, resource, action string) (bool, error) {
	return c.enforcer.Enforce(userID, tenantID, resource, action)
}

func (c *CasbinEnforcer) AddRoleForUser(userID, role, tenantID string) (bool, error) {
	return c.enforcer.AddRoleForUserInDomain(userID, role, tenantID)
}

// Additional helper methods for RBAC management
func (c *CasbinEnforcer) RemoveRoleForUser(userID, role, tenantID string) (bool, error) {
	return c.enforcer.DeleteRoleForUserInDomain(userID, role, tenantID)
}

func (c *CasbinEnforcer) GetRolesForUser(userID, tenantID string) ([]string, error) {
	roles := c.enforcer.GetRolesForUserInDomain(userID, tenantID)
	return roles, nil
}

func (c *CasbinEnforcer) GetUsersForRole(role, tenantID string) ([]string, error) {
	users := c.enforcer.GetUsersForRoleInDomain(role, tenantID)
	return users, nil
}

func (c *CasbinEnforcer) HasRoleForUser(userID, role, tenantID string) (bool, error) {
	roles, err := c.GetRolesForUser(userID, tenantID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == role {
			return true, nil
		}
	}
	return false, nil
}

func (c *CasbinEnforcer) AddPolicy(role, tenantID, resource, action string) (bool, error) {
	return c.enforcer.AddPolicy(role, tenantID, resource, action)
}

func (c *CasbinEnforcer) RemovePolicy(role, tenantID, resource, action string) (bool, error) {
	return c.enforcer.RemovePolicy(role, tenantID, resource, action)
}

func (c *CasbinEnforcer) GetPolicies() ([][]string, error) {
	policies, err := c.enforcer.GetPolicy()
	return policies, err
}

func loadInitialPolicies(enforcer *casbin.Enforcer) error {
	// Add default policies for multi-tenant RBAC
	policies := [][]string{
		// Admin policies - can manage users and tenants
		{"admin", "*", "user", "read"},
		{"admin", "*", "user", "write"},
		{"admin", "*", "user", "create"},
		{"admin", "*", "user", "delete"},
		{"admin", "*", "tenant", "read"},
		{"admin", "*", "tenant", "write"},
		{"admin", "*", "tenant", "create"},
		{"admin", "*", "tenant", "delete"},

		// User policies - can manage their own profile
		{"user", "*", "profile", "read"},
		{"user", "*", "profile", "write"},

		// Super admin policies - can manage everything across all tenants
		{"super_admin", "*", "*", "*"},
	}

	for _, policy := range policies {
		_, err := enforcer.AddPolicy(policy)
		if err != nil {
			log.Printf("Warning: Failed to add policy %v: %v", policy, err)
			continue
		}
	}

	// Add default role assignments (these would typically be set when users are created)
	// For now, we'll add some example assignments
	roleAssignments := [][]string{
		{"user1", "admin", "*"}, // user1 is admin in all tenants
		{"user2", "user", "*"},  // user2 is regular user in all tenants
	}

	for _, assignment := range roleAssignments {
		_, err := enforcer.AddRoleForUserInDomain(assignment[0], assignment[1], assignment[2])
		if err != nil {
			log.Printf("Warning: Failed to add role assignment %v: %v", assignment, err)
			continue
		}
	}

	return nil
}
