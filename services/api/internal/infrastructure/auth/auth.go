package auth

import (
	"github.com/SirNacou/refract/services/api/internal/config"
	"github.com/appwrite/sdk-for-go/account"
	"github.com/appwrite/sdk-for-go/appwrite"
)

type Auth struct {
	endpoint  string
	apiKey    string
	projectID string
}

func NewAuth(cfg *config.AppwriteConfig) *Auth {
	return &Auth{
		endpoint:  cfg.Endpoint,
		apiKey:    cfg.APIKey,
		projectID: cfg.ProjectID,
	}
}

func (a *Auth) WithJWT(jwt string) *account.Account {
	clt := appwrite.NewClient(
		appwrite.WithEndpoint(a.endpoint),
		appwrite.WithProject(a.projectID),
		appwrite.WithKey(a.apiKey),
		appwrite.WithJWT(jwt),
	)
	return appwrite.NewAccount(clt)
}
