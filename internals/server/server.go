package server

import (
	"fmt"
	http "myserver/internals/http"
	"net"
	"strings"
	"time"
)

type Server struct {
	closed      bool
	listener    net.Listener
	idleTimeout time.Duration
	routes      Routes
}

func NewServer() *Server {
	return &Server{
		closed:      false,
		idleTimeout: 10 * time.Second,
		routes:      make(Routes),
	}
}

func handleConnection(conn net.Conn, s *Server) {
	defer func() {
		_ = conn.Close()
	}()

	for {
		if s.idleTimeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(s.idleTimeout))
		}

		req, err := http.ParseRequest(conn)
		response := http.NewResponseWriter(conn, s.idleTimeout)
		if err != nil {
			response.SendBadRequest("Method Not Allowed")
			return
		}

		handler, err := s.FindRoute(req.RequestLine.Path, req.RequestLine.Method)
		if err != nil {
			fmt.Println(err)
			if strings.Contains(err.Error(), "method not allowed") {
				response.SendBadRequest("Method Not Allowed")
			} else if strings.Contains(err.Error(), "path not found") {
				response.SendNotFound("Path Not Found")
			} else {
				response.SendInternalServerError("Internal Server Error")
			}
			return
		}
		if err := (*handler)(response, req); err != nil {
			response.SendInternalServerError(err.Error())
			return
		}
	}
}

func (s *Server) acceptor() {
	for {
		if s.closed {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed {
				return
			}
			continue
		}

		go handleConnection(conn, s)
	}
}

func ServeHTTP(port uint16) (*Server, error) {
	server := NewServer()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server.listener = listener
	go server.acceptor()
	return server, nil
}

func (s *Server) Close() error {
	// set closed flag first
	s.closed = true

	// close listener to unblock Accept()
	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}
