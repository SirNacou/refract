package service

import (
	"github.com/valkey-io/valkey-go/valkeyaside"
)

type Cache interface {
	valkeyaside.CacheAsideClient
}
