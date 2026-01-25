package service

import "context"

type SafeBrowsing interface {
	CheckURLv5Proto(ctx context.Context, targetURL string) (ok bool, err error)
}
