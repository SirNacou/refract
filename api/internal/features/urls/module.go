package urls

import (
	listurls "github.com/SirNacou/refract/api/internal/features/urls/list_urls"
	"github.com/danielgtaylor/huma/v2"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) RegisterRoutes(api huma.API) error {
	grp := huma.NewGroup(api, "/urls")

	huma.Register(grp, huma.Operation{
		Method: "GET",
		Path:   "/",
	}, listurls.NewHandler(listurls.NewQueryHandler()).Handle)

	return nil
}
