package handlers

import (
	"context"
	"encoding/gob"
	"github.com/cortezaproject/corteza-server/auth/federated"
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/cortezaproject/corteza-server/auth/settings"
	"github.com/cortezaproject/corteza-server/pkg/options"
	"github.com/cortezaproject/corteza-server/system/types"
	"github.com/go-chi/chi"
	"github.com/go-chi/httprate"
	oauth2server "github.com/go-oauth2/oauth2/v4/server"
	"github.com/gorilla/csrf"
	"github.com/markbates/goth"
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type (
	authService interface {
		External(ctx context.Context, profile goth.User) (u *types.User, err error)
		InternalSignUp(ctx context.Context, input *types.User, password string) (u *types.User, err error)
		InternalLogin(ctx context.Context, email string, password string) (u *types.User, err error)
		SetPassword(ctx context.Context, userID uint64, password string) (err error)
		//Impersonate(ctx context.Context, userID uint64) (u *types.User, err error)
		ChangePassword(ctx context.Context, userID uint64, oldPassword, newPassword string) (err error)
		//CheckPasswordStrength(password string) bool
		EmailConfirmationRequired() bool
		//SetPasswordCredentials(ctx context.Context, userID uint64, password string) (err error)
		//IssueAuthRequestToken(ctx context.Context, user *types.User) (token string, err error)
		//ValidateAuthRequestToken(ctx context.Context, token string) (u *types.User, err error)
		ValidateEmailConfirmationToken(ctx context.Context, token string) (user *types.User, err error)
		ValidatePasswordResetToken(ctx context.Context, token string) (user *types.User, err error)
		//ExchangePasswordResetToken(ctx context.Context, token string) (u *types.User, t string, err error)
		SendEmailAddressConfirmationToken(ctx context.Context, u *types.User) (err error)
		SendPasswordResetToken(ctx context.Context, email string) (err error)
		//CanRegister(ctx context.Context) error
		GetProviders() types.ExternalAuthProviderSet
	}

	clientService interface {
		LookupByID(context.Context, uint64) (*types.AuthClient, error)
	}

	AuthHandlers struct {
		Log *zap.Logger

		Templates      *template.Template
		OAuth2         *oauth2server.Server
		SessionManager *session.Manager
		AuthService    authService
		ClientService  clientService
		Opt            options.AuthOpt
		Settings       *settings.Settings
	}

	handlerFn func(p *request.AuthReq) error
)

const (
	tplInternalError = "error-internal.html.tpl"
)

func init() {
	gob.Register(&types.User{})
	gob.Register(&types.AuthClient{})
	gob.Register([]request.Alert{})
	gob.Register(url.Values{})
}

func (h *AuthHandlers) MountHttpRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		if h.Opt.RequestRateLimit > 0 {
			r.Use(httprate.LimitByIP(h.Opt.RequestRateLimit, h.Opt.RequestRateWindowLength)) // @todo make configurable
		}

		r.Use(request.ExtraReqInfoMiddleware)
		//r.Use(request.SaveSessions(h.SessionManager))
		r.Use(csrf.Protect(
			[]byte(h.Opt.CsrfSecret),
			csrf.SameSite(csrf.SameSiteStrictMode),
			csrf.Secure(h.Opt.SessionSecureCookies),
			csrf.FieldName("same-site-authenticity-token"),
		))

		r.HandleFunc("/", h.handle(h.auth(h.profileView)))
		r.HandleFunc("/logout", h.handle(h.logoutProc))
		r.Get("/sessions", h.handle(h.auth(h.sessionsView)))
		r.Post("/sessions", h.handle(h.auth(h.sessionsProc)))

		r.Get("/signup", h.handle(h.anony(h.signupForm)))
		r.Post("/signup", h.handle(h.anony(h.signupProc)))
		r.Get("/pending-email-confirmation", h.handle(h.pendingEmailConfirmation))
		r.Get("/confirm-email", h.handle(h.confirmEmail))

		r.Get("/login", h.handle(h.anony(h.loginForm)))
		r.Post("/login", h.handle(h.anony(h.loginProc)))

		r.Get("/request-password-reset", h.handle(h.anony(h.requestPasswordResetForm)))
		r.Post("/request-password-reset", h.handle(h.anony(h.requestPasswordResetProc)))
		r.Get("/password-reset-requested", h.handle(h.anony(h.passwordResetRequested)))
		r.Get("/reset-password", h.handle(h.resetPasswordForm))
		r.Post("/reset-password", h.handle(h.auth(h.resetPasswordProc)))

		r.Get("/change-password", h.handle(h.auth(h.changePasswordForm)))
		r.Post("/change-password", h.handle(h.auth(h.changePasswordProc)))

		r.Route("/oauth2", func(r chi.Router) {
			r.HandleFunc("/authorize", h.handle(h.oauth2Authorize))
			r.Get("/authorize-client", h.handle(h.auth(h.oauth2AuthorizeClient)))
			r.Post("/authorize-client", h.handle(h.auth(h.oauth2AuthorizeClientProc)))

		})

		r.Route("/federated/{provider}", func(r chi.Router) {
			r.Get("/", h.federatedInit)
			r.Get("/callback", h.federatedCallback)
		})
	})

	r.Handle("/assets/*", http.StripPrefix(
		"/auth/assets",
		http.FileServer(http.Dir("auth/assets")),
	))

	// Excluded from csrf
	r.HandleFunc("/oauth2/token", h.handle(h.oauth2Token))
	r.HandleFunc("/oauth2/info", h.oauth2Info)
}

