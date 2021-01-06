package user_group

import (
	"context"
	"errors"
)

var (
	ErrNoGroup = errors.New("no user group found.")
)

type Service interface {
	GetById(context.Context, string) (*UserGroup, error)
}
