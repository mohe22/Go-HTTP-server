package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	types "myserver/internals/type"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ResponseWriter struct {
	write       io.Writer
	idleTimeout time.Duration
	isKeepAlive bool
	Version     types.Version
	Status      types.StatusCode
	Headers     *Header
}

func NewResponseWriter(w io.Writer, idleTimeout time.Duration) *ResponseWriter {
	return &ResponseWriter{
		write:       w,
		idleTimeout: idleTimeout,
		isKeepAlive: false,
		Version:     types.HTTP1_1,
		Status:      types.OK,
		Headers:     NewHeader(),
	}
}

func (w *ResponseWriter) SetKeppAlive(isAlive bool) {
	w.isKeepAlive = isAlive
}

func (w *ResponseWriter) WriteStatusLine() error {
	code := w.Status
	text, ok := types.StatusText[code]
	if !ok {
		return ErrUnknownStatusCode
	}
	_, err := fmt.Fprintf(w.write, "%s %d %s\r\n", w.Version.String(), code, text)
	return err
}

// WriteHeader writes headers to the response
func (w *ResponseWriter) WriteHeader() error {
	for key, value := range *w.Headers {
		if _, err := fmt.Fprintf(w.write, "%s: %s\r\n", key, value); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w.write, "\r\n")
	return err
}

func (w *ResponseWriter) WriteBody(data []byte) error {
	_, err := w.write.Write(data)
	return err
}
func detectContentType(data *[]byte) types.ContentType {
	if len(*data) == 0 {
		return types.TextPlain
	}

	trimmed := bytes.TrimSpace(*data)

	// JSON detection
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return types.AppJSON
	}

	// HTML detection
	if bytes.HasPrefix(trimmed, []byte("<!DOCTYPE html")) ||
		bytes.HasPrefix(trimmed, []byte("<html")) {
		return types.TextHTML
	}

	// Images detection by magic numbers
	if len(*data) > 4 {
		switch {
		case (*data)[0] == 0x89 && bytes.HasPrefix((*data)[1:], []byte("PNG")):
			return types.ImagePNG
		case bytes.HasPrefix(*data, []byte{0xFF, 0xD8, 0xFF}):
			return types.ImageJPEG
		case bytes.HasPrefix(*data, []byte("GIF87a")) || bytes.HasPrefix(*data, []byte("GIF89a")):
			return types.ImageGIF
		}
	}

	return types.AppOctet
}
func getContentTypeFromExtension(path string) types.ContentType {
	switch {
	case strings.HasSuffix(path, ".html"):
		return types.TextHTML
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return types.AppJSON
	case strings.HasSuffix(path, ".xml"):
		return types.AppXML
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return types.ImageJPEG
	case strings.HasSuffix(path, ".png"):
		return types.ImagePNG
	case strings.HasSuffix(path, ".gif"):
		return types.ImageGIF
	default:
		return types.AppOctet
	}
}

func (w *ResponseWriter) SetDefaultHeaders(body *[]byte) {
	// Content-Length
	if _, exists := (*w.Headers)["Content-Length"]; !exists {
		w.Headers.Set("Content-Length", strconv.Itoa(len(*body)))
	}

	// Date
	if _, exists := (*w.Headers)["Date"]; !exists {
		w.Headers.Set("Date", time.Now().UTC().Format(time.RFC1123))
	}

	// Connection / Keep-Alive
	if _, exists := (*w.Headers)["Connection"]; !exists {
		if w.isKeepAlive {
			w.Headers.Set("Connection", "keep-alive")
			w.Headers.Set("Keep-Alive", fmt.Sprintf("timeout=%d", int(w.idleTimeout.Seconds())))
		} else {
			w.Headers.Set("Connection", "close")
		}
	}

	// Content-Type
	if len(*body) > 0 {
		if _, exists := (*w.Headers)["Content-Type"]; !exists {
			ct := detectContentType(body)
			w.Headers.Set("Content-Type", string(ct))
		}
	}
}

func (w *ResponseWriter) SendResponse(body []byte) *types.RouteError {
	w.SetDefaultHeaders(&body)

	if err := w.WriteStatusLine(); err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
	}
	if err := w.WriteHeader(); err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
	}
	if len(body) > 0 {
		if err := w.WriteBody(body); err != nil {
			return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
		}
	}
	return nil
}
func (w *ResponseWriter) SendFile(path string) *types.RouteError {
	file, err := os.Open(path)

	if err != nil {
		return &types.RouteError{Code: types.NotFound, Message: "File not found"}
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: "Failed to read file info"}
	}

	w.Headers.Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	if _, exists := (*w.Headers)["Content-Type"]; !exists {
		ct := getContentTypeFromExtension(filepath.Ext(path))
		w.Headers.Set("Content-Type", string(ct))
	}

	if err := w.WriteStatusLine(); err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
	}
	if err := w.WriteHeader(); err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
	}

	if _, err := io.Copy(w.write, file); err != nil {
		return &types.RouteError{Code: types.InternalServerError, Message: err.Error()}
	}

	return nil
}

func (w *ResponseWriter) SendJSON(data any, status types.StatusCode) *types.RouteError {
	body, err := json.Marshal(data)
	if err != nil {
		w.Status = types.InternalServerError
		return &types.RouteError{Code: types.InternalServerError, Message: "Failed to encode JSON"}
	}

	w.Status = status

	return w.SendResponse(body)
}

// SendBadRequest sends a 400 Bad Request response
func (w *ResponseWriter) SendBadRequest(message string) error {
	w.Status = types.BadRequest
	body := []byte(message)
	return w.SendResponse(body)
}

// SendNotFound sends a 404 Not Found response
func (w *ResponseWriter) SendNotFound(message string) error {
	w.Status = types.NotFound
	body := []byte(message)
	return w.SendResponse(body)
}

// SendInternalServerError sends a 500 Internal Server Error response
func (w *ResponseWriter) SendInternalServerError(message string) error {
	w.Status = types.InternalServerError
	body := []byte(message)
	return w.SendResponse(body)
}
func NotFoundHandler(w *ResponseWriter, r *Request) *types.RouteError {
	return &types.RouteError{
		Code:    types.NotFound,
		Message: "Path Not Found",
	}
}

func MethodNotAllowedHandler(w *ResponseWriter, r *Request) *types.RouteError {
	return &types.RouteError{
		Code:    types.MethodNotAllowed,
		Message: "Method Not Allowed",
	}
}
