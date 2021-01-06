package proxy

import "net"

type Proxy struct {
	Authenticators  []Authenticator
	NetworkPolicies []NetworkPolicy
}

func New(auths []Authenticator, nps []NetworkPolicy) *Proxy {
	return &Proxy{
		Authenticators:  auths,
		NetworkPolicies: nps,
	}
}

type Server interface {
	Listen(l net.Listener) error
	// ConnectionHandler(conn net.Conn) error
}
