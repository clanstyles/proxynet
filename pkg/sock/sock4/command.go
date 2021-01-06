package sock4

type Command byte

const (
	Connect = Command(0x01)
	Bind    = Command(0x20)
	UDP     = Command(0x03)
)

func (c Command) String() string {
	switch c {
	case Connect:
		return "Connect"
	case Bind:
		return "Bind"
	case UDP:
		return "UDP Associate"
	}

	return ""
}
