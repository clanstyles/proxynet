package proxy

import (
	"context"
	"errors"
	"net"
)

var (
	ErrAuthenticationFailed = errors.New("failed to authenticate")
)

// Authorization is the token
type Authorization struct {
	Username string
	GroupID  string
}

type AuthenticationRequest struct {
	// Information about who's making the request
	Username    string
	Password    string
	RequestorIP net.IP

	// Information about our destination
	DestinationIP       net.IP
	DestinationHostname string
}

type Authenticator interface {
	Authenticate(context.Context, AuthenticationRequest) (*Authorization, error)
}
