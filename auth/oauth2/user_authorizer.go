package oauth2

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/auth/session"
	"github.com/go-oauth2/oauth2/v4/server"
	"net/http"
)

func NewUserAuthorizer(sm *session.Manager, loginURL, clientAuthURL string) server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (identity string, err error) {
		var (
			ses  = sm.Get(r)
			user = session.GetUser(ses)
		)

		// temporary break oauth2 flow by redirecting to
		// login form and ask user to authenticate
		session.SetOauth2AuthParams(ses, r.Form)

		// @todo harden security by enforcing login
		//       for each new authorization flow
		if user == nil {
			// user is currently not logged-in;
			http.Redirect(w, r, loginURL, http.StatusSeeOther)
			return
		} else if session.IsOauth2ClientAuthorized(ses) {
			// user logged in but we need to re-authenticate the client
			http.Redirect(w, r, clientAuthURL, http.StatusSeeOther)
			return
		}

		// User authenticated, client authorized!
		// remove authorization values from session
		session.SetOauth2AuthParams(ses, nil)
		session.SetOauth2Client(ses, nil)
		session.SetOauth2ClientAuthorized(ses, false)

		// @todo extra checks if user is valid!
		//   -- invalid user means:
		//   -- unauthorized email

		// Pack user's ID and IDs of all roles they are member of
		// into space delimited string
		//
		// Main reason to do this is to simplify JWT claims encoding;
		// we do not have  access to user's membership info from there)
		identity = fmt.Sprintf("%d", user.ID)
		for _, roleID := range session.GetRoleMemberships(ses) {
			identity += fmt.Sprintf(" %d", roleID)
		}

		return
	}
}
