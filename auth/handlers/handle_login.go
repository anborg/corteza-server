package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/cortezaproject/corteza-server/system/service"
	"github.com/cortezaproject/corteza-server/system/types"
	"go.uber.org/zap"
)

type (
	provider struct {
		Label, Handle, Icon string
	}
)

const (
	tplLogin = "login.html.tpl"
)

func (h *AuthHandlers) loginForm(req *request.AuthReq) error {
	req.Template = tplLogin
	req.Data["form"] = req.GetKV()
	return nil
}

func (h *AuthHandlers) loginProc(req *request.AuthReq) (err error) {
	req.RedirectTo = templates.GetLinks().Login
	req.SetKV(nil)

	var (
		authUser *types.User
		email    = req.Request.PostFormValue("email")
	)

	authUser, err = h.AuthService.InternalLogin(
		req.Context(),
		email,
		req.Request.PostFormValue("password"),
	)

	if err == nil {
		req.NewAlerts = append(req.NewAlerts, request.Alert{
			Type: "primary",
			Text: "You are now logged-in",
		})

		h.Log.Info("login successful")
		h.storeUserToSession(req, authUser)

		if len(req.Request.PostFormValue("keep-session")) > 0 {
			session.SetPerm(req.Session, h.Opt.SessionPermLifetime)
		}

		if session.GetOAuth2AuthParams(req.Session) == nil {
			// Not in the OAuth2 flow, go to profile
			req.RedirectTo = templates.GetLinks().Profile
		} else {
			req.RedirectTo = templates.GetLinks().OAuth2AuthorizeClient
		}

		return nil
	}

	switch {
	case service.AuthErrInteralLoginDisabledByConfig().Is(err),
		service.AuthErrInvalidEmailFormat().Is(err),
		service.AuthErrInvalidCredentials().Is(err),
		service.AuthErrCredentialsLinkedToInvalidUser().Is(err):
		req.SetKV(map[string]string{
			"error": err.Error(),
			"email": email,
		})

		h.Log.Warn("handled error", zap.Error(err))
		return nil

	default:
		h.Log.Error("unhandled error", zap.Error(err))
		return err
	}
}
