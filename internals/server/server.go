package server

import (
	"fmt"
	"net"
	"time"

	http "myserver/internals/http"
	types "myserver/internals/type"
)

type Server struct {
	closed      bool
	listener    net.Listener
	idleTimeout time.Duration
	middlewares *MiddlewareChain
	routes      Routes
}

func NewServer(keepAlive time.Duration) *Server {
	return &Server{
		closed:      false,
		idleTimeout: keepAlive,
		middlewares: NewMiddlewareChain(),
		routes:      make(Routes),
	}
}

func (s *Server) Use(middleware Middleware) {
	s.middlewares.Use(middleware)
}

func handleConnection(conn net.Conn, s *Server) {
	defer conn.Close()

	for {
		if s.idleTimeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(s.idleTimeout))
		}

		req, err := http.ParseRequest(conn)
		response := http.NewResponseWriter(conn, s.idleTimeout)
		if err != nil {
			response.SendBadRequest(err.Error())
			return
		}

		handler, params := s.FindRoute(req.RequestLine.Path, req.RequestLine.Method)

		finalHandler := s.middlewares.Apply(handler)

		req.Params = params
		keepAlive := req.IsKeepAlive()
		response.SetKeppAlive(keepAlive)

		if routeErr := finalHandler(response, req); routeErr != nil {
			switch routeErr.Code {
			case types.NotFound:
				response.SendNotFound(routeErr.Message)
			case types.MethodNotAllowed:
				response.SendBadRequest(routeErr.Message)
			default:
				response.SendInternalServerError(routeErr.Message)
			}
		}
		if !keepAlive {
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
	server := NewServer(10 * time.Second)
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	server.listener = listener
	go server.acceptor()
	return server, nil
}

func (s *Server) Close() error {
	s.closed = true

	if s.listener != nil {
		_ = s.listener.Close()
	}

	return nil
}
