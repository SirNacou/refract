package validator

import (
	"context"
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance Validator
	once     sync.Once
)

type Validator interface {
	Struct(s any) error
	StructCtx(ctx context.Context, s any) (err error)
}

func GetValidator() Validator {
	once.Do(func() {
		instance = validator.New(validator.WithRequiredStructEnabled())
	})

	return instance
}
