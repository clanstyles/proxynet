package network

import (
	"context"
	"net"
	"resnetworking/pkg/proxy"
	"resnetworking/services/user_group"

	"github.com/pkg/errors"
)

type Blacklist struct {
	user_group.Service
}

func NewBlacklist(ugs user_group.Service) *Blacklist {
	return &Blacklist{
		Service: ugs,
	}
}

func (b Blacklist) HasAccess(ctx context.Context, auth *proxy.Authorization, hostname string, address net.IP) (bool, error) {
	ug, err := b.Service.GetById(ctx, auth.GroupID)
	if err != nil {
		return false, errors.Wrap(err, "failed to get user group by id")
	}

	if ug.HasAddress(address.String()) {
		return false, nil
	}

	if ug.HasFQDN(hostname) {
		return false, nil
	}

	return true, nil
}
