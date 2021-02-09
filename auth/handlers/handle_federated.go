package handlers

import (
	"context"
	"fmt"
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/cortezaproject/corteza-server/pkg/api"
	"github.com/cortezaproject/corteza-server/pkg/logger"
	"github.com/cortezaproject/corteza-server/system/types"
	"github.com/go-chi/chi"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"strings"
)

func copyProviderToContext(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), "provider", chi.URLParam(r, "provider")))
}

func (h AuthHandlers) federatedInit(w http.ResponseWriter, r *http.Request) {
	r = copyProviderToContext(r)
	h.Log.Info("starting federated authentication flow")

	gothic.BeginAuthHandler(w, r)
}

func (h AuthHandlers) federatedCallback(w http.ResponseWriter, r *http.Request) {
	r = copyProviderToContext(r)
	h.Log.Info("federated authentication callback")

	if user, err := gothic.CompleteUserAuth(w, r); err != nil {
		h.log(r.Context(), zap.Error(err)).Error("failed to complete user auth")
		h.handleFailedFederatedAuth(w, r, err)
	} else {
		h.handleSuccessfulFederatedAuth(w, r, user)
	}
}

func (h AuthHandlers) log(ctx context.Context, fields ...zapcore.Field) *zap.Logger {
	return logger.ContextValue(ctx).Named("external-auth").With(fields...)
}

// Handles authentication via external auth providers of
// unknown an user + appending authentication on external providers
// to a current user
func (h AuthHandlers) handleSuccessfulFederatedAuth(w http.ResponseWriter, r *http.Request, cred goth.User) {
	var (
		authUser *types.User
		err      error
		ctx      = r.Context()
	)

	h.log(ctx, zap.String("provider", cred.Provider)).Info("external login successful")

	// Try to login/sign-up external user
	if authUser, err = h.AuthService.External(ctx, cred); err != nil {
		api.Send(w, r, err)
		return
	}

	h.handle(func(req *request.AuthReq) error {
		h.storeUserToSession(req, authUser)

		if session.GetOAuth2AuthParams(req.Session) != nil {
			// If we have oauth2 auth params stored in the session,
			// try and continue with the oauth2 flow
			req.RedirectTo = templates.GetLinks().OAuth2Authorize
			return nil
		}

		req.RedirectTo = templates.GetLinks().Profile
		return nil
	})(w, r)
}

func (h AuthHandlers) handleFailedFederatedAuth(w http.ResponseWriter, r *http.Request, err error) {
	//provider := chi.URLParam(r, "provider")

	if strings.Contains(err.Error(), "Error processing your OAuth request: Invalid oauth_verifier parameter") {
		// Just take user through the same loop again
		w.Header().Set("Location", templates.GetLinks().Profile)
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "SSO Error: %v", err.Error())
	w.WriteHeader(http.StatusOK)
}
