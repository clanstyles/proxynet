package sock5

type Address byte

const (
	IPv4   = Address(0x01)
	Domain = Address(0x03)
	IPv6   = Address(0x04)
)

func (at Address) String() string {
	switch at {
	case IPv4:
		return "IPv4"
	case Domain:
		return "Domain name"
	case IPv6:
		return "IPv6"
	}

	return ""
}

// func (srv *Server) Handshake(src net.Conn) error {
// 	reader := bufio.NewReader(src)

// 	auth, err := reader.ReadByte()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to read auth method")
// 	}

// }
