package http

import (
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"strings"

	"resnetworking/pkg/inet"
	"resnetworking/pkg/proxy"
)

func (p HTTP) Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := inet.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("[http proxy] authorization failed to split host port: %s", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		username, password, _ := parseBasicAuth(r.Header.Get("Proxy-Authorization"))

		ar := proxy.AuthenticationRequest{
			RequestorIP: net.ParseIP(ip),
			Username:    username,
			Password:    password,
		}

		for _, authorizor := range p.Proxy.Authenticators {
			auth, err := authorizor.Authenticate(r.Context(), ar)
			switch {
			case err == proxy.ErrAuthenticationFailed:
				continue

			case err != nil:
				log.Printf("[http proxy] authorizor failed: %s", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			r = r.WithContext(proxy.NewContext(r.Context(), auth))
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Proxy-Authenticate", "Basic realm=\"\"")
		http.Error(w, http.StatusText(http.StatusProxyAuthRequired), http.StatusProxyAuthRequired)
		return
	})
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
