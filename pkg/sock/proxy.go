package sock

import (
	"io"
	"log"
	"net"
	"sync"

	"resnetworking/pkg/proxy"
	"resnetworking/pkg/sock/sock4"
	"resnetworking/pkg/sock/sock5"

	"github.com/pkg/errors"
)

type Server struct {
	*proxy.Proxy

	sock4 *sock4.Server
	sock5 *sock5.Server
}

func New(p *proxy.Proxy) *Server {
	return &Server{
		Proxy: p,

		sock4: sock4.New(p),
		sock5: sock5.New(p),
	}
}

func (p Server) Listen(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			defer conn.Close()

			if err := p.handleConnection(conn); err != nil {
				log.Printf("[socks proxy] connection handling failed: %s", err)
			}

		}(conn)
	}
}

func (p Server) handleConnection(conn net.Conn) error {
	defer conn.Close()

	// Read the first byte of the stream
	var preamble [2]byte
	if _, err := conn.Read(preamble[:]); err != nil {
		log.Printf("[socks proxy] failed to get version byte: %v", err)
		return err
	}

	switch version(preamble[0]) {
	case SOCKS4:
		log.Printf("[socks] handling SOCKS4 proxy")

		if err := p.sock4.Handler(conn, preamble[1]); err != nil {
			return errors.Wrap(err, "failed to handle socks4 connection")
		}
		// if _, err := p.handleSocks4(buff, preamble[1]); err != nil {
		// 	log.Printf("[sock proxy] failed to handle socks4: %s", err)
		// 	return errors.Wrap(err, "failed to process sock4")
		// }

	case SOCKS5:
		log.Printf("[socks] handling SOCKS5 proxy")

		if err := p.sock5.Handler(conn, preamble[1]); err != nil {
			return errors.Wrap(err, "failed to handle socks5 connection")
		}

	default:
		return errors.New("unsupported socks version")
	}

	return nil
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
