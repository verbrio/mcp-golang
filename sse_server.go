package mcp

import (
	"fmt"
	"io"
	"net/http"
)

// SSEServerTransport implements a server-side SSE transport
type SSEServerTransport struct {
	transport *SSETransport
}

// NewSSEServerTransport creates a new SSE server transport
func NewSSEServerTransport(endpoint string, w http.ResponseWriter) (*SSEServerTransport, error) {
	transport, err := NewSSETransport(endpoint, w)
	if err != nil {
		return nil, err
	}

	return &SSEServerTransport{
		transport: transport,
	}, nil
}

// Start initializes the SSE connection
func (s *SSEServerTransport) Start() error {
	return s.transport.Start()
}

// HandlePostMessage processes an incoming POST request containing a JSON-RPC message
func (s *SSEServerTransport) HandlePostMessage(r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed: %s", r.Method)
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxMessageSize))
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	defer r.Body.Close()

	return s.transport.HandleMessage(body)
}

// Send sends a message over the SSE connection
func (s *SSEServerTransport) Send(msg JSONRPCMessage) error {
	return s.transport.Send(msg)
}

// Close closes the SSE connection
func (s *SSEServerTransport) Close() error {
	return s.transport.Close()
}

// OnClose sets the callback for when the connection is closed
func (s *SSEServerTransport) OnClose(fn func()) {
	s.transport.OnClose = fn
}

// OnError sets the callback for when an error occurs
func (s *SSEServerTransport) OnError(fn func(error)) {
	s.transport.OnError = fn
}

// OnMessage sets the callback for when a message is received
func (s *SSEServerTransport) OnMessage(fn func(JSONRPCMessage)) {
	s.transport.OnMessage = fn
}

// SessionID returns the unique session identifier for this transport
func (s *SSEServerTransport) SessionID() string {
	return s.transport.SessionID()
}
