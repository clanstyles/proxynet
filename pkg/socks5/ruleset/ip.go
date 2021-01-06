package ruleset

import (
	"context"
	"log"

	"resnetworking/pkg/socks5"
	"resnetworking/services/user"
)

type IPAuthenticator struct {
	user.Service
}

func (ipa IPAuthenticator) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	_, err := ipa.Service.GetByIP(string(req.RemoteAddr.IP))
	switch {
	case err == user.ErrNoUser:
		return ctx, false
	case err != nil:
		log.Printf("[domain rules] failed to verify if ip exists: %s", err)
		return ctx, false
	}

	return ctx, true
}
