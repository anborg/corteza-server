package oauth2

import (
	"fmt"
	"github.com/cortezaproject/corteza-server/pkg/logger"
	"github.com/cortezaproject/corteza-server/pkg/options"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"go.uber.org/zap"
	"strings"
)

const (
	RedirectUriSeparator = " "
)

func NewManager(opt options.AuthOpt, cs oauth2.ClientStore, ts oauth2.TokenStore) *manage.Manager {
	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// token store
	manager.MapTokenStorage(ts)

	// generate jwt access token
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte(opt.Secret), jwt.SigningMethodHS512))
	manager.MapClientStorage(cs)

	manager.SetValidateURIHandler(func(baseURI, redirectURI string) (err error) {
		for _, baseURI = range strings.Split(baseURI, RedirectUriSeparator) {
			if err = manage.DefaultValidateURI(baseURI, redirectURI); err != nil {
				return
			}
		}

		return nil
	})

	return manager
}

func NewServer(manager *manage.Manager, uah server.UserAuthorizationHandler) *server.Server {
	srv := server.NewServer(&server.Config{
		TokenType: "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{
			oauth2.Code,
			oauth2.Token,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
			oauth2.PasswordCredentials,
			oauth2.ClientCredentials,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengePlain,
			oauth2.CodeChallengeS256,
		},
	}, manager)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		return errors.NewResponse(err, 500)
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		logger.Default().
			WithOptions(zap.AddStacktrace(zap.PanicLevel)).
			Error(re.Description)
	})

	srv.SetPasswordAuthorizationHandler(func(username, password string) (userID string, err error) {
		return "0", fmt.Errorf("pending implementation")
	})

	// Called after oauth2 authorization request is validated
	// We'll try to get valid user out of the session or redirect user to login page
	srv.SetUserAuthorizationHandler(uah)

	return srv
}
