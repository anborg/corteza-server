package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/system/service"
	"go.uber.org/zap"
)

const (
	tplChangePassword = "change-password.html.tpl"
)

func (h *AuthHandlers) changePasswordForm(req *request.AuthReq) error {
	h.Log.Debug("showing password change form")
	req.Template = tplChangePassword
	req.Data["form"] = req.GetKV()
	return nil
}

func (h *AuthHandlers) changePasswordProc(req *request.AuthReq) (err error) {
	err = h.AuthService.ChangePassword(
		req.Context(),
		req.User.ID,
		req.Request.PostFormValue("oldPassword"),
		req.Request.PostFormValue("newPassword"),
	)

	if err == nil {
		req.NewAlerts = append(req.NewAlerts, request.Alert{
			Type: "primary",
			Text: "Password successfully changed.",
		})

		req.RedirectTo = templates.GetLinks().Profile
		return nil
	}

	switch {
	case service.AuthErrInteralLoginDisabledByConfig().Is(err),
		service.AuthErrPasswordNotSecure().Is(err),
		service.AuthErrPasswordChangeFailedForUnknownUser().Is(err),
		service.AuthErrPasswodResetFailedOldPasswordCheckFailed().Is(err):
		req.SetKV(map[string]string{
			"error": err.Error(),
		})
		req.RedirectTo = templates.GetLinks().ChangePassword

		h.Log.Warn("handled error", zap.Error(err))
		return nil

	default:
		h.Log.Error("unhandled error", zap.Error(err))
		return err
	}
}
