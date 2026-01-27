package urls

import (
	"github.com/SirNacou/refract/api/internal/domain"
	listurls "github.com/SirNacou/refract/api/internal/features/urls/list_urls"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/repository"
	"github.com/danielgtaylor/huma/v2"
)

type Module struct {
	repo domain.URLRepository
}

func NewModule(db *persistence.DB) *Module {
	repo := repository.NewPostgresURLRepository(db.Querier)

	return &Module{repo}
}

func (m *Module) RegisterRoutes(api huma.API) error {

	grp := huma.NewGroup(api, "/urls")

	huma.Register(grp, huma.Operation{
		OperationID: "list-urls",
		Method:      "GET",
		Path:        "/",
	}, listurls.NewHandler(listurls.NewQueryHandler(m.repo)).Handle)

	return nil
}
