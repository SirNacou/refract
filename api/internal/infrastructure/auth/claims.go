package auth

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	claimsContextKey = "claims"
)

type Claims struct {
	jwt.Token
}

func GetClaimsFromContext(ctx context.Context) *Claims {
	token, ok := ctx.Value(claimsContextKey).(Claims)
	if !ok {
		return nil
	}
	return &Claims{Token: token}
}

func SetClaimsToContext(ctx context.Context, token *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, token)
}

func (c *Claims) GetUserID() (string, error) {
	sub, ok := c.Subject()
	if !ok {
		return "", errors.New("User ID not found")
	}

	return sub, nil
}
