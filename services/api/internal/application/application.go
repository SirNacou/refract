package application

import (
	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/application/queries"
	"github.com/SirNacou/refract/services/api/internal/application/service"
	"github.com/SirNacou/refract/services/api/internal/domain"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
)

// Application holds all application services (commands and queries)
type Application struct {
	Commands Commands
	Queries  Queries
}

// Commands holds all command handlers
type Commands struct {
	CreateURL *commands.CreateURLHandler
}

// Queries holds all query handlers
type Queries struct {
	GetURL *queries.GetURLByShortCodeHandler
}

// NewApplication creates a new Application instance with all dependencies
func NewApplication(
	generator idgen.IDGenerator,
	sb service.SafeBrowsing,
	store domain.Store,
	cache service.Cache,
) *Application {
	return &Application{
		Commands: Commands{
			CreateURL: commands.NewCreateURLHandler(generator, sb, store, cache),
		},
		Queries: Queries{
			GetURL: queries.NewGetURLByShortCodeHandler(store, cache),
		},
	}
}
