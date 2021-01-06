package user

import "errors"

var (
	ErrNoUser             = errors.New("no user exists.")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoIP               = errors.New("IP not found")
)

type Service interface {
	Create(username, password string) (*User, error)
	ValidCredentials(username, password string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByIP(ip string) (*User, error)
	Disable(username string) error
}
