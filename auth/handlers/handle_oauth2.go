package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cortezaproject/corteza-server/auth/handlers/templates"
	"github.com/cortezaproject/corteza-server/auth/oauth2"
	"github.com/cortezaproject/corteza-server/auth/request"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/cortezaproject/corteza-server/pkg/errors"
	"github.com/cortezaproject/corteza-server/system/types"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const (
	tplOAuth2AuthorizeClient = "oauth2-authorize-client.html.tpl"
)

// oauth2 flow authorize step
//
// OA2 server internals first run user check (see SetUserAuthorizationHandler lambda)
// to ensure user is authenticated;
func (h AuthHandlers) oauth2Authorize(req *request.AuthReq) (err error) {
	if form := session.GetOAuth2AuthParams(req.Session); form != nil {
		req.Request.Form = form
		h.Log.Debug("restarting oauth2 authorization flow", zap.Any("params", req.Request.Form))
	} else {
		h.Log.Debug("starting new oauth2 authorization flow", zap.Any("params", req.Request.Form))

	}

	session.SetOauth2AuthParams(req.Session, nil)

	var (
		ctx    context.Context
		client *types.AuthClient
	)

	if client, ctx, err = h.loadRequestedClient(req); err != nil {
		return err
	}

	if client != nil {
		session.SetOauth2Client(req.Session, client)
	}

	// set to -1 to make sure that wrapping request handler
	// does not send status code!
	req.Status = -1

	// handle authorize request with extended context that now holds client!
	err = h.OAuth2.HandleAuthorizeRequest(req.Response, req.Request.Clone(ctx))
	if err != nil {
		req.Status = http.StatusInternalServerError
		req.Template = tplInternalError
		req.Data["error"] = err
	}

	return nil
}

func (h AuthHandlers) oauth2AuthorizeClient(req *request.AuthReq) (err error) {
	client := session.GetOauth2Client(req.Session)

	if client == nil {
		return fmt.Errorf("flow broken; client missing")
	}

	if client.Trusted {
		// Client is trusted, no need to show this screen
		// move forward and authorize oauth2 request
		req.RedirectTo = templates.GetLinks().OAuth2Authorize
	}

	h.Log.Debug("showing oauth2 client auth form")

	req.Template = tplOAuth2AuthorizeClient
	req.Data["client"] = &types.AuthClient{
		Name:          "Dummy name",
		RedirectURI:   "",
		DenyAccessURI: "",
	}
	return nil
}

func (h AuthHandlers) oauth2AuthorizeClientProc(req *request.AuthReq) (err error) {
	// handle deny client action from authorize-client form
	//
	// This occurs when user pressed "DENY" button on authorize-client form
	// Remove all and redirect to profile
	//
	session.SetOauth2Client(req.Session, nil)
	if _, allow := req.Request.Form["allow"]; allow {
		session.SetOauth2ClientAuthorized(req.Session, true)
		req.RedirectTo = templates.GetLinks().OAuth2Authorize
		return
	}

	session.SetOauth2AuthParams(req.Session, nil)
	req.RedirectTo = templates.GetLinks().Profile
	req.NewAlerts = append(req.NewAlerts, request.Alert{
		Type: "primary",
		Text: "Access for client denied",
	})
	return
}

func (h AuthHandlers) oauth2Token(req *request.AuthReq) (err error) {
	// Cleanup
	session.SetOauth2ClientAuthorized(req.Session, false)

	req.Status = -1

	_, ctx, err := h.loadRequestedClient(req)

	if err != nil {
		// handle token request with extended context that now holds client!
		err = h.OAuth2.HandleTokenRequest(req.Response, req.Request.Clone(ctx))
	}

	return
}

func (h AuthHandlers) oauth2Info(w http.ResponseWriter, r *http.Request) {
	token, err := h.OAuth2.ValidationBearerToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := map[string]interface{}{
		"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
		"client_id":  token.GetClientID(),
		"user_id":    token.GetUserID(),
	}

	json.NewEncoder(w).Encode(data)
}

// loads client from the request params and verifies other request params against client settings
func (h AuthHandlers) loadRequestedClient(req *request.AuthReq) (client *types.AuthClient, ctx context.Context, err error) {
	return client, ctx, func() (err error) {
		ctx = req.Context()

		var (
			clientID uint64
		)

		if _, pExists := req.Request.Form["client_id"]; !pExists {
			return
		}

		h.Log.Debug("loading client", zap.String("info", req.Request.Form.Get("client_id")))

		if clientID, err = strconv.ParseUint(req.Request.Form.Get("client_id"), 10, 64); err != nil {
			return errors.InvalidData("failed to parse client ID from params: %v", err)

		} else if clientID == 0 {
			return errors.InvalidData("invalid client ID")
		}

		if client = session.GetOauth2Client(req.Session); client != nil {
			h.Log.Debug("client loaded from session", zap.Any("info", client))

			// ensure that session holds the right client and
			// not some leftover from a previous flow
			if client.ID == clientID {
				ctx = context.WithValue(ctx, &oauth2.ContextClientStore{}, client)
				return
			}

			h.Log.Debug("stale client found in session")

			// cleanup leftovers
			client = nil
			// session.SetOauth2Client(ac.Session, nil) // @todo fix
		}

		client, err = h.ClientService.LookupByID(req.Context(), clientID)
		if err != nil {
			return fmt.Errorf("invalid client: %w", err)
		}

		// add client to context so we can reach it
		// from client store via context.Value() fn
		ctx = context.WithValue(ctx, &oauth2.ContextClientStore{}, client)

		h.Log.Debug("client loaded from store", zap.Any("info", client))
		return
	}()
}

func (h AuthHandlers) verifyClient(client *types.AuthClient) (err error) {
	switch {
	case !client.Enabled:
		return fmt.Errorf("client disabled")
	case client.ExpiresAt != nil && client.ExpiresAt.After(time.Now()):
		return fmt.Errorf("client expired")
	case client.ValidFrom != nil && client.ValidFrom.Before(time.Now()):
		return fmt.Errorf("client not yet valid")
	}

	return nil
}
