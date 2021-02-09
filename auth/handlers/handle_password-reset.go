package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/pkg/errors"
	"github.com/cortezaproject/corteza-server/system/service"
	"go.uber.org/zap"
)

const (
	tplRequestPasswordReset   = "request-password-reset.html.tpl"
	tplPasswordResetRequested = "password-reset-requested.html.tpl"
	tplResetPassword          = "reset-password.html.tpl"
)

func (h *AuthHandlers) requestPasswordResetForm(req *request.AuthReq) error {
	h.Log.Debug("showing request password reset form")
	req.Template = tplRequestPasswordReset
	req.Data["form"] = req.GetKV()
	return nil
}

func (h *AuthHandlers) requestPasswordResetProc(req *request.AuthReq) (err error) {
	h.Log.Debug("processing password change request")

	email := req.Request.PostFormValue("email")
	err = h.AuthService.SendPasswordResetToken(req.Context(), email)

	if err == nil || errors.IsNotFound(err) {
		req.RedirectTo = templates.GetLinks().PasswordResetRequested
		return nil
	}

	switch {
	case service.AuthErrPasswordResetDisabledByConfig().Is(err):
		req.SetKV(map[string]string{
			"error": err.Error(),
			"email": email,
		})
		req.RedirectTo = templates.GetLinks().RequestPasswordReset

		h.Log.Warn("handled error", zap.Error(err))
		return nil

	default:
		h.Log.Error("unhandled error", zap.Error(err))
		return err
	}
}

func (h *AuthHandlers) passwordResetRequested(req *request.AuthReq) error {
	req.Template = tplPasswordResetRequested
	return nil
}

func (h *AuthHandlers) resetPasswordForm(req *request.AuthReq) (err error) {
	h.Log.Debug("password reset form")

	req.Template = tplResetPassword

	if req.User == nil {
		// user not set, expecting valid token in URL
		if token := req.Request.URL.Query().Get("token"); len(token) > 0 {
			req.User, err = h.AuthService.ValidatePasswordResetToken(req.Context(), token)
			if err == nil {
				// redirect back to self (but without token and with user in session
				h.Log.Debug("valid password reset token found, refreshing page with stored user")
				req.RedirectTo = templates.GetLinks().ResetPassword
				h.storeUserToSession(req, req.User)
				return nil
			}
		}

		h.Log.Warn("invalid password reset token used", zap.Error(err))
		req.RedirectTo = templates.GetLinks().RequestPasswordReset
		req.NewAlerts = append(req.NewAlerts, request.Alert{
			Type: "warning",
			Text: "Invalid or expired password reset token, please repeat password reset request.",
		})
	}

	req.Data["form"] = req.GetKV()
	return nil
}

func (h *AuthHandlers) resetPasswordProc(req *request.AuthReq) (err error) {
	h.Log.Debug("password reset proc")

	err = h.AuthService.SetPassword(req.Context(), req.User.ID, req.Request.PostFormValue("password"))

	if err == nil {
		req.NewAlerts = append(req.NewAlerts, request.Alert{
			Type: "primary",
			Text: "Password successfully reset.",
		})

		req.RedirectTo = templates.GetLinks().Profile
		return nil
	}

	switch {
	case service.AuthErrPasswordResetDisabledByConfig().Is(err):
		req.SetKV(map[string]string{
			"error": err.Error(),
		})
		req.RedirectTo = templates.GetLinks().RequestPasswordReset

		h.Log.Warn("handled error", zap.Error(err))
		return nil

	default:
		h.Log.Error("unhandled error", zap.Error(err))
		return err
	}
}
