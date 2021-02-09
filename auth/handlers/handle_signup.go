package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/system/service"
	"github.com/cortezaproject/corteza-server/system/types"
	"go.uber.org/zap"
)

const (
	tplSignup                   = "signup.html.tpl"
	tplPendingEmailConfirmation = "pending-email-confirmation.html.tpl"
)

func (h *AuthHandlers) signupForm(req *request.AuthReq) error {
	req.Template = tplSignup
	req.Data["form"] = req.GetKV()
	return nil
}

func (h *AuthHandlers) signupProc(req *request.AuthReq) error {
	req.RedirectTo = templates.GetLinks().Signup
	req.SetKV(nil)

	newUser := &types.User{
		Email:  req.Request.PostFormValue("email"),
		Handle: req.Request.PostFormValue("handle"),
		Name:   req.Request.PostFormValue("name"),
	}

	newUser, err := h.AuthService.InternalSignUp(
		req.Context(),
		newUser,
		req.Request.PostFormValue("password"),
	)

	if err == nil {
		if newUser.EmailConfirmed {
			req.NewAlerts = append(req.NewAlerts, request.Alert{
				Type: "primary",
				Text: "Sign-up successful.",
			})

			h.Log.Info("signup successful")
			req.RedirectTo = templates.GetLinks().Profile
			h.storeUserToSession(req, newUser)
		} else {
			req.RedirectTo = templates.GetLinks().PendingEmailConfirmation
		}

		return nil
	}

	switch {
	case
		service.AuthErrInternalSignupDisabledByConfig().Is(err),
		service.AuthErrInvalidEmailFormat().Is(err),
		service.AuthErrInvalidHandle().Is(err),
		service.AuthErrPasswordNotSecure().Is(err),
		service.AuthErrInvalidCredentials().Is(err):
		req.SetKV(map[string]string{
			"error":  err.Error(),
			"email":  newUser.Email,
			"handle": newUser.Handle,
			"name":   newUser.Name,
		})

		h.Log.Warn("handled error", zap.Error(err))
		return nil

	default:
		h.Log.Error("unhandled error", zap.Error(err))
		return err
	}
}

func (h *AuthHandlers) pendingEmailConfirmation(req *request.AuthReq) error {
	req.Template = tplPendingEmailConfirmation

	if _, has := req.Request.URL.Query()["resend"]; has {
		err := h.AuthService.SendEmailAddressConfirmationToken(req.Context(), req.User)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *AuthHandlers) confirmEmail(req *request.AuthReq) (err error) {
	if token := req.Request.URL.Query().Get("token"); len(token) > 0 {
		req.User, err = h.AuthService.ValidateEmailConfirmationToken(req.Context(), token)
		if err == nil {
			// redirect back to self (but without token and with user in session
			h.Log.Debug("valid email confirmation token found, redirecting to profile")
			req.RedirectTo = templates.GetLinks().Profile
			h.storeUserToSession(req, req.User)
			return nil
		}
	}

	h.Log.Warn("invalid email confirmation token used", zap.Error(err))

	// redirect to the right page
	// not doing this here and relying on handler on subseq. request
	// will cause alerts to be removed
	if req.User == nil {
		req.RedirectTo = templates.GetLinks().Login
	} else {
		req.RedirectTo = templates.GetLinks().Profile
	}

	req.NewAlerts = append(req.NewAlerts, request.Alert{
		Type: "warning",
		Text: "Invalid or expired email confirmation token, please resend confirmation request.",
	})

	return nil
}
