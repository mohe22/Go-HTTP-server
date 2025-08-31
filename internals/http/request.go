package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

type Method string
type ParseState string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	HEAD    Method = "HEAD"
	OPTIONS Method = "OPTIONS"
	PATCH   Method = "PATCH"
	CONNECT Method = "CONNECT"
	TRACE   Method = "TRACE"
)

const (
	StateRequestLine ParseState = "RequestLine"
	StateHeader      ParseState = "Header"
	StateBody        ParseState = "Body"
	StateDone        ParseState = "Done"
)

type RequestLine struct {
	Method  Method
	Path    string
	Version Version
}
type Request struct {
	RequestLine RequestLine
	Body        []byte
	Headers     Header
	status      ParseState
}

func NewRequestParser() *Request {
	return &Request{
		status:  StateRequestLine,
		Body:    nil,
		Headers: make(Header),
	}
}
func parseMethod(data []byte) (Method, error) {
	switch {
	case bytes.Equal(data, []byte(GET)):
		return GET, nil
	case bytes.Equal(data, []byte(POST)):
		return POST, nil
	case bytes.Equal(data, []byte(DELETE)):
		return DELETE, nil
	case bytes.Equal(data, []byte(PUT)):
		return PUT, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRequestMethod, data)
	}
}

func (req *Request) parseRequestLine(data []byte) (int, error) {
	// method path version \r\n
	idx := bytes.Index(data, []byte(SEPARATOR))
	// we did not get the full line
	// so do not return error
	if idx == -1 {
		return 0, nil
	}
	parts := bytes.Split(data[:idx], []byte(" "))
	if len(parts) != 3 {
		return 0, fmt.Errorf("%w: expected 3 parts, got %d", ErrInvalidRequestLine, len(parts))
	}
	method, err := parseMethod(parts[0])
	if err != nil {
		return 0, err
	}
	version, err := parseVersion(parts[2])
	if err != nil {
		return 0, err
	}
	req.RequestLine.Method = method
	req.RequestLine.Version = version
	req.RequestLine.Path = string(parts[1])
	return idx + len(SEPARATOR), nil
}
func (req *Request) parseRequestHeader(data []byte) (int, error) {
	Idx := bytes.Index(data, []byte(SEPARATOR+SEPARATOR))
	if Idx == -1 {
		return 0, nil
	}
	HeadersBytes := data[:Idx]
	if len(HeadersBytes) == 0 {
		return Idx + len(SEPARATOR+SEPARATOR), nil
	}

	for line := range bytes.SplitSeq(HeadersBytes, []byte(SEPARATOR)) {
		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("%w: invalid header line %q", ErrInvalidHeader, line)
		}
		key := bytes.TrimSpace(parts[0])
		value := bytes.TrimSpace(parts[1])
		if len(key) == 0 || bytes.Contains(key, []byte(" ")) {
			return 0, fmt.Errorf("%w: invalid header name %q", ErrInvalidHeaderName, key)
		}
		req.Headers.Set(string(key), string(value))
	}

	return Idx + len(SEPARATOR+SEPARATOR), nil

}
func (req *Request) parsing(data []byte) (int, error) {
	consumed := 0
	for {
		currentData := data[consumed:]
		switch req.status {
		case StateRequestLine:
			reqLen, err := req.parseRequestLine(currentData)

			if err != nil {
				return consumed, err
			}
			if reqLen == 0 {
				return consumed, nil
			}

			consumed += reqLen
			req.status = StateHeader
		case StateHeader:
			headerLen, err := req.parseRequestHeader(currentData)
			if err != nil {
				return consumed, err
			}
			if headerLen == 0 {
				return consumed, nil
			}
			consumed += headerLen
			req.status = StateBody
		case StateBody:
			length, err := req.Headers.Get("Content-Length")
			if err != nil {
				req.status = StateDone
			}
			n, err := strconv.Atoi(length)
			if err != nil || n == 0 {
				req.status = StateDone
			}
			readLength := min(n-len(req.Body), len(currentData))
			if readLength == 0 {
				// We've read everything available so far, but not the full body
				// Return control so caller can read more data into buffer
				return consumed, nil
			}
			req.Body = append(req.Body, currentData...)
			consumed += readLength
			if len(req.Body) >= n {
				req.status = StateDone
			}
		case StateDone:
			return consumed, nil
		}
	}
}
func ParseRequest(reader io.Reader) (*Request, error) {
	buff := make([]byte, DefaultBufferSize)
	req := NewRequestParser()
	buffLen := 0
	for req.status != StateDone {

		bytesRead, err := reader.Read(buff[buffLen:])
		if err != nil {
			return nil, err
		}
		buffLen += bytesRead

		consumed, err := req.parsing(buff[:buffLen])
		if err != nil {
			// if the error EOF. change the status to Done
			if errors.Is(err, io.EOF) {
				req.status = StateDone
				break
			}
			return nil, err
		}

		// Move leftover data (buff[consumed:buffLen]) to the front (index 0)
		// so we can reuse the buffer for the next read.
		copy(buff, buff[consumed:buffLen])
		buffLen -= consumed
	}

	return req, nil

}
