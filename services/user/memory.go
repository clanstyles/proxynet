package user

type MemDatastore struct {
	users     []*User
	usersByIP map[string]*User
}

func NewMemDatastore() *MemDatastore {
	return &MemDatastore{}
}

func (ds MemDatastore) Create(username, password string) (*User, error) {
	user := User{
		Password: password,
		Username: username,
	}

	ds.users = append(ds.users, &user)
	return &user, nil
}

func (ds MemDatastore) GetByUsername(username string) (*User, error) {
	for _, user := range ds.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, ErrNoUser
}

func (ds MemDatastore) ValidCredentials(username, password string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := ds.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	if user.Password == password {
		return user, nil
	}

	return nil, ErrInvalidCredentials
}

func (ds MemDatastore) GetByIP(ip string) (*User, error) {
	user, ok := ds.usersByIP[ip]
	if !ok {
		return nil, ErrNoIP
	}

	return user, nil
}

func (ds MemDatastore) Disable(username string) error {
	return nil
}
