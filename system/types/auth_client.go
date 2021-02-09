package types

import (
	"github.com/cortezaproject/corteza-server/pkg/filter"
	"github.com/cortezaproject/corteza-server/pkg/rbac"
	"time"
)

type (
	// AuthClient - An organisation may have many authClients. AuthClients may have many channels available. Access to channels may be shared between authClients.
	AuthClient struct {
		ID     uint64 `json:"authClientID,string"`
		Handle string `json:"handle"`
		Name   string `json:"name"`
		Secret string `json:"secret,omitempty"`

		Scope         string `json:"scope"`
		ValidGrant    string `json:"grant"`
		RedirectURI   string `json:"redirectURI"`
		DenyAccessURI string `json:"denyAccessURI"`

		Trusted   bool       `json:"trusted"`
		Enabled   bool       `json:"enabled"`
		ValidFrom *time.Time `json:"validFrom,omitempty"`
		ExpiresAt *time.Time `json:"expiresAt,omitempty"`

		Labels map[string]string `json:"labels,omitempty"`

		OwnedBy   uint64     `json:"ownedBy"`
		CreatedBy uint64     `json:"createdBy"`
		CreatedAt time.Time  `json:"createdAt"`
		UpdatedBy uint64     `json:"updatedBy,omitempty"`
		UpdatedAt *time.Time `json:"updatedAt,omitempty"`
		DeletedBy uint64     `json:"deletedBy,omitempty"`
		DeletedAt *time.Time `json:"deletedAt,omitempty"`
	}

	AuthClientFilter struct {
		AuthClientID []uint64 `json:"authClientID"`

		Handle string `json:"handle"`

		Deleted filter.State `json:"deleted"`

		LabeledIDs []uint64          `json:"-"`
		Labels     map[string]string `json:"labels,omitempty"`

		// Check fn is called by store backend for each resource found function can
		// modify the resource and return false if store should not return it
		//
		// Store then loads additional resources to satisfy the paging parameters
		Check func(*AuthClient) (bool, error) `json:"-"`

		// Standard helpers for paging and sorting
		filter.Sorting
		filter.Paging
	}
)

// Resource returns a resource ID for this type
func (r *AuthClient) RBACResource() rbac.Resource {
	return AuthClientRBACResource.AppendID(r.ID)
}

// FindByHandle finds authClient by it's handle
func (set AuthClientSet) FindByHandle(handle string) *AuthClient {
	for i := range set {
		if set[i].Handle == handle {
			return set[i]
		}
	}

	return nil
}
