# 🛠️ Custom HTTP Server in Go

---

## 🗂 Project Structure

```
my-server/
├─ cmd/
│  └─ main.go                 # Entry point of the server
├─ internals/
│  ├─ http/
│  │  ├─ response.go          # ResponseWriter, headers, SendResponse, SendFile, etc.
│  │  ├─ request.go           # Parsing HTTP requests
│  ├─ server/
│  │  ├─ server.go            # Server struct, connection handling, routing, middlewares
│  │  ├─ routes.go            # Route handling and lookup logic
│  │  └─ middleware.go        # Middleware chain implementation
│  ├─ type/
│  │  └─ types.go             # Status codes, HTTP methods, content types, route errors
│  └─ utils/
│     └─ url.go               # URL and query parameter utilities
├─ static/
│  ├─ index.html
│  ├─ style.css
│  └─ script.js
├─ go.mod
└─ go.sum
```

---

## 🔑 Key Features

- **Custom Server Structure**: Encapsulates server state, routes, and middleware, providing a clean and manageable architecture.
- **Dynamic Routing**: Supports parameterized routes, enabling flexible endpoint definitions like `/user/{id}`.
- **Request Parsing**: Reads and parses incoming HTTP requests into structured objects, including headers, body, and query parameters, making it easy to access client data.
- **Middleware Support**: Allows chaining of middleware functions for tasks such as logging, authentication, and error handling.
- **Keep-Alive Handling**: Manages persistent connections, ensuring efficient resource utilization.
- **Static File Serving**: Serves static assets like HTML, CSS, and JavaScript files, facilitating frontend integration.

---

## 🔍 Key Components

- **Server Struct**: Manages the server's state, including routes, middleware chain, and listener.
- **Handle Method**: Registers route handlers for specific HTTP methods and paths.
- **FindRoute Method**: Matches incoming requests to registered routes, extracting parameters as needed.
- **Middleware Chain**: Allows for the application of multiple middleware functions in a specified order.
- **Connection Handler**: Processes incoming connections, handles requests, applies middleware, and manages the response lifecycle.

---

## 🚀 Next Steps / Improvements

- **Chunked Transfer Encoding**: For streaming large responses efficiently.
- **Binary Data Handling**: Support for serving and processing binary files.
- **Advanced Middleware**: Implement features like rate limiting, CORS handling, and request validation.
- **Graceful Shutdown**: Ensure the server can shut down gracefully, handling ongoing requests appropriately.
- **Testing Suite**: Develop unit and integration tests to ensure reliability and facilitate future development.

---

For a more in-depth exploration and examples, check out my blog:  
[Go HTTP Server Blog](https://portfolio-three-alpha-27.vercel.app/blogs/go-http-server)
