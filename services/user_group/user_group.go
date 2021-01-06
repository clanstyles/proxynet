package user_group

type UserGroup struct {
	ID              string
	Name            string
	DomainBlacklist map[string][]string `db:"domain_blacklist"`
	IPBlacklist     []string            `db:"ip_blacklist"`
}

func (ug UserGroup) HasFQDN(fqdn string) bool {
	for domain, _ := range ug.DomainBlacklist {
		if domain == fqdn {
			return true
		}
	}

	return false
}

func (ug UserGroup) HasAddress(addr string) bool {
	for _, ips := range ug.DomainBlacklist {
		for _, ip := range ips {
			if addr == ip {
				return true
			}
		}
	}

	for _, ip := range ug.IPBlacklist {
		if ip == addr {
			return true
		}
	}

	return false
}
