package urls

import (
	"net/http"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/SirNacou/refract/api/internal/config"
	"github.com/SirNacou/refract/api/internal/domain"
	getdashboard "github.com/SirNacou/refract/api/internal/features/urls/get_dashboard"
	listurls "github.com/SirNacou/refract/api/internal/features/urls/list_urls"
	shortenurl "github.com/SirNacou/refract/api/internal/features/urls/shorten_url"
	"github.com/SirNacou/refract/api/internal/infrastructure/persistence"
	"github.com/SirNacou/refract/api/internal/infrastructure/repository"
	"github.com/danielgtaylor/huma/v2"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type Module struct {
	repo   domain.URLRepository
	valkey valkeyaside.CacheAsideClient
	ch     clickhouse.Conn
	cfg    *config.Config
}

func NewModule(db *persistence.DB, valkey valkeyaside.CacheAsideClient, clickhouse clickhouse.Conn, cfg *config.Config) *Module {
	repo := repository.NewPostgresURLRepository(db.Querier)

	return &Module{repo, valkey, clickhouse, cfg}
}

func (m *Module) RegisterRoutes(api huma.API) error {

	grp := huma.NewGroup(api, "/urls")

	huma.Register(grp, huma.Operation{
		OperationID: "list-urls",
		Method:      http.MethodGet,
		Path:        "/",
	}, listurls.NewHandler(listurls.NewQueryHandler(m.repo, m.cfg.DefaultBaseURL)).Handle)

	huma.Register(grp, huma.Operation{
		OperationID: "shorten-url",
		Method:      http.MethodPost,
		Path:        "/",
	}, shortenurl.NewHandler(shortenurl.NewCommandHandler(m.repo, m.valkey, m.cfg.DefaultBaseURL, m.cfg.Valkey.RedirectKey)).Handle)

	huma.Register(grp, huma.Operation{
		OperationID: "dashboard",
		Method:      http.MethodGet,
		Path:        "/dashboard",
	}, getdashboard.NewHandler(getdashboard.NewQueryHandler(m.repo, m.ch)).Handle)

	return nil
}
