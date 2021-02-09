package provision

import (
	"context"
	"github.com/cortezaproject/corteza-server/pkg/errors"
	"github.com/cortezaproject/corteza-server/pkg/id"
	"github.com/cortezaproject/corteza-server/pkg/options"
	"github.com/cortezaproject/corteza-server/pkg/rand"
	"github.com/cortezaproject/corteza-server/store"
	"github.com/cortezaproject/corteza-server/system/service"
	"github.com/cortezaproject/corteza-server/system/types"
	"go.uber.org/zap"
	"time"
)

func Run(ctx context.Context, log *zap.Logger, s store.Storer, provisionOpt options.ProvisionOpt, authOpt options.AuthOpt) error {
	ffn := []func() error{
		func() error { return roles(ctx, s) },
		func() error { return importConfig(ctx, log, s, provisionOpt.Path) },
		func() error { return authSettingsAutoDiscovery(ctx, log, service.DefaultSettings) },
		func() error { return authAddExternals(ctx, log) },
		func() error { return service.DefaultSettings.UpdateCurrent(ctx) },
		func() error { return oidcAutoDiscovery(ctx, log, authOpt) },
		func() error { return defaultAuthClient(ctx, log, s) },
	}

	for _, fn := range ffn {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func defaultAuthClient(ctx context.Context, log *zap.Logger, s store.AuthClients) error {
	clients := types.AuthClientSet{
		&types.AuthClient{
			ID:        id.Next(),
			Handle:    "corteza-webapp",
			Name:      "Corteza Web Applications",
			Secret:    string(rand.Bytes(64)),
			Enabled:   true,
			Trusted:   true,
			Labels:    nil,
			CreatedAt: time.Now(),
		},
	}

	for _, c := range clients {
		_, err := store.LookupAuthClientByHandle(ctx, s, c.Handle)
		if err == nil {
			continue
		}

		if !errors.IsNotFound(err) {
			return err
		}

		if err = store.CreateAuthClient(ctx, s, c); err != nil {
			return err
		}

		log.Info("Added OAuth2 client", zap.String("name", c.Name), zap.Uint64("clientId", c.ID))
	}

	return nil
}
