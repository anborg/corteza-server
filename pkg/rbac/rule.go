package rbac

import (
	"fmt"
)

type (
	Rule struct {
		RoleID    uint64    `json:"roleID,string"`
		Resource  Resource  `json:"resource"`
		Operation Operation `json:"operation"`
		Access    Access    `json:"access,string"`

		// Do we need to flush it to storage?
		dirty bool
	}
)

const (
	// Allow - Operation over a resource is allowed
	Allow Access = 1

	// Deny - Operation over a resource is denied
	Deny Access = 0

	// Inherit - Operation over a resource is not defined, inherit
	Inherit Access = -1
)

func (r Rule) String() string {
	return fmt.Sprintf("%s %d to %s on %s", r.Access, r.RoleID, r.Operation, r.Resource)
}

func (r Rule) Equals(cmp *Rule) bool {
	if cmp == nil {
		return false
	}

	return r.RoleID == cmp.RoleID &&
		r.Resource == cmp.Resource &&
		r.Operation == cmp.Operation
}

// AllowRule helper func to create allow rule
func AllowRule(id uint64, r Resource, o Operation) *Rule {
	return &Rule{id, r, o, Allow, false}
}

// DenyRule helper func to create deny rule
func DenyRule(id uint64, r Resource, o Operation) *Rule {
	return &Rule{id, r, o, Deny, false}
}

// InheritRule helper func to create inherit rule
func InheritRule(id uint64, r Resource, o Operation) *Rule {
	return &Rule{id, r, o, Inherit, false}
}
