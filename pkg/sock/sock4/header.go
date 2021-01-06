package sock4

import (
	"bufio"
	"encoding/binary"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var ErrNoSuchHost = errors.New("no such host")

type Header struct {
	Destination   string
	DestinationIP net.IP
	Port          int
	Ident         string
}

func (h *Header) ProcessTarget() error {
	if h.Destination != "" {
		log.Printf("[sock proxy] trying to resolve %s", h.Destination)

		ips, err := net.LookupIP(h.Destination)
		switch {
		case err != nil && strings.HasSuffix(err.Error(), "no such host"):
			log.Printf("[sock proxy] no such host: %s", err.Error())
			return ErrNoSuchHost

		case err != nil:
			log.Printf("[sock proxy] failed to resolve %s: %s", h.Destination, err)
			return errors.Wrapf(err, "failed to lookup %s: %s", h.Destination, err)
		}

		log.Printf("[sock proxy] resolved %s to %s", h.Destination, ips[0])
		h.DestinationIP = ips[0]
	}

	return nil
}

func (h Header) Address() (string, error) {
	port := strconv.Itoa(h.Port)
	return net.JoinHostPort(h.DestinationIP.String(), port), nil
}

func ReadHeader(src *bufio.Reader) (*Header, error) {
	var (
		h   Header
		err error
	)

	// Get the destination port
	port := make([]byte, 2)
	if _, err = src.Read(port); err != nil {
		return nil, errors.Wrap(err, "failed to read dst port")
	}
	h.Port = int(binary.BigEndian.Uint16(port[:]))

	// Get the destination ip
	ip := make([]byte, 4)
	if _, err = src.Read(ip); err != nil {
		return nil, errors.Wrap(err, "failed to read dst ip")
	}
	h.DestinationIP = net.IP(ip[:])

	// Read the user's identd
	if h.Ident, err = src.ReadString(0x00); err != nil {
		return nil, errors.Wrap(err, "failed to read ident")
	}

	log.Println("checking for domain")
	// If the IP is 0.0.0.x, we have another payload to process, the hostname
	// This is what we need for Socks4A support
	if h.DestinationIP[0] == 0x00 &&
		h.DestinationIP[1] == 0x00 &&
		h.DestinationIP[2] == 0x00 &&
		h.DestinationIP[3] != 0x00 {

		h.Destination, err = src.ReadString(0x00)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read ident")
		}
		log.Println(h.Destination)
	}

	return &h, nil
}
