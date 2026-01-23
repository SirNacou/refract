package auth

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type Auth = authorization.Authorizer[*oauth.IntrospectionContext]

func NewAuth(ctx context.Context, domain, keyPath string) (*Auth, error) {
	auth, err := authorization.New(ctx,
		zitadel.New(domain, zitadel.WithInsecureSkipVerifyTLS()),
		oauth.DefaultAuthorization(keyPath))
	return auth, err
}
