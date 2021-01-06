package authenticators

import (
	"log"
	"resnetworking/services/user"
)

type credentialStore struct {
	user.Service
}

func NewCredentialStore(ds user.Service) *credentialStore {
	return &credentialStore{ds}
}

func (cs credentialStore) Valid(username, password string) bool {
	u, err := cs.Service.ValidCredentials(username, password)
	switch {
	case err == user.ErrInvalidCredentials:
		return false
	case err != nil:
		log.Printf("[credential store] failed to check for valid credentails: %s", err)
		return false
	}

	return ok
}
