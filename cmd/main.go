package main

import (
	"fmt"
	http "myserver/internals/http"
	internals "myserver/internals/server"

	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
)

const port = 8080

func send(res *http.ResponseWriter, req *http.Request) *internals.RouteError {
	res.Status = http.OK
	err := res.SendFile("/home/mohe/Documents/github/my-server/static/indx.html")
	if err != nil {
		return &internals.RouteError{
			Code:    http.InternalServerError,
			Message: "Something went wrong",
		}
	}
	return nil
}

func serveStatic(res *http.ResponseWriter, req *http.Request) *internals.RouteError {

	path := req.RequestLine.Path
	if path == "/" {
		path = "/index.html"
	}

	fullPath := "/home/mohe/Documents/github/my-server/static" + path

	err := res.SendFile(fullPath)
	if err != nil {
		return &internals.RouteError{
			Code:    http.InternalServerError,
			Message: "Failed to serve file",
		}
	}
	return nil
}

func handleLogin(res *http.ResponseWriter, req *http.Request) *internals.RouteError {
	// Parse body (assumes application/x-www-form-urlencoded)
	values, err := url.ParseQuery(string(req.Body))
	if err != nil {
		return &internals.RouteError{
			Code:    http.BadRequest,
			Message: "Invalid form data",
		}
	}

	username := values.Get("username")
	password := values.Get("password")

	// Simple login logic

	if username == "admin" && password == "1234" {
		res.Status = http.OK
		res.Headers.Set("Content-Type", string(http.AppJSON))
		if err := res.SendResponse([]byte(`{"status":"success","message":"Login successful"}`)); err != nil {
			return &internals.RouteError{
				Code:    http.InternalServerError,
				Message: "Failed to send response",
			}
		}
	} else {
		res.Status = http.Unauthorized
		res.Headers.Set("Content-Type", string(http.AppJSON))
		if err := res.SendResponse([]byte(`{"status":"error","message":"Invalid credentials"}`)); err != nil {
			return &internals.RouteError{
				Code:    http.InternalServerError,
				Message: "Failed to send response",
			}
		}
	}

	return nil
}

func Search(res *http.ResponseWriter, req *http.Request) *internals.RouteError {
	d, _ := req.Params.Get("firstID")
	fmt.Println(d)
	return nil
}

func LoggingMiddleware(next internals.Handler) internals.Handler {
	return func(w *http.ResponseWriter, r *http.Request) *internals.RouteError {
		log.Printf("[%s] %s\n", r.RequestLine.Method, r.RequestLine.Path)

		// Call the next handler and return its result
		return next(w, r)
		// return nil  or error
	}
}

func main() {
	server, err := internals.ServeHTTP(port)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	log.Println("Server running on: ", port)

	server.Handle(
		http.GET,
		"/",
		LoggingMiddleware,
		send,
	)
	server.Handle(http.GET, "/style.css", nil, serveStatic)

	server.Handle(http.GET, "/search/{firstID}/ds/{secondID}", nil, Search)
	server.Handle(http.GET, "/script.js", nil, serveStatic)
	server.Handle(http.POST, "/login", nil, handleLogin)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
