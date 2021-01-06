package sock5

type Replay byte

const (
	Succeeded               = Replay(0x00)
	GeneralFailure          = Replay(0x01)
	NotAllowedRuleset       = Replay(0x02)
	NetworkUnreachable      = Replay(0x03)
	HostUnreachable         = Replay(0x04)
	ConnectionRefused       = Replay(0x05)
	TTLExpired              = Replay(0x06)
	CommandNotSupported     = Replay(0x07)
	AddressTypeNotSupported = Replay(0x08)
)
