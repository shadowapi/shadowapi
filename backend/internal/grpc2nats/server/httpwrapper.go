package server

import (
	"bytes"
	"io"
	"log/slog"
	"net"
)

// httpFallbackListener wraps a net.Listener to detect HTTP/1.1 requests
// and respond with a friendly error message instead of dropping the connection.
// This allows browsers accessing the gRPC endpoint to see a helpful message
// rather than a 502 Bad Gateway error.
type httpFallbackListener struct {
	net.Listener
	log *slog.Logger
}

// wrapListenerWithHTTPFallback wraps the given listener with HTTP/1.1 detection
func wrapListenerWithHTTPFallback(l net.Listener, log *slog.Logger) net.Listener {
	return &httpFallbackListener{
		Listener: l,
		log:      log,
	}
}

// Accept waits for and returns the next connection to the listener.
// Each connection is wrapped with protocol detection.
func (l *httpFallbackListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &protocolDetectConn{Conn: conn, log: l.log}, nil
}

// protocolDetectConn wraps a net.Conn to detect the protocol on first read.
// If HTTP/1.1 is detected, it sends a friendly error response.
// If HTTP/2 (gRPC) is detected, it passes through normally.
type protocolDetectConn struct {
	net.Conn
	log     *slog.Logger
	buf     []byte
	checked bool
}

// http2ClientPreface is the client connection preface for HTTP/2
// Per RFC 7540, it's "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
var http2ClientPreface = []byte("PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")

// Read reads data from the connection, detecting the protocol on first read.
func (c *protocolDetectConn) Read(b []byte) (int, error) {
	if !c.checked {
		c.checked = true

		// Read enough bytes to detect the protocol
		// HTTP/2 preface is 24 bytes: "PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n"
		peek := make([]byte, 24)
		n, err := c.Conn.Read(peek)
		if err != nil {
			return 0, err
		}
		peek = peek[:n]

		// Check if this is HTTP/2 (gRPC uses HTTP/2)
		if bytes.HasPrefix(peek, http2ClientPreface[:min(n, len(http2ClientPreface))]) {
			// HTTP/2 connection - buffer the preface and continue normally
			c.buf = peek
		} else {
			// Not HTTP/2 - likely HTTP/1.1 browser request
			c.log.Debug("received non-gRPC request, sending error response",
				"remote_addr", c.Conn.RemoteAddr().String(),
				"peek", string(peek[:min(n, 20)]),
			)
			c.sendHTTP1Error()
			return 0, io.EOF
		}
	}

	// Return buffered data first
	if len(c.buf) > 0 {
		n := copy(b, c.buf)
		c.buf = c.buf[n:]
		return n, nil
	}

	return c.Conn.Read(b)
}

// sendHTTP1Error sends an HTTP/1.1 error response for non-gRPC clients
func (c *protocolDetectConn) sendHTTP1Error() {
	response := "HTTP/1.1 426 Upgrade Required\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Upgrade: h2c\r\n" +
		"Connection: close\r\n" +
		"Content-Length: 89\r\n" +
		"\r\n" +
		"This is a gRPC endpoint for MeshPump workers.\n" +
		"Please use a gRPC client to connect.\n"
	c.Conn.Write([]byte(response))
	c.Conn.Close()
}
