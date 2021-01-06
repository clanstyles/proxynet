package http

import (
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"resnetworking/pkg/proxy"
)

type HTTP struct {
	*proxy.Proxy
}

func New(p *proxy.Proxy) *HTTP {
	return &HTTP{
		Proxy: p,
	}
}

func (p HTTP) Listen(l net.Listener) error {
	server := &http.Server{
		Handler: p.Authorize(p.NetworkPolicy(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				p.handleTunneling(w, r)
			} else {
				p.handleHTTP(w, r)
			}
		}))),
		// Disable HTTP/2.
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return server.Serve(l)
}

func (p HTTP) handleTunneling(w http.ResponseWriter, r *http.Request) {
	dst, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("[http proxy] Hijacking is not supported.")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	client, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("[http proxy] failed to hijack: %s", err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	dstTcpClient, _ := dst.(*net.TCPConn)
	srcTcpClient, _ := client.(*net.TCPConn)

	go transfer(dstTcpClient, srcTcpClient)
	go transfer(srcTcpClient, dstTcpClient)
}

func (p HTTP) handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Printf("[http proxy] failed to create round tripper: %s", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("[http proxy] failed to copy data handle HTTP: %s", err)
		return
	}
}

func transfer(dst, src *net.TCPConn) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("[http proxy] failed to copy data transfer: %s", err)
		return
	}

	dst.CloseWrite()
	src.CloseRead()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			log.Printf("%s ] %s", k, v)
			dst.Add(k, v)
		}
	}
}
