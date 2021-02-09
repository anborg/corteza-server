package auth

import (
	"context"
	"github.com/cortezaproject/corteza-server/store"
	"github.com/cortezaproject/corteza-server/system/types"
)

type (
	clientService struct {
		store store.AuthClients
	}
)

func (svc clientService) LookupByID(ctx context.Context, clientID uint64) (*types.AuthClient, error) {
	return store.LookupAuthClientByID(ctx, svc.store, clientID)
}
