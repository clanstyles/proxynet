package user

// import (
// 	"io"
// 	"log"
// 	"net"
// 	"resnetworking/pkg/socks5"
// )

// type ipStore struct {
// 	Service
// }

// func NewIPStore(ds Service) *ipStore {
// 	return &ipStore{ds}
// }

// func (ips ipStore) Authenticate(remoteAddr net.Addr, reader io.Reader, writer io.Writer) (*socks5.AuthContext, error) {
// 	log.Println(remoteAddr.String())
// 	_, err := ips.Service.GetByIP(remoteAddr.String())
// 	switch {
// 	case err == ErrNoUser:
// 		return nil, socks5.UserAuthFailed

// 	case err != nil:
// 		log.Printf("[domain rules] failed to verify if ip exists: %s", err)
// 		return nil, err
// 	}

// 	log.Println("got to get by ip")

// 	return &socks5.AuthContext{socks5.NoAuth, nil}, nil
// }

// func (ips ipStore) GetCode() uint8 {
// 	return socks5.NoAuth
// }
