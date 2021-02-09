package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/request"
)

const (
	tplProfile = "profile.html.tpl"
)

func (h *AuthHandlers) profileView(areq *request.AuthReq) error {
	areq.Template = tplProfile

	areq.Data["emailConfirmationRequired"] = !areq.User.EmailConfirmed || !h.AuthService.EmailConfirmationRequired()
	return nil
}
