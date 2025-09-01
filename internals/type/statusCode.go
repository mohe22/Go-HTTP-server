package types

type StatusCode int

const (
	OK                  StatusCode = 200
	Created             StatusCode = 201
	NoContent           StatusCode = 204
	BadRequest          StatusCode = 400
	Unauthorized        StatusCode = 401
	Forbidden           StatusCode = 403
	NotFound            StatusCode = 404
	MethodNotAllowed    StatusCode = 405
	InternalServerError StatusCode = 500
)

var StatusText = map[StatusCode]string{
	OK:                  "OK",
	Created:             "Created",
	NoContent:           "No Content",
	BadRequest:          "Bad Request",
	Unauthorized:        "Unauthorized",
	Forbidden:           "Forbidden",
	NotFound:            "Not Found",
	MethodNotAllowed:    "Method Not Allowed",
	InternalServerError: "Internal Server Error",
}
