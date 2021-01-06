package inet

import (
	"net"
	"strings"

	"github.com/pkg/errors"
)

func SplitHostPort(addr string) (string, string, error) {
	ip, port, err := net.SplitHostPort(addr)
	switch {
	case err != nil && strings.HasSuffix(err.Error(), "missing port in address"):
		ip = addr

	case err != nil:
		return "", "", err
	}

	return ip, port, nil
}

func GetIP(addr string) (net.IP, error) {
	host, _, err := SplitHostPort(addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ip")
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return nil, errors.New("ip is nil")
	}

	return ip, nil
}
