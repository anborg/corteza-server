package options

import (
	"crypto/md5"
	"fmt"
)

func (o *AuthOpt) Defaults() {
	if o.Secret == "" {
		// use a couple of env values + salt for default secret
		o.Secret = getSecretFromEnv("jwt secret")
	}

	if o.CsrfSecret == "" {
		// use a couple of env values + salt for default secret
		o.CsrfSecret = getSecretFromEnv("csrf secret")
	}

	if o.FederatedCookieSecret == "" {
		// use a couple of env values + salt for default secret
		o.FederatedCookieSecret = getSecretFromEnv("federated cookie secret")
	}
}

func getSecretFromEnv(salt string) string {
	gen := salt
	// generate default secrets from virtualhost/hostname and DB_DSN value.
	// this will keep the secret the same through restarts
	gen += EnvString("DB_DSN", "memory")
	// pick one of the env that holds hostname
	gen += EnvString("HOSTNAME", "localhost")

	return fmt.Sprintf("%x", md5.Sum([]byte(gen)))
}
