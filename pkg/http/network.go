package http

import (
	"log"
	"net"
	"net/http"

	"resnetworking/pkg/inet"
	"resnetworking/pkg/proxy"
)

func (p HTTP) NetworkPolicy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
		)

		auth, ok := proxy.FromContext(ctx)
		if !ok {
			log.Println("[http proxy] network policy couldn't find the auth credential")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		host, _, err := inet.SplitHostPort(r.Host)
		if err != nil {
			log.Printf("[http proxy] authorization failed to split host port: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		for _, np := range p.Proxy.NetworkPolicies {
			ok, err := np.HasAccess(ctx, auth, host, net.IP{})
			if err != nil {
				log.Printf("[http proxy] failed to check for network policy access: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if !ok {
				log.Printf("[http proxy] user %s doesn't have access to %s", auth.Username, host)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
