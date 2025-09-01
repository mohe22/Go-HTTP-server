package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	types "myserver/internals/type"
	url "myserver/internals/utils"
)

type RequestLine struct {
	Method  types.Method
	Path    string
	Version types.Version
}
type Request struct {
	RequestLine RequestLine
	Body        []byte
	Headers     Header
	status      types.ParseState
	Params      url.Params
}

func NewRequestParser() *Request {
	return &Request{
		status:  types.StateRequestLine,
		Body:    nil,
		Headers: make(Header),
	}
}

func (req *Request) IsKeepAlive() bool {
	switch req.RequestLine.Version {
	case types.HTTP1_0:
		conn, _ := req.Headers.Get("Connection")
		return conn != "" && strings.EqualFold(conn, "keep-alive")
	case types.HTTP1_1:
		conn, _ := req.Headers.Get("Connection")
		return conn == "" || !strings.EqualFold(conn, "close")
	case types.HTTP2:
		return true
	default:
		// fallback: close connection
		return false
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
	method, err := types.ParseMethod(parts[0])
	if err != nil {
		return 0, err
	}
	version, err := types.ParseVersion(parts[2])
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
		case types.StateRequestLine:
			reqLen, err := req.parseRequestLine(currentData)

			if err != nil {
				return consumed, err
			}
			if reqLen == 0 {
				return consumed, nil
			}

			consumed += reqLen
			req.status = types.StateHeader
		case types.StateHeader:
			headerLen, err := req.parseRequestHeader(currentData)
			if err != nil {
				return consumed, err
			}
			if headerLen == 0 {
				return consumed, nil
			}
			consumed += headerLen
			req.status = types.StateBody
		case types.StateBody:
			length, err := req.Headers.Get("Content-Length")
			if err != nil {
				req.status = types.StateDone
			}
			n, err := strconv.Atoi(length)
			if err != nil || n == 0 {
				req.status = types.StateDone
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
				req.status = types.StateDone
			}
		case types.StateDone:
			return consumed, nil
		}
	}
}
func ParseRequest(reader io.Reader) (*Request, error) {
	buff := make([]byte, DefaultBufferSize)
	req := NewRequestParser()
	buffLen := 0
	for req.status != types.StateDone {

		bytesRead, err := reader.Read(buff[buffLen:])
		if err != nil {
			return nil, err
		}
		buffLen += bytesRead

		consumed, err := req.parsing(buff[:buffLen])
		if err != nil {
			// if the error EOF. change the status to Done
			if errors.Is(err, io.EOF) {
				req.status = types.StateDone
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
