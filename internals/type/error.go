package types

type RouteError struct {
	Code    StatusCode
	Message string
}

func (r *RouteError) Error() string {
	return r.Message
}
