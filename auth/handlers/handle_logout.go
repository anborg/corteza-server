package handlers

import (
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/markbates/goth/gothic"
)

const (
	tplLogout = "logout.html.tpl"
)

func (h *AuthHandlers) logoutProc(req *request.AuthReq) (err error) {
	req.Session.Options.MaxAge = -1
	req.PermSession.Options.MaxAge = -1

	if err = req.Session.Save(req.Request, req.Response); err != nil {
		return
	}
	if err = req.PermSession.Save(req.Request, req.Response); err != nil {
		return
	}

	if err = gothic.Logout(req.Response, req.Request); err != nil {
		return
	}

	h.Log.Info("logout successful")
	req.Template = tplLogout
	return
}
