package server

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"
)

// mockConn implements net.Conn for testing
type mockConn struct {
	readData  []byte
	readPos   int
	writeData bytes.Buffer
	closed    bool
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readPos >= len(m.readData) {
		return 0, io.EOF
	}
	n = copy(b, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeData.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestProtocolDetectConn_HTTP2Passthrough(t *testing.T) {
	// HTTP/2 client preface
	http2Preface := []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")
	additionalData := []byte("additional gRPC data")

	mock := &mockConn{
		readData: append(http2Preface, additionalData...),
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	conn := &protocolDetectConn{Conn: mock, log: log}

	// First read should return the buffered preface
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(buf[:n], http2Preface) {
		t.Errorf("expected HTTP/2 preface, got %q", buf[:n])
	}

	// Subsequent read should return additional data
	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("unexpected error on second read: %v", err)
	}

	if !bytes.Equal(buf[:n], additionalData) {
		t.Errorf("expected additional data %q, got %q", additionalData, buf[:n])
	}

	// Connection should not be closed
	if mock.closed {
		t.Error("connection should not be closed for HTTP/2")
	}
}

func TestProtocolDetectConn_HTTP1Error(t *testing.T) {
	// HTTP/1.1 GET request
	http1Request := []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")

	mock := &mockConn{
		readData: http1Request,
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	conn := &protocolDetectConn{Conn: mock, log: log}

	// Read should return EOF after sending error response
	buf := make([]byte, 100)
	_, err := conn.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}

	// Connection should be closed
	if !mock.closed {
		t.Error("connection should be closed for HTTP/1.1")
	}

	// Should have written an HTTP error response
	response := mock.writeData.String()
	if !strings.Contains(response, "426 Upgrade Required") {
		t.Errorf("expected 426 response, got %q", response)
	}
	if !strings.Contains(response, "gRPC endpoint") {
		t.Errorf("expected friendly message, got %q", response)
	}
}

func TestProtocolDetectConn_POSTRequest(t *testing.T) {
	// HTTP/1.1 POST request (might look like gRPC to naive detection)
	http1Request := []byte("POST /grpc.service/Method HTTP/1.1\r\n")

	mock := &mockConn{
		readData: http1Request,
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	conn := &protocolDetectConn{Conn: mock, log: log}

	buf := make([]byte, 100)
	_, err := conn.Read(buf)
	if err != io.EOF {
		t.Errorf("expected EOF for HTTP/1.1 POST, got %v", err)
	}

	if !mock.closed {
		t.Error("connection should be closed for HTTP/1.1 POST")
	}
}
