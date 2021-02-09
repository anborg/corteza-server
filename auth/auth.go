package auth

import (
	"context"
	"github.com/cortezaproject/corteza-server/auth/federated"
	"github.com/cortezaproject/corteza-server/auth/handlers"
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/oauth2"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/cortezaproject/corteza-server/auth/settings"
	"github.com/cortezaproject/corteza-server/pkg/options"
	"github.com/cortezaproject/corteza-server/store"
	systemService "github.com/cortezaproject/corteza-server/system/service"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"html/template"
)

type (
	service struct {
		handlers *handlers.AuthHandlers
		log      *zap.Logger
		opt      options.AuthOpt
		s        *settings.Settings
	}
)

func New(ctx context.Context, log *zap.Logger, s store.Storer, opt options.AuthOpt) (svc *service, err error) {
	var (
		tpls *template.Template
	)

	svc = &service{
		opt: opt,
		log: log,
		s:   &settings.Settings{
			// all disabled by default.
		},
	}

	if tpls, err = templates.Load(); err != nil {
		return
	}

	sesManager := session.NewManager(s, opt, log)

	oauth2Manager := oauth2.NewManager(
		opt,
		&oauth2.ContextClientStore{},
		&oauth2.CortezaTokenStore{Store: s},
	)

	oauth2Server := oauth2.NewServer(
		oauth2Manager,
		oauth2.NewUserAuthorizer(
			sesManager,
			templates.GetLinks().Login,
			templates.GetLinks().OAuth2AuthorizeClient,
		),
	)

	svc.handlers = &handlers.AuthHandlers{
		Log:            zap.NewNop(),
		Templates:      tpls,
		SessionManager: sesManager,
		OAuth2:         oauth2Server,
		AuthService:    systemService.DefaultAuth,
		ClientService:  &clientService{s},
		Opt:            svc.opt,
		Settings:       svc.s,
	}

	federated.Init(sesManager.Store())

	if opt.LogEnabled {
		svc.handlers.Log = log.
			Named("auth").
			WithOptions(zap.AddStacktrace(zap.PanicLevel))
	}

	return
}

func (svc *service) UpdateSettings(s *settings.Settings) {
	if svc.s.LocalEnabled != s.LocalEnabled {
		svc.log.Debug("setting changed", zap.Bool("localEnabled", s.LocalEnabled))
	}

	if svc.s.SignupEnabled != s.SignupEnabled {
		svc.log.Debug("setting changed", zap.Bool("signupEnabled", s.SignupEnabled))
	}

	if svc.s.EmailConfirmationRequired != s.EmailConfirmationRequired {
		svc.log.Debug("setting changed", zap.Bool("emailConfirmationRequired", s.EmailConfirmationRequired))
	}

	if svc.s.PasswordResetEnabled != s.PasswordResetEnabled {
		svc.log.Debug("setting changed", zap.Bool("passwordResetEnabled", s.PasswordResetEnabled))
	}

	if svc.s.FederatedEnabled != s.FederatedEnabled {
		svc.log.Debug("setting changed", zap.Bool("federatedEnabled", s.FederatedEnabled))
	}

	if len(svc.s.Providers) != len(s.Providers) {
		svc.log.Debug("setting changed", zap.Int("providers", len(s.Providers)))
		federated.SetupGothProviders(svc.opt.FederatedRedirectURL, s.Providers...)
	}

	svc.s = s
	svc.handlers.Settings = s
}

func (svc service) MountHttpRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		svc.handlers.MountHttpRoutes(r)
	})
}

//func (svc service) WellKnownOpenIDConfiguration() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		json.NewEncoder(w).Encode(map[string]interface{}{
//			"issuer":                                svc.opt.BaseURL,
//			"authorization_endpoint":                svc.opt.BaseURL + "/oauth2/authorize",
//			"token_endpoint":                        svc.opt.BaseURL + "/oauth2/token",
//			"jwks_uri":                              svc.opt.BaseURL + "/oauth2/public-keys", // @todo
//			"subject_types_supported":               []string{"public"},
//			"response_types_supported":              []string{"public"},
//			"id_token_signing_alg_values_supported": []string{"RS256", "HS512"},
//		})
//
//		w.Header().Set("Content-Type", "application/json")
//	}
//}
