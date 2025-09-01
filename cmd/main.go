package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	http "myserver/internals/http"
	internals "myserver/internals/server"
	types "myserver/internals/type"
)

const port = 8080
const staticDir = "/home/mohe/Documents/github/my-server/static"

// Serve root page
func send(res *http.ResponseWriter, req *http.Request) *types.RouteError {
	res.Status = types.OK
	return res.SendFile(filepath.Join(staticDir, "indx.html"))
}

// Serve static files
func serveStatic(res *http.ResponseWriter, req *http.Request) *types.RouteError {
	path := req.RequestLine.Path
	if path == "/" || path == "." {
		path = "/index.html"
	}

	fullPath := filepath.Join(staticDir, path)
	res.Status = types.OK

	return res.SendFile(fullPath)
}

// Handle login with simple logic
func handleLogin(res *http.ResponseWriter, req *http.Request) *types.RouteError {

	values, err := url.ParseQuery(string(req.Body))
	if err != nil {
		res.Status = types.BadRequest
		return &types.RouteError{Code: types.BadRequest, Message: "Invalid form data"}
	}

	username := values.Get("username")
	password := values.Get("password")

	if username == "admin" && password == "1234" {
		return res.SendJSON(map[string]string{
			"status":  "success",
			"message": "Login successful",
		}, types.OK)
	}

	return res.SendJSON(map[string]string{
		"status":  "error",
		"message": "Invalid credentials",
	}, types.Unauthorized)
}

// Example route with params
func Search(res *http.ResponseWriter, req *http.Request) *types.RouteError {
	firstID, _ := req.Params.Get("firstID")
	secondID, _ := req.Params.Get("secondID")
	fmt.Println("FirstID:", firstID, "SecondID:", secondID)
	return nil
}

func LoggingMiddleware(next internals.Handler) internals.Handler {
	return func(w *http.ResponseWriter, r *http.Request) *types.RouteError {
		log.Printf("[%s] %s %s\n", r.RequestLine.Method, r.RequestLine.Path, r.RequestLine.Version)
		return next(w, r)
	}
}

// Info endpoint
func handleInfo(res *http.ResponseWriter, req *http.Request) *types.RouteError {
	data := map[string]any{
		"status":  "success",
		"message": "Server is running",
		"time":    time.Now().Format(time.RFC3339),
	}
	return res.SendJSON(data, types.OK)
}

// Main function
func main() {
	server, err := internals.ServeHTTP(port)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	log.Println("Server running on:", port)

	server.Use(LoggingMiddleware)

	// Routes with middleware
	server.Handle(types.GET, "/", send)
	server.Handle(types.GET, "/style.css", serveStatic)
	server.Handle(types.GET, "/script.js", serveStatic)
	server.Handle(types.GET, "/search/{firstID}/ds/{secondID}", Search)
	server.Handle(types.GET, "/api/info", handleInfo)
	server.Handle(types.POST, "/login", handleLogin)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
