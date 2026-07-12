package lsp

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"
)

func frame(body string) string {
	return "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body
}

func TestReadWriteRoundTrip(t *testing.T) {
	in := frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"x":1}}`)
	conn := NewConn(strings.NewReader(in), &bytes.Buffer{})
	msg, err := conn.Read()
	if err != nil {
		t.Fatal(err)
	}
	if !msg.IsRequest() || msg.Method != "initialize" {
		t.Errorf("parsed = %+v, want initialize request", msg)
	}

	var out bytes.Buffer
	wconn := NewConn(strings.NewReader(""), &out)
	resp, err := NewResponse(msg.ID, map[string]string{"ok": "yes"})
	if err != nil {
		t.Fatal(err)
	}
	if err := wconn.Write(resp); err != nil {
		t.Fatal(err)
	}
	// The written bytes must be a valid frame that reads back.
	back, err := NewConn(bytes.NewReader(out.Bytes()), &bytes.Buffer{}).Read()
	if err != nil {
		t.Fatalf("re-read written frame: %v", err)
	}
	if back.Result == nil {
		t.Errorf("response missing result: %+v", back)
	}
	if string(*back.ID) != "1" {
		t.Errorf("response id = %s, want 1", string(*back.ID))
	}
}

func TestNotificationHasNoID(t *testing.T) {
	in := frame(`{"jsonrpc":"2.0","method":"initialized","params":{}}`)
	msg, err := NewConn(strings.NewReader(in), &bytes.Buffer{}).Read()
	if err != nil {
		t.Fatal(err)
	}
	if !msg.IsNotification() || msg.IsRequest() {
		t.Errorf("expected a notification, got %+v", msg)
	}
}

func TestMalformedHeadersError(t *testing.T) {
	cases := []string{
		"no headers at all\r\n\r\n{}",   // missing Content-Length
		"Content-Length: abc\r\n\r\n{}", // non-numeric length
		"Content-Length: 5\r\n\r\n{}",   // body shorter than declared
	}
	for _, in := range cases {
		if _, err := NewConn(strings.NewReader(in), &bytes.Buffer{}).Read(); err == nil {
			t.Errorf("expected error for input %q", in)
		}
	}
}

// FuzzJSONRPC asserts the framing reader never panics on arbitrary bytes
// (malformed protocol input must not crash the server — spec safety/robustness).
func FuzzJSONRPC(f *testing.F) {
	f.Add([]byte("Content-Length: 2\r\n\r\n{}"))
	f.Add([]byte("garbage"))
	f.Add([]byte("Content-Length: -1\r\n\r\n"))
	f.Add([]byte("Content-Length: 999999\r\n\r\n{}"))
	f.Fuzz(func(t *testing.T, data []byte) {
		conn := NewConn(bytes.NewReader(data), &bytes.Buffer{})
		// Read until error/EOF; must terminate without panicking.
		for i := 0; i < 100; i++ {
			m, err := conn.Read()
			if err != nil {
				return
			}
			_ = m
		}
	})
}

// TestJSONValues keeps encoding/json referenced and documents the id shape.
func TestJSONValues(t *testing.T) {
	raw := json.RawMessage(`7`)
	m := &Message{JSONRPC: "2.0", ID: &raw, Method: "x"}
	if !m.IsRequest() {
		t.Error("message with id+method should be a request")
	}
}
