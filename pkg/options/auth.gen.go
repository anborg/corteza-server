package options

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated.
//
// Definitions file that controls how this file is generated:
// pkg/options/auth.yaml

import (
	"strings"
	"time"
)

type (
	AuthOpt struct {
		Secret                          string        `env:"AUTH_JWT_SECRET"`
		Expiry                          time.Duration `env:"AUTH_JWT_EXPIRY"`
		FederatedRedirectURL            string        `env:"AUTH_FEDERATED_REDIRECT_URL"`
		FederatedCookieSecret           string        `env:"AUTH_FEDERATED_COOKIE_SECRET"`
		BaseURL                         string        `env:"AUTH_BASE_URL"`
		SessionCookieName               string        `env:"AUTH_SESSION_COOKIE_NAME"`
		SessionLifetime                 time.Duration `env:"AUTH_SESSION_LIFETIME"`
		SessionPermLifetime             time.Duration `env:"AUTH_SESSION_PERM_LIFETIME"`
		SessionSecureCookies            bool          `env:"AUTH_SESSION_SECURE_COOKIES"`
		SessionGarbageCollectorInterval time.Duration `env:"AUTH_SESSION_GARBAGE_COLLECTOR_INTERVAL"`
		RequestRateLimit                int           `env:"AUTH_REQUEST_RATE_LIMIT"`
		RequestRateWindowLength         time.Duration `env:"AUTH_REQUEST_RATE_WINDOW_LENGTH"`
		CsrfSecret                      string        `env:"AUTH_CSRF_SECRET"`
		LogEnabled                      bool          `env:"AUTH_LOG_ENABLED"`
	}
)

// Auth initializes and returns a AuthOpt with default values
func Auth() (o *AuthOpt) {
	o = &AuthOpt{
		Expiry:                          time.Hour * 24 * 30,
		FederatedRedirectURL:            guestBaseURL() + "/auth/federated/{provider}/callback",
		BaseURL:                         guestBaseURL() + "/auth",
		SessionCookieName:               "session",
		SessionLifetime:                 24 * time.Hour,
		SessionPermLifetime:             360 * 24 * time.Hour,
		SessionSecureCookies:            strings.HasPrefix(guestBaseURL(), "https://"),
		SessionGarbageCollectorInterval: 15 * time.Minute,
		RequestRateLimit:                30,
		RequestRateWindowLength:         time.Minute,
	}

	fill(o)

	// Function that allows access to custom logic inside the parent function.
	// The custom logic in the other file should be like:
	// func (o *Auth) Defaults() {...}
	func(o interface{}) {
		if def, ok := o.(interface{ Defaults() }); ok {
			def.Defaults()
		}
	}(o)

	return
}
