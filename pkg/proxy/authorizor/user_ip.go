package authorizor

import (
	"context"
	"log"
	"resnetworking/pkg/proxy"
	"resnetworking/services/user"
)

type UserIP struct {
	user.Service
}

func NewUserIP(us user.Service) *UserIP {
	return &UserIP{
		Service: us,
	}
}

func (c UserIP) Authenticate(ctx context.Context, ar proxy.AuthenticationRequest) (*proxy.Authorization, error) {
	u, err := c.Service.GetByIP(ar.RequestorIP.String())
	switch {
	case err == user.ErrNoIP:
		return nil, proxy.ErrAuthenticationFailed

	case err != nil:
		log.Printf("[credential store] failed to check for get by ip: %s", err)
		return nil, err
	}

	return &proxy.Authorization{
		Username: u.Username,
		GroupID:  u.UserGroupID,
	}, nil
}
