package sock5

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"resnetworking/pkg/inet"
	"resnetworking/pkg/proxy"

	"github.com/pkg/errors"
)

var (
	ErrAuthFailed = errors.New("authorization failed")
)

type Server struct {
	*proxy.Proxy
}

func New(p *proxy.Proxy) *Server {
	return &Server{
		Proxy: p,
	}
}

func (srv Server) Handler(src net.Conn, nmethods byte) error {
	auth, err := srv.Handshake(src, int(nmethods))
	switch {
	case err == ErrAuthFailed:
		log.Printf("[socks5] authorization failed")
		return nil

	case err != nil:
		return errors.Wrap(err, "handshake failed")
	}

	log.Println(auth)

	header, err := srv.ReadHeader(src)
	if err != nil {
		return errors.Wrap(err, "failed to read header")
	}

	log.Println(header)

	// Convert the domain to an IP if we have to
	if err := header.ProcessTarget(); err != nil {
		return errors.Wrap(err, "failed to process header target")
	}

	for _, np := range srv.Proxy.NetworkPolicies {
		log.Printf("[socks proxy] checking if user has access to %s from %s", header.Address, header.AddressIP)

		ok, err := np.HasAccess(context.Background(), auth, header.Address, header.AddressIP)
		if err != nil {
			return errors.Wrap(err, "failed to check of user has network policy")
		}

		if !ok {
			if err := srv.Reply(src, NotAllowedRuleset, header.AddressType, net.IP{}, 0); err != nil {
				return errors.Wrap(err, "failed to reply with not allowed ruleset")
			}

			return errors.Errorf("user doesn't have access to %s", header.AddressIP)
		}
	}

	dst, err := net.DialTimeout("tcp", header.TargetAddress(), 10*time.Second)
	if err != nil {
		if err := srv.Reply(src, GeneralFailure, header.AddressType, net.IP{}, 0); err != nil {
			return errors.Wrap(err, "failed to reply to generic rejection")
		}

		log.Printf("[sock proxy] failed to dial %s: %s", header.TargetAddress(), err)
		return errors.Wrap(err, "failed to connect to target")
	}

	_, err = src.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01})
	if err != nil {
		return errors.Wrap(err, "failed to write success")
	}
	log.Printf("success")
	// if err := srv.Reply(src, Succeeded, header.AddressType, net.IP{}, 0); err != nil {
	// 	return errors.Wrap(err, "failed to reply to generic rejection")
	// }

	dstClient := dst.(*net.TCPConn)
	srcClient := src.(*net.TCPConn)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Transfer the data
	go transfer(&wg, dstClient, srcClient)
	go transfer(&wg, srcClient, dstClient)

	wg.Wait()
	return nil
}

func (srv Server) Reply(src net.Conn, reply Replay, addrType Address, bindAddr net.IP, bindPort int) error {
	log.Printf("[socks5] replying to server")

	var buff bytes.Buffer
	writer := bufio.NewWriter(&buff)

	if _, err := writer.Write([]byte{0x05}); err != nil {
		return errors.Wrap(err, "failed to write socket version")
	}

	if _, err := writer.Write([]byte{byte(reply)}); err != nil {
		return errors.Wrap(err, "failed to write reply type")
	}

	if _, err := writer.Write([]byte{0x00}); err != nil {
		return errors.Wrap(err, "failed to write the reserve bit")
	}

	if _, err := writer.Write([]byte{byte(addrType)}); err != nil {
		return errors.Wrap(err, "failed to write address type")
	}

	if _, err := writer.Write(bindAddr); err != nil {
		return errors.Wrap(err, "failed to write bind address")
	}

	dstBindPort := make([]byte, 2)
	binary.BigEndian.PutUint16(dstBindPort, uint16(bindPort))

	if _, err := writer.Write(dstBindPort); err != nil {
		return errors.Wrap(err, "failed to write bind port")
	}

	if _, err := src.Write(buff.Bytes()); err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}

// func (srv Server) HandleRequest(ar *proxy.AuthenticationRequest, src net.Conn) error {

// }

func (srv Server) Handshake(src net.Conn, nmethods int) (*proxy.Authorization, error) {
	// Get the auth methods from the headers
	methods := make([]byte, nmethods)

	if _, err := src.Read(methods); err != nil {
		return nil, errors.Wrap(err, "failed to read nmethods")
	}

	srcIP, err := inet.GetIP(src.RemoteAddr().String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ip")
	}

	// Build our auth request
	ar := proxy.AuthenticationRequest{
		RequestorIP: srcIP,
	}

	// Do we have UserAuth type?
	for _, m := range methods {
		switch Method(m) {
		case UserPass:
			// Respond with a User/Pass auth request
			var (
				res = []byte{
					0x05,
					byte(UserPass),
				}
			)

			if _, err := src.Write(res); err != nil {
				return nil, errors.Wrap(err, "failed to reply")
			}

			authVersion := make([]byte, 1)
			if _, err := src.Read(authVersion); err != nil {
				return nil, errors.Wrap(err, "failed to read auth version")
			}

			if authVersion[0] != 0x01 {
				return nil, errors.New("unsupported auth version")
			}

			usernameLen := make([]byte, 1)
			if _, err := src.Read(usernameLen); err != nil {
				return nil, errors.Wrap(err, "failed to read auth version")
			}

			username := make([]byte, int(usernameLen[0]))
			if _, err := src.Read(username); err != nil {
				return nil, errors.Wrap(err, "failed to read username")
			}

			passwordLen := make([]byte, 1)
			if _, err := src.Read(passwordLen); err != nil {
				return nil, errors.Wrap(err, "failed to read auth version")
			}

			password := make([]byte, int(passwordLen[0]))
			if _, err := src.Read(password); err != nil {
				return nil, errors.Wrap(err, "failed to read password")
			}

			ar.Username = string(username)
			ar.Password = string(password)

			// case NoAuth:
			// 	break

			// default:
			// 	var (
			// 		res = []byte{
			// 			0x05,
			// 			0xff,
			// 		}
			// 	)
			// 	if _, err := src.Write(res); err != nil {
			// 		return nil, errors.Wrap(err, "failed to reply")
			// 	}

			// 	// if err := srv.Reply(src, res); err != nil {
			// 	// 	return nil, errors.Wrap(err, "failed to reply")
			// 	// }

			// 	return nil, errors.New("auth method not supported")
		}
	}

	var auth *proxy.Authorization
	for _, authorizor := range srv.Proxy.Authenticators {
		auth, err = authorizor.Authenticate(context.Background(), ar)

		if err == proxy.ErrAuthenticationFailed {
			continue
		}

		if err != nil {
			return nil, errors.Wrap(err, "authorizor failed")
		}

		break
	}

	if auth == nil {
		var (
			res = []byte{
				0x01,
				0xFF,
			}
		)

		log.Println("wrote failure")
		if _, err := src.Write(res); err != nil {
			return nil, errors.Wrap(err, "failed to send access denied")
		}

		return nil, ErrAuthFailed
	}

	// A successful auth happened
	var (
		res = []byte{0x01, 0x00}
	)

	log.Println("wrote success")
	if _, err := src.Write(res); err != nil {
		return nil, errors.Wrap(err, "failed to send access granted")
	}

	return auth, nil
}

func transfer(wg *sync.WaitGroup, dst, src *net.TCPConn) {
	defer func() {
		dst.CloseWrite()
		src.CloseRead()

		wg.Done()
	}()

	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("[http proxy] failed to copy data transfer: %s", err)
		return
	}
}
