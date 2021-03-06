// Package server provides a generic TCP server implementation.
package server

import (
	"net"
	"sync"
	"time"
	"fmt"

	"github.com/fengyfei/tcp-zero/interfaces"
)

// Server represents a general TCP server.
type Server struct {
	Addr     string
	Protocol interfaces.Protocol

	listener net.Listener
	close    chan struct{}
	once     sync.Once

	hubMutex sync.Mutex
	Hub      interfaces.Hub
}

// NewServer creates a non-TLS TCP server.
func NewServer(addr string, protocol interfaces.Protocol) *Server {
	return &Server{
		Addr:     addr,
		Protocol: protocol,
		close:    make(chan struct{}),
	}
}

// ListenAndServe listens on a address and serves the incomming connections.
func (srv *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	if srv.Hub == nil {
		srv.Hub = newHub()
	}

	return srv.Serve(l)
}

// Serve on the given listener.
func (srv *Server) Serve(l net.Listener) error {
	srv.listener = l
	defer func() {
		srv.listener.Close()
		srv.close <- struct{}{}
	}()

	for {
		var (
			conn net.Conn
			err  error
		)

		select {
		case <-srv.close:
			return nil
		default:
		}

		conn, err = srv.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			return err
		}

		session := newSession(conn)

		go func() {
			srv.Send(*session)
		}()

		if srv.Protocol != nil {
			go func() {
				srv.Put(session)
				srv.Protocol.Handler(session, srv.close)
				srv.Remove(session)
			}()
		}
	}
}

// Close the server immediately.
func (srv *Server) Close() (err error) {
	srv.once.Do(func() {
		close(srv.close)
		srv.Destroy()
	})

	return nil
}

// Put a new connection to hub.
func (srv *Server) Put(session interfaces.Session) error {
	if srv.Hub == nil {
		return nil
	}

	srv.hubMutex.Lock()
	defer srv.hubMutex.Unlock()

	return srv.Hub.Put(session)
}

// Remove a connection from hub, not responsible for closing the connection.
func (srv *Server) Remove(session interfaces.Session) error {
	if srv.Hub == nil {
		return nil
	}

	srv.hubMutex.Lock()
	defer srv.hubMutex.Unlock()

	return srv.Hub.Remove(session)
}

// Destroy a hub.
func (srv *Server) Destroy() error {
	if srv.Hub == nil {
		return nil
	}

	srv.hubMutex.Lock()
	defer srv.hubMutex.Unlock()

	return srv.Hub.Destroy()
}

func (srv *Server) Send(session session) {
	for {
		select {
		case <-srv.close:
			return
		default:
		}

		msg, _ := session.queue.Wait()
		b, _ := msg.Encode()

		_, err := session.conn.Write(b)
		if err != nil {
			fmt.Errorf(err.Error())
		}
	}
}
