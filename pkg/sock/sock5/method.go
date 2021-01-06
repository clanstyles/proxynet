package sock5

type Method byte

const (
	NoAuth   = Method(0x00)
	GSSAPI   = Method(0x01)
	UserPass = Method(0x02)
)
