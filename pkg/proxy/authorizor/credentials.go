package authorizor

import (
	"context"
	"log"
	"resnetworking/pkg/proxy"
	"resnetworking/services/user"
)

type Credentials struct {
	user.Service
}

func NewCredential(us user.Service) *Credentials {
	return &Credentials{
		Service: us,
	}
}

func (c Credentials) Authenticate(ctx context.Context, ar proxy.AuthenticationRequest) (*proxy.Authorization, error) {
	u, err := c.Service.ValidCredentials(ar.Username, ar.Password)
	switch {
	case err == user.ErrInvalidCredentials:
		return nil, proxy.ErrAuthenticationFailed

	case err != nil:
		log.Printf("[credential store] failed to check for valid credentails: %s", err)
		return nil, err
	}

	return &proxy.Authorization{
		Username: u.Username,
		GroupID:  u.UserGroupID,
	}, nil
}
