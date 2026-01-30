package auth

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v3/jwt"
)

type contextKey string

const (
	claimsContextKey = contextKey("claims")
)

type Claims struct {
	jwt.Token
}

func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	token, ok := ctx.Value(claimsContextKey).(*Claims)
	if !ok {
		return nil, errors.New("user not found")
	}
	return token, nil
}

func SetClaimsToContext(ctx context.Context, token *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, token)
}

func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, err := GetClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	return claims.GetUserID()
}

func (c *Claims) GetUserID() (string, error) {
	sub, ok := c.Subject()
	if !ok {
		return "", errors.New("user ID not found")
	}

	return sub, nil
}
