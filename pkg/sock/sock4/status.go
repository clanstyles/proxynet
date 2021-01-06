package sock4

type Status byte

const (
	Granted     = Status(0x5a)
	Rejected    = Status(0x5b)
	NoIdentd    = Status(0x5c)
	IdentFailed = Status(0x5d)
)
