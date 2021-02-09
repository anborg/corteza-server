package auth

import (
	"github.com/cortezaproject/corteza-server/pkg/api"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	token struct {
		// Expiration time in minutes
		expiry    int64
		tokenAuth *jwtauth.JWTAuth
	}
)

var (
	DefaultJwtHandler TokenHandler
)

func SetupDefault(secret string, expiry int) {
	// Use JWT secret for hmac signer for now
	DefaultSigner = HmacSigner(secret)
	DefaultJwtHandler, _ = JWT(secret, int64(expiry))

}

func JWT(secret string, expiry int64) (tkn *token, err error) {
	if len(secret) == 0 {
		return nil, errors.New("JWT secret missing")
	}

	tkn = &token{
		expiry:    expiry,
		tokenAuth: jwtauth.New(jwt.SigningMethodHS512.Alg(), []byte(secret), nil),
	}

	return tkn, nil
}

// Verifies JWT and stores it into context
func (t *token) HttpVerifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(t.tokenAuth)
}

func (t *token) Decode(ts string) (Identifiable, error) {
	var (
		decoded, err = t.tokenAuth.Decode(ts)

		rr     []uint64
		userID uint64
	)
	if err != nil {
		return nil, err
	}

	if err = decoded.Claims.Valid(); err != nil {
		return nil, err
	}

	if c, ok := decoded.Claims.(jwt.MapClaims); ok {
		userID, _ = strconv.ParseUint(c["userID"].(string), 10, 64)

		if memberOf, ok := c["memberOf"].(string); ok {
			for _, str := range strings.Split(memberOf, " ") {
				if id, _ := strconv.ParseUint(str, 10, 64); id > 0 {
					rr = append(rr, id)
				}
			}
		}
	}

	if userID > 0 {
		return NewIdentity(userID, rr...), nil
	}

	return nil, errors.New("invalid claims")

}

func (t *token) Encode(identity Identifiable) string {
	claims := jwt.MapClaims{
		"userID": strconv.FormatUint(identity.Identity(), 10),
		"exp":    time.Now().Add(time.Duration(t.expiry) * time.Minute).Unix(),
	}

	if rr := identity.Roles(); len(rr) > 0 {
		var memberOf string
		for _, r := range identity.Roles() {
			memberOf = memberOf + " " + strconv.FormatUint(r, 10)
		}

		claims["memberOf"] = memberOf[1:] // trim leading space
	}

	_, jwt, _ := t.tokenAuth.Encode(claims)
	return jwt
}

// HttpAuthenticator converts JWT claims into Identity and stores it into context
func (t *token) HttpAuthenticator() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			jwt, claims, err := jwtauth.FromContext(r.Context())

			// When token is present, expect no errors and valid claims!
			if jwt != nil {
				if err != nil {
					// But if token is present, the shouldn't be an error
					api.Send(w, r, err)
					return
				}

				// decodes string with space delimited set of uint64
				// (user id + set of roles user is member of)
				identity := &Identity{}
				if sub, has := claims["sub"]; has {
					if str, is := sub.(string); is {
						parts := strings.Split(str, " ")
						identity.memberOf = make([]uint64, len(parts)-1)
						for p := range parts {
							id, _ := strconv.ParseUint(parts[p], 10, 64)
							if p == 0 {
								identity.id = id
							} else {
								identity.memberOf[p-1] = id
							}
						}
					}
				}

				r = r.WithContext(SetIdentityToContext(r.Context(), identity))
			}

			next.ServeHTTP(w, r)
		})
	}
}
