package sock

type version byte

var (
	SOCKS4 = version(0x04)
	SOCKS5 = version(0x05)
)

func (v version) String() string {
	switch v {
	case SOCKS4:
		return "SOCKS4/4a"
	case SOCKS5:
		return "SOCKS5"
	}
	return ""
}
