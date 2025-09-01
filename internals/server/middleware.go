package server

// Request → Middleware1 → Middleware2 → Middleware3 → Final Handler → Response

type Middleware func(next Handler) Handler

type MiddlewareChain struct {
	middlewares []Middleware
}

func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}
func (mc *MiddlewareChain) Use(middleware Middleware) {
	mc.middlewares = append(mc.middlewares, middleware)
}

// Apply applies all middlewares to a handler in reverse order
// M1 → M2 → M3 → FinalHandler
func (mc *MiddlewareChain) Apply(finalHandler Handler) Handler {
	handler := finalHandler
	for _, m := range mc.middlewares {
		handler = m(handler)
	}
	return handler
}