// redirects anonymous users to login
func (h *AuthHandlers) auth(fn handlerFn) handlerFn {
	return func(p *request.AuthReq) error {
		if p.User == nil {
			p.RedirectTo = templates.GetLinks().Login
			return nil
		} else {
			return fn(p)
		}
	}
}

// redirects authenticated users to profile
func (h *AuthHandlers) anony(fn handlerFn) handlerFn {
	return func(p *request.AuthReq) error {
		if p.User != nil {
			p.RedirectTo = templates.GetLinks().Profile
			return nil
		} else {
			return fn(p)
		}
	}
}

// Stores user & roles
//
// We need to store roles separately because they do not get serialized alongside with user
// due to unexported field
func (h *AuthHandlers) storeUserToSession(req *request.AuthReq, u *types.User) {
	session.SetUser(req.Session, u)
	session.SetRoleMemberships(req.Session, u.Roles())
	// @todo refresh token!
}

// handles auth request and prepares request struct with request, session and response helper
func (h *AuthHandlers) handle(fn handlerFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Log.Debug(
			"handling request",
			zap.String("url", r.RequestURI),
			zap.String("method", r.Method),
		)

		var (
			req = &request.AuthReq{
				Response:   w,
				Request:    r,
				Data:       make(map[string]interface{}),
				NewAlerts:  make([]request.Alert, 0),
				PrevAlerts: make([]request.Alert, 0),
				Session:    h.SessionManager.Get(r),
			}
		)

		err := func() (err error) {
			if err = r.ParseForm(); err != nil {
				return
			}

			req.User = session.GetUser(req.Session)
			req.Data["user"] = req.User

			// Alerts show for 1 session only!
			req.PrevAlerts = req.PopAlerts()
			if err = fn(req); err != nil {
				return
			}

			if len(req.NewAlerts) > 0 {
				req.SetAlerts(req.NewAlerts...)
			}

			h.SessionManager.Save(w, r)

			if req.Status == 0 {
				switch {
				case req.RedirectTo != "":
					req.Status = http.StatusSeeOther
					req.Template = ""
				case req.Template != "":
					req.Status = http.StatusOK
				default:
					req.Status = http.StatusInternalServerError
					req.Template = tplInternalError
				}
			}

			return nil
		}()

		if err == nil {

			if req.Status >= 300 && req.Status < 400 {
				// redirect, nothing special to handle
				http.Redirect(w, r, req.RedirectTo, req.Status)
				return
			}

			if req.Status > 0 {
				// in cases when something else already wrote the status
				w.WriteHeader(req.Status)
			}
		}

		if err == nil && req.Template != "" {
			err = h.Templates.ExecuteTemplate(w, req.Template, h.enrichTplData(req))
			h.Log.Debug("template executed", zap.String("name", req.Template), zap.Error(err))

		}

		if err != nil {
			err = h.Templates.ExecuteTemplate(w, tplInternalError, map[string]interface{}{
				"error": err,
			})
			h.Log.Debug("request handled", zap.Error(err))

			if err == nil {
				return
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// Add alerts, settings, providers, csrf token
func (h *AuthHandlers) enrichTplData(req *request.AuthReq) interface{} {
	d := req.Data

	d[csrf.TemplateTag] = csrf.TemplateField(req.Request)

	// In case we did not redirect, join previous alerts with new ones
	d["alerts"] = append(req.PrevAlerts, req.NewAlerts...)

	dSettings := *h.Settings
	dSettings.Providers = nil
	d["settings"] = dSettings

	providers := h.AuthService.GetProviders()
	sort.Sort(providers)

	var pp = make([]provider, 0, len(providers))
	for i := range providers {
		if !providers[i].Enabled {
			continue
		}

		p := provider{
			Label:  providers[i].Label,
			Handle: providers[i].Handle,
			Icon:   providers[i].Handle,
		}

		if strings.HasPrefix(p.Icon, federated.OIDC_PROVIDER_PREFIX) {
			p.Icon = "key"
		}

		pp = append(pp, p)
	}

	d["providers"] = pp

	return d
}
