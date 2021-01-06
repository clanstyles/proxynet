package user

type User struct {
	BillingID   string `db:"billing_id"`
	UserGroupID string `db:"user_group_id"`
	Username    string
	Password    string
	Status      Status
	IPs         []string `db:"ips"`
}

type Status int

const (
	Enabled Status = iota
	Disabled
)
