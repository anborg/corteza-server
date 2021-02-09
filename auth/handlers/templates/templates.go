package templates

import (
	"github.com/Masterminds/sprig"
	"github.com/cortezaproject/corteza-server/pkg/version"
	"html/template"
)

type (
	Links struct {
		Profile,
		Signup,
		PendingEmailConfirmation,
		Login,
		ChangePassword,
		RequestPasswordReset,
		PasswordResetRequested,
		ResetPassword,
		Sessions,
		Logout,

		OAuth2Authorize,
		OAuth2AuthorizeClient,
		OAuth2Token,
		OAuth2Info,

		Federated,

		Assets string
	}
)

func GetLinks() Links {
	return Links{
		Profile:                  "/auth",
		Signup:                   "/auth/signup",
		PendingEmailConfirmation: "/auth/pending-email-confirmation",
		Login:                    "/auth/login",
		ChangePassword:           "/auth/change-password",
		RequestPasswordReset:     "/auth/request-password-reset",
		PasswordResetRequested:   "/auth/password-reset-requested",
		ResetPassword:            "/auth/reset-password",
		Sessions:                 "/auth/sessions",
		Logout:                   "/auth/logout",

		OAuth2Authorize:       "/auth/oauth2/authorize",
		OAuth2AuthorizeClient: "/auth/oauth2/authorize-client",
		OAuth2Token:           "/auth/oauth2/token",
		OAuth2Info:            "/auth/oauth2/info",

		Federated: "/auth/federated",

		Assets: "/auth/assets",
	}
}

func Load() (*template.Template, error) {
	return template.New("").
		Funcs(sprig.FuncMap()).
		Funcs(map[string]interface{}{
			"version":   func() string { return version.Version },
			"buildtime": func() string { return version.BuildTime },
			"links":     func() Links { return GetLinks() },
		}).
		ParseGlob("auth/handlers/templates/*.tpl")
}
