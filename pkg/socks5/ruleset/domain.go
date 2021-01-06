package ruleset

import (
	"context"
	"log"
	"resnetworking/pkg/socks5"
	"resnetworking/services/user"
	"resnetworking/services/user_group"
)

type DomainRules struct {
	UserService      user.Service
	UserGroupService user_group.Service
}

func (dr DomainRules) Allow(ctx context.Context, req *socks5.Request) (context.Context, bool) {
	username, ok := req.AuthContext.Payload["Username"]
	if !ok {
		log.Println("[domain rules] user doesn't have a username")
		return ctx, false
	}

	usr, err := dr.UserService.GetByUsername(username)
	switch {
	case err == user.ErrNoUser:
		log.Println("[domain rules] user doesn't exist")
		return ctx, false
	case err != nil:
		log.Printf("[domain rules] failed to get user by username: %s", err)
		return ctx, false
	}

	if usr.UserGroupID == "" {
		log.Printf("[domain rules] no user group found for %s", usr.Username)
		return ctx, true
	}

	grp, err := dr.UserGroupService.GetById(ctx, usr.UserGroupID)
	if err != nil {
		log.Printf("[domain rules] failed to get user group by id: %s", err)
		return ctx, false
	}

	if req.DestAddr.FQDN != "" {
		if grp.HasFQDN(req.DestAddr.FQDN) {
			log.Printf("[domain rules] domain %s isn't allowed for the group %s", req.DestAddr.FQDN, grp.Name)
			return ctx, false
		}
	}

	if grp.HasAddress(req.RemoteAddr.IP.String()) {
		log.Printf("[domain rules] ip %s isn't allowed for the group %s", req.RemoteAddr.IP.String(), grp.Name)
		return ctx, false
	}

	return ctx, true
	// // user, err := dr.UserService.GetByUsername()
	// exists, err := dr.BlacklistService.IPExists(req.DestAddr.IP.String())
	// if err != nil {
	// 	log.Printf("[domain rules] failed to verify if ip exists: %s", err)
	// 	return ctx, false
	// }

	// if exists {
	// 	log.Printf("[proxy] ip should be blocked: %s", req.DestAddr.FQDN)
	// 	return ctx, false
	// }

	// if req.DestAddr.FQDN != "" {
	// 	log.Printf("[proxy] FQDN is: %s", req.DestAddr.FQDN)
	// 	exists, err := dr.BlacklistService.DomainExists(req.DestAddr.FQDN)
	// 	if err != nil {
	// 		log.Printf("[domain rules] failed to verify if domain exists: %s", err)
	// 		return ctx, false
	// 	}

	// 	if exists {
	// 		log.Printf("[proxy] domain should be blocked: %s", req.DestAddr.FQDN)
	// 		return ctx, false
	// 	}
	// }

	// return ctx, true
}
