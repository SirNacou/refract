package urls

import (
	"net/http"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/domain"
	listurls "github.com/SirNacou/refract/api/internal/features/urls/list_urls"
	shortenurl "github.com/SirNacou/refract/api/internal/features/urls/shorten_url"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/repository"
	"github.com/danielgtaylor/huma/v2"
)

type Module struct {
	repo domain.URLRepository
	cfg  *config.Config
}

func NewModule(db *persistence.DB, cfg *config.Config) *Module {
	repo := repository.NewPostgresURLRepository(db.Querier)

	return &Module{repo, cfg}
}

func (m *Module) RegisterRoutes(api huma.API) error {

	grp := huma.NewGroup(api, "/urls")

	huma.Register(grp, huma.Operation{
		OperationID: "list-urls",
		Method:      http.MethodGet,
		Path:        "/",
	}, listurls.NewHandler(listurls.NewQueryHandler(m.repo)).Handle)

	huma.Register(grp, huma.Operation{
		OperationID: "shorten-url",
		Method:      http.MethodPost,
		Path:        "/",
	}, shortenurl.NewHandler(shortenurl.NewCommandHandler(m.repo, m.cfg.DefaultDomain)).Handle)

	return nil
}
