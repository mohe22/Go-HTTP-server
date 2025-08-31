package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

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

type ContentType string

const (
	TextPlain ContentType = "text/plain; charset=utf-8"
	TextHTML  ContentType = "text/html; charset=utf-8"
	AppJSON   ContentType = "application/json"
	AppXML    ContentType = "application/xml"
	AppOctet  ContentType = "application/octet-stream"
	ImagePNG  ContentType = "image/png"
	ImageJPEG ContentType = "image/jpeg"
	ImageGIF  ContentType = "image/gif"
)

var statusText = map[StatusCode]string{
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

type ResponseWriter struct {
	write       io.Writer
	idleTimeout time.Duration
	Version     Version
	Status      StatusCode
	Headers     *Header
}

func NewResponseWriter(w io.Writer, idleTimeout time.Duration) *ResponseWriter {
	return &ResponseWriter{
		write:       w,
		idleTimeout: idleTimeout,
		Version:     HTTP1_1,
		Status:      OK,
		Headers:     NewHeader(),
	}
}
func (w *ResponseWriter) WriteStatusLine() error {
	code := w.Status
	text, ok := statusText[code]
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
func detectContentType(data *[]byte) ContentType {
	if len(*data) == 0 {
		return TextPlain
	}

	trimmed := bytes.TrimSpace(*data)

	// JSON detection
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		return AppJSON
	}

	// HTML detection
	if bytes.HasPrefix(trimmed, []byte("<!DOCTYPE html")) ||
		bytes.HasPrefix(trimmed, []byte("<html")) {
		return TextHTML
	}

	// Images detection by magic numbers
	if len(*data) > 4 {
		switch {
		case (*data)[0] == 0x89 && bytes.HasPrefix((*data)[1:], []byte("PNG")):
			return ImagePNG
		case bytes.HasPrefix(*data, []byte{0xFF, 0xD8, 0xFF}):
			return ImageJPEG
		case bytes.HasPrefix(*data, []byte("GIF87a")) || bytes.HasPrefix(*data, []byte("GIF89a")):
			return ImageGIF
		}
	}

	return AppOctet
}
func (w *ResponseWriter) SendResponse(body []byte) error {
	if _, exists := (*w.Headers)["Content-Length"]; !exists {
		w.Headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}
	if _, exists := (*w.Headers)["Date"]; !exists {
		w.Headers.Set("Date", time.Now().UTC().Format(time.RFC1123))
	}
	if _, exists := (*w.Headers)["Connection"]; !exists {
		if w.Version == HTTP1_0 {
			w.Headers.Set("Connection", "close")
		} else {
			w.Headers.Set("Connection", "keep-alive")
			w.Headers.Set("Keep-Alive", fmt.Sprintf("timeout=%d", int(w.idleTimeout.Seconds())))
		}
	}
	if _, exists := (*w.Headers)["Content-Type"]; !exists && len(body) > 0 {
		ct := detectContentType(&body)
		w.Headers.Set("Content-Type", string(ct))
	}
	if err := w.WriteStatusLine(); err != nil {
		return err
	}
	if err := w.WriteHeader(); err != nil {
		return err
	}
	if len(body) > 0 {
		if err := w.WriteBody(body); err != nil {
			return err
		}
	}
	return nil
}
func getContentTypeFromExtension(path string) ContentType {
	switch {
	case strings.HasSuffix(path, ".html"):
		return TextHTML
	case strings.HasSuffix(path, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		return "application/javascript; charset=utf-8"
	case strings.HasSuffix(path, ".json"):
		return AppJSON
	case strings.HasSuffix(path, ".xml"):
		return AppXML
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		return ImageJPEG
	case strings.HasSuffix(path, ".png"):
		return ImagePNG
	case strings.HasSuffix(path, ".gif"):
		return ImageGIF
	default:
		return AppOctet
	}
}

func (w *ResponseWriter) SendFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return w.SendNotFound("File not found")
	}

	// Use extension-based content type
	ct := getContentTypeFromExtension(path)
	w.Headers.Set("Content-Type", string(ct))

	return w.SendResponse(data)
}

// SendBadRequest sends a 400 Bad Request response
func (w *ResponseWriter) SendBadRequest(message string) error {
	w.Status = BadRequest
	body := []byte(message)
	return w.SendResponse(body)
}

// SendNotFound sends a 404 Not Found response
func (w *ResponseWriter) SendNotFound(message string) error {
	w.Status = NotFound
	body := []byte(message)
	return w.SendResponse(body)
}

// SendInternalServerError sends a 500 Internal Server Error response
func (w *ResponseWriter) SendInternalServerError(message string) error {
	w.Status = InternalServerError
	body := []byte(message)
	return w.SendResponse(body)
}
