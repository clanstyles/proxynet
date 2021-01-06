package proxy

import (
	"context"
	"net"
)

type NetworkPolicy interface {
	HasAccess(ctx context.Context, auth *Authorization, hostname string, address net.IP) (bool, error)
}
