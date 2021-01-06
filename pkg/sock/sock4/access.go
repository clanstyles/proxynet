package sock4

import (
	"context"
	"log"
	"net"

	"resnetworking/pkg/inet"
	"resnetworking/pkg/proxy"

	"github.com/pkg/errors"
)

func (srv Server) HasAccess(src *net.TCPConn, h *Header) (bool, error) {
	srcIP, err := inet.GetIP(src.RemoteAddr().String())
	if err != nil {
		return false, errors.Wrap(err, "failed to get ip")
	}

	ar := proxy.AuthenticationRequest{
		Username: "",
		Password: "",

		RequestorIP:         srcIP,
		DestinationIP:       h.DestinationIP,
		DestinationHostname: h.Destination,
	}
	log.Println(srcIP.String())

	var (
		auth *proxy.Authorization
	)
	for _, authorizor := range srv.Proxy.Authenticators {
		auth, err = authorizor.Authenticate(context.Background(), ar)
		log.Println(auth)
		switch {
		case err == proxy.ErrAuthenticationFailed:
			continue

		case err != nil:
			return false, errors.Wrap(err, "authorizor failed")

		default:
			break
		}
	}

	if auth == nil {
		log.Printf("[auth] auth is nil")
		return false, nil
	}

	for _, np := range srv.Proxy.NetworkPolicies {
		log.Println(auth)
		ok, err := np.HasAccess(context.Background(), auth, h.Destination, h.DestinationIP)
		if err != nil {
			return false, errors.Wrap(err, "failed to check for network policy access")
		}

		if !ok {
			return false, nil
		}
	}

	return true, nil
}
