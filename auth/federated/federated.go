package federated

import (
	"github.com/cortezaproject/corteza-server/pkg/logger"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
)

const (
	OIDC_PROVIDER_PREFIX = "openid-connect."
)

func Init(store sessions.Store) {
	gothic.Store = store
}

func Disable() {
	goth.ClearProviders()
}

func log() *zap.Logger {
	return logger.Default().Named("auth.federated")
}
