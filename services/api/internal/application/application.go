package application

import (
	"github.com/SirNacou/refract/services/api/internal/application/commands"
	"github.com/SirNacou/refract/services/api/internal/domain/url"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/idgen"
	"github.com/SirNacou/refract/services/api/internal/infrastructure/safebrowsing"
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
	// Future query handlers will go here
}

// NewApplication creates a new Application instance with all dependencies
func NewApplication(
	generator *idgen.SnowflakeGenerator,
	sb *safebrowsing.SafeBrowsing,
	urlRepo url.URLRepository,
) *Application {
	return &Application{
		Commands: Commands{
			CreateURL: commands.NewCreateURLHandler(generator, sb, urlRepo),
		},
		Queries: Queries{
			// Initialize queries here when added
		},
	}
}
