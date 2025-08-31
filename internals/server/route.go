package server

import (
	"fmt"
	"strings"

	http "myserver/internals/http"
	url "myserver/internals/utils"
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

// TODO: make sure the logic is working..
func (s *Server) FindRoute(path string, method http.Method) (*Handler, url.Params, error) {
	methodRoutes, ok := s.routes[method]
	if !ok {
		return nil, nil, fmt.Errorf("method not allowed: %s", method)
	}

	reqSegments := strings.Split(strings.Trim(path, "/"), "/")

	for routePath, handler := range methodRoutes {
		routeSegments := strings.Split(strings.Trim(routePath, "/"), "/")
		if len(routeSegments) != len(reqSegments) {
			continue
		}

		params := make(url.Params)
		matched := true

		for i, seg := range reqSegments {
			rSeg := routeSegments[i]
			if url.IsParam(rSeg) {
				paramName := rSeg[1 : len(rSeg)-1]
				params[paramName] = seg
			} else if rSeg != seg {
				matched = false
				break
			}
		}

		if matched {
			return &handler, params, nil
		}
	}

	return nil, nil, fmt.Errorf("path not found: %s (method: %s)", path, method)
}
