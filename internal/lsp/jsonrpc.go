package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// JSON-RPC 2.0 error codes used by the server (LSP base protocol).
const (
	ParseError       = -32700
	InvalidRequest   = -32600
	MethodNotFound   = -32601
	InvalidParams    = -32602
	InternalError    = -32603
	RequestCancelled = -32800
)

// Message is a JSON-RPC 2.0 message. A request has ID + Method; a notification
// has Method and no ID; a response has ID + (Result | Error).
type Message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`
}

// IsNotification reports whether the message is a notification (no ID).
func (m *Message) IsNotification() bool { return m.ID == nil && m.Method != "" }

// IsRequest reports whether the message is a request (has ID and Method).
func (m *Message) IsRequest() bool { return m.ID != nil && m.Method != "" }

// ResponseError is a JSON-RPC error object.
type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *ResponseError) Error() string { return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message) }

// Conn is a framed JSON-RPC connection over a reader/writer pair (stdin/stdout
// for the LSP). Writes are serialized so concurrent responses and server
// notifications never interleave on the wire.
type Conn struct {
	r   *bufio.Reader
	w   io.Writer
	wmu sync.Mutex
}

// NewConn wraps a reader/writer as a framed JSON-RPC connection.
func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{r: bufio.NewReader(r), w: w}
}

// Read reads one Content-Length-framed message. It returns io.EOF at end of
// input. A malformed header or body yields a non-nil error but never panics.
func (c *Conn) Read() (*Message, error) {
	length, err := readContentLength(c.r)
	if err != nil {
		return nil, err
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, &ResponseError{Code: ParseError, Message: err.Error()}
	}
	return &m, nil
}

// Write frames and writes one message.
func (c *Conn) Write(m *Message) error {
	if m.JSONRPC == "" {
		m.JSONRPC = "2.0"
	}
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	c.wmu.Lock()
	defer c.wmu.Unlock()
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err = c.w.Write(body)
	return err
}

// readContentLength reads the header block and returns the Content-Length. It
// tolerates arbitrary header lines and rejects a missing/invalid length.
func readContentLength(r *bufio.Reader) (int, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" { // blank line terminates the header block
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return 0, &ResponseError{Code: ParseError, Message: "malformed header line: " + line}
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 {
				return 0, &ResponseError{Code: ParseError, Message: "invalid Content-Length: " + value}
			}
			length = n
		}
	}
	if length < 0 {
		return 0, &ResponseError{Code: ParseError, Message: "missing Content-Length header"}
	}
	return length, nil
}

// --- response/notification helpers ---

// NewResponse builds a success response for the given request id and result.
func NewResponse(id *json.RawMessage, result any) (*Message, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	rm := json.RawMessage(raw)
	return &Message{JSONRPC: "2.0", ID: id, Result: rm}, nil
}

// NewErrorResponse builds an error response for the given request id.
func NewErrorResponse(id *json.RawMessage, code int, msg string) *Message {
	return &Message{JSONRPC: "2.0", ID: id, Error: &ResponseError{Code: code, Message: msg}}
}

// NewNotification builds a server-to-client notification.
func NewNotification(method string, params any) (*Message, error) {
	raw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &Message{JSONRPC: "2.0", Method: method, Params: json.RawMessage(raw)}, nil
}
