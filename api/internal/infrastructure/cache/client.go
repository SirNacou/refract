package cache

import (
	"context"
	"fmt"

	"github.com/SirNacou/refract/api/internal/config"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

func NewCache(ctx context.Context, cfg *config.ValkeyConfig) (valkeyaside.CacheAsideClient, error) {
	client, err := valkeyaside.NewClient(
		valkeyaside.ClientOption{
			ClientOption: valkey.ClientOption{
				InitAddress: []string{fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)},
			}},
	)
	if err != nil {
		return nil, err
	}

	err = client.Client().Do(ctx, client.Client().B().Ping().Build()).Error()

	return client, err
}
