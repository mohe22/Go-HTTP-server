package server

import (
	"strings"

	http "myserver/internals/http"
	types "myserver/internals/type"
	url "myserver/internals/utils"
)

type Handler func(w *http.ResponseWriter, r *http.Request) *types.RouteError
type Routes map[types.Method]map[string]Handler

func (s *Server) Handle(method types.Method, path string, handler Handler) {

	if s.routes[method] == nil {
		s.routes[method] = make(map[string]Handler)
	}

	s.routes[method][path] = handler
}

// TODO: make sure the logic is working..
func (s *Server) FindRoute(path string, method types.Method) (Handler, url.Params) {

	methodRoutes, ok := s.routes[method]
	if !ok {
		return http.MethodNotAllowedHandler, nil
	}

	url.CleanURL(&path)

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

		if matched && handler != nil {
			return handler, nil
		}
	}

	return http.NotFoundHandler, nil
}
