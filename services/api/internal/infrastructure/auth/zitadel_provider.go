package auth

import (
	"context"

	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

type ZitadelProvider struct {
	AuthZ *authorization.Authorizer[*oauth.IntrospectionContext]
}

func NewZitadelProvider(ctx context.Context, domain string, clientID string) (*ZitadelProvider, error) {
	authZ, err := authorization.New(ctx,
		zitadel.New(domain),
		oauth.DefaultJWTAuthorization(clientID))
	if err != nil {
		return nil, err
	}

	return &ZitadelProvider{AuthZ: authZ}, nil
}
