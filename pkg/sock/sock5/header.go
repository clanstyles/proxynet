package sock5

import (
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrNoSuchHost = errors.New("no such host")
)

type Header struct {
	Command     Command
	AddressType Address
	Address     string
	AddressIP   net.IP
	Port        uint16
}

func (h *Header) ProcessTarget() error {

	// Update the authorization request to check for network policies
	switch h.AddressType {
	case IPv4:
		h.AddressIP = net.ParseIP(h.Address)

	case IPv6:
		h.AddressIP = net.ParseIP(h.Address)

	case Domain:
		if net.ParseIP(h.Address) != nil {
			h.AddressIP = net.ParseIP(h.Address)
			return nil
		}

		ips, err := net.LookupIP(h.Address)
		switch {
		case err != nil && strings.HasSuffix(err.Error(), "no such host"):
			return ErrNoSuchHost
		case err != nil:
			log.Printf("[sock5] failed to resolve %s: %s", h.Address, err)
			return errors.Wrapf(err, "failed to lookup %s: %s", h.Address, err)
		}

		h.AddressIP = ips[0]
	}

	log.Printf("header: %+v", h)
	return nil
}

func (h Header) TargetAddress() string {
	port := strconv.Itoa(int(h.Port))
	return net.JoinHostPort(h.AddressIP.String(), port)
}

// Reply(src net.Conn, reply Replay, addrType Address, bindAddr net.IP, bindPort int) error
func (srv Server) ReadHeader(src net.Conn) (*Header, error) {
	// Grab the socks version
	version := make([]byte, 1)
	if _, err := src.Read(version); err != nil {
		return nil, errors.Wrap(err, "failed to read sock version")
	}

	if version[0] != 0x05 {
		return nil, errors.New("socks version not supported")
		// var (
		// 	res = []byte{
		// 		0x01,
		// 	}
		// )
		// if err := srv.Reply(src, res); err != nil {
		// 	return nil, errors.Wrap(err, "failed to send access granted")
		// }
	}

	cmd := make([]byte, 1)
	if _, err := src.Read(cmd); err != nil {
		return nil, errors.New("failed to read command")
	}

	switch Command(cmd[0]) {
	case Connect:
		// // trash data
		// reserved := make([]byte, 2)
		// if _, err := src.Read([2]byte); err != nil {
		// 	return nil, errors.Wrap(err, "failed to read reserved")
		// }
		reservedLen := 1
		addressType := make([]byte, reservedLen+1)
		if _, err := src.Read(addressType); err != nil {
			return nil, errors.Wrap(err, "failed to read address type")
		}

		log.Println("address type:", addressType[0], addressType[1])
		// Resolve the address field
		var target string
		switch Address(addressType[1]) {
		case IPv4:
			address := make([]byte, 4)
			if _, err := src.Read(address); err != nil {
				return nil, errors.Wrap(err, "failed to read ipv4 address")
			}

			target = net.IP(address).String()

		case IPv6:
			address := make([]byte, 116)
			if _, err := src.Read(address); err != nil {
				return nil, errors.Wrap(err, "failed to read ipv6 address")
			}

			target = net.IP(address).String()

		case Domain:
			domainLen := make([]byte, 1)
			if _, err := src.Read(domainLen); err != nil {
				return nil, errors.Wrap(err, "failed to read domain length")
			}

			targetDst := make([]byte, int(domainLen[0]))
			if _, err := src.Read(targetDst); err != nil {
				return nil, errors.Wrap(err, "failed to read domain")
			}

			target = string(targetDst)

		default:
			if err := srv.Reply(src, AddressTypeNotSupported, Address(0x00), net.IP{}, 0); err != nil {
				return nil, errors.Wrap(err, "failed to reply address type not supported")
			}

			return nil, errors.New("address type not found")
		}

		// Resolve the port
		portDst := make([]byte, 2)
		if _, err := src.Read(portDst); err != nil {
			return nil, errors.Wrap(err, "failed to read port")
		}

		port := binary.BigEndian.Uint16(portDst)
		return &Header{
			Command:     Command(cmd[0]),
			AddressType: Address(addressType[1]),
			Address:     target,
			Port:        port,
		}, nil

	default:
		if err := srv.Reply(src, CommandNotSupported, Address(0x00), net.IP{}, 0); err != nil {
			return nil, errors.Wrap(err, "failed to reply command not supported")
		}

		return nil, errors.New("command not supported")
	}
}
