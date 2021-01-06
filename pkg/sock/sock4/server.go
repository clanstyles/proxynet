package sock4

import (
	"bufio"
	"io"
	"log"
	"net"
	"resnetworking/pkg/proxy"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type Server struct {
	*proxy.Proxy
}

func New(p *proxy.Proxy) *Server {
	return &Server{p}
}

func (srv Server) Handler(src net.Conn, cmd byte) error {
	switch Command(cmd) {
	case Connect:
		if err := srv.HandleConnect(src); err != nil {
			return errors.Wrap(err, "failed to handle connect command")
		}

	default:
		glog.Info("[socks4] command isn't supported.")
	}

	return nil
}

func (srv Server) HandleConnect(src net.Conn) error {
	h, err := ReadHeader(bufio.NewReader(src))
	if err != nil {
		return errors.Wrap(err, "failed to read header")
	}

	log.Printf("header read: %+v", h)

	err = h.ProcessTarget()
	switch {
	case err == ErrNoSuchHost:
		if err := Reply(src, Rejected); err != nil {
			return errors.Wrap(err, "failed to reject no such host")
		}

		return nil

	case err != nil:
		return errors.Wrap(err, "failed to process header target")
	}

	srcClient := src.(*net.TCPConn)
	ok, err := srv.HasAccess(srcClient, h)
	if err != nil {
		return errors.Wrap(err, "failed to check for auth access")
	}

	if !ok {
		if err := Reply(srcClient, Rejected); err != nil {
			return errors.Wrap(err, "failed to reply to rejection")
		}

		log.Printf("[sock proxy] user was denied access to: %s", h.DestinationIP)
		return nil
	}

	// resolve the dst address
	log.Printf("[sock proxy] trying to resolve address ...")

	dstAddr, err := h.Address()
	if err != nil {
		if err := Reply(srcClient, Rejected); err != nil {
			return errors.Wrap(err, "failed to reply to rejection")
		}

		return errors.Wrap(err, "failed to resolve address")
	}

	// Connect to the remote server
	log.Printf("[sock proxy] trying to connect to remote %s...", dstAddr)

	dst, err := net.DialTimeout("tcp", dstAddr, 10*time.Second)
	if err != nil {
		if err := Reply(srcClient, Rejected); err != nil {
			return errors.Wrap(err, "failed to reply to rejection")
		}

		log.Printf("[sock proxy] failed to dial %s: %s", dstAddr, err)
		return errors.Wrap(err, "failed to connect to target")
	}

	dstClient := dst.(*net.TCPConn)

	if err := Reply(srcClient, Granted); err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Transfer the data
	go transfer(&wg, dstClient, srcClient)
	go transfer(&wg, srcClient, dstClient)

	wg.Wait()
	return nil
}

func Reply(src net.Conn, status Status) error {
	var (
		res = [8]byte{
			0x00,
			byte(status),
			0x00,
			0x00,
			0x00,
			0x00,
			0x00,
		}
	)

	if _, err := src.Write(res[:]); err != nil {
		return errors.Wrap(err, "failed to reply")
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
