package server

import (
	"fmt"
	http "myserver/internals/http"
)

type RouteError struct {
	Code    http.StatusCode
	Message string
}

func (r *RouteError) Error() string {
	return r.Message
}

type Handler func(w *http.ResponseWriter, r *http.Request) *RouteError
type Routes map[http.Method]map[string]Handler

func (s *Server) Handle(method http.Method, path string, handler Handler) {
	if s.routes[method] == nil {
		s.routes[method] = make(map[string]Handler)
	}
	s.routes[method][path] = handler
}

func (s *Server) FindRoute(path string, method http.Method) (*Handler, error) {
	methodRoutes, ok := s.routes[method]
	if !ok {
		return nil, fmt.Errorf("method not allowed: %s", method)
	}

	handler, ok := methodRoutes[path]
	if !ok {
		return nil, fmt.Errorf("path not found: %s (method: %s)", path, method)
	}

	return &handler, nil
}
