package webserver

import (
	"github.com/smancke/guble/protocol"
	"net"
	"net/http"
	"strings"
	"time"
)

type WebServer struct {
	server *http.Server
	ln     net.Listener
	mux    *http.ServeMux
	addr   string
}

func New(addr string) *WebServer {
	return &WebServer{
		mux:  http.NewServeMux(),
		addr: addr,
	}
}

func (ws *WebServer) Start() (err error) {
	protocol.Info("webserver: starting up at %v", ws.addr)
	ws.server = &http.Server{Addr: ws.addr, Handler: ws.mux}
	ws.ln, err = net.Listen("tcp", ws.addr)
	if err != nil {
		return
	}

	go func() {
		err = ws.server.Serve(tcpKeepAliveListener{ws.ln.(*net.TCPListener)})

		if err != nil && !strings.HasSuffix(err.Error(), "use of closed network connection") {
			protocol.Err("ListenAndServe %s", err.Error())
		}
		protocol.Info("webserver: http server stopped")
	}()
	return
}

func (ws *WebServer) Stop() error {
	if ws.ln != nil {
		return ws.ln.Close()
	}
	return nil
}

func (ws *WebServer) Check() error {
	return nil
}

func (ws *WebServer) Handle(prefix string, handler http.Handler) {
	ws.mux.Handle(prefix, handler)
}

func (ws *WebServer) GetAddr() string {
	if ws.ln == nil {
		return "::unknown::"
	}
	return ws.ln.Addr().String()
}

// copied from golang: net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(10 * time.Second)
	return tc, nil
}
