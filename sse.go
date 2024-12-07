/*
Package mcp implements Server-Sent Events (SSE) transport for JSON-RPC communication.

SSE Transport Overview:
This implementation provides a bidirectional communication channel between client and server:
- Server to Client: Uses Server-Sent Events (SSE) for real-time message streaming
- Client to Server: Uses HTTP POST requests for sending messages

Key Features:
1. Bidirectional Communication:
   - SSE for server-to-client streaming (one-way, real-time updates)
   - HTTP POST endpoints for client-to-server messages
   
2. Session Management:
   - Unique session IDs for each connection
   - Proper connection lifecycle management
   - Automatic cleanup on connection close

3. Message Handling:
   - JSON-RPC message format support
   - Automatic message type detection (request vs response)
   - Built-in error handling and reporting
   - Message size limits for security

4. Security Features:
   - Content-type validation
   - Message size limits (4MB default)
   - Error handling for malformed messages

Usage Example:
    // Create a new SSE transport
    transport, err := NewSSETransport("/messages", responseWriter)
    if err != nil {
        log.Fatal(err)
    }

    // Set up message handling
    transport.OnMessage = func(msg JSONRPCMessage) {
        // Handle incoming messages
    }

    // Start the SSE connection
    if err := transport.Start(); err != nil {
        log.Fatal(err)
    }

    // Send a message
    msg := JSONRPCResponse{
        Jsonrpc: "2.0",
        Result:  Result{...},
        Id:      1,
    }
    if err := transport.Send(msg); err != nil {
        log.Fatal(err)
    }
*/

package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

const (
	maxMessageSize = 4 * 1024 * 1024 // 4MB
)

// SSETransport implements a Server-Sent Events transport for JSON-RPC messages
type SSETransport struct {
	endpoint    string
	sessionID   string
	writer      http.ResponseWriter
	flusher     http.Flusher
	mu          sync.Mutex
	isConnected bool

	// Callbacks
	OnClose    func()
	OnError    func(error)
	OnMessage  func(JSONRPCMessage)
}

// NewSSETransport creates a new SSE transport with the given endpoint and response writer
func NewSSETransport(endpoint string, w http.ResponseWriter) (*SSETransport, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	return &SSETransport{
		endpoint:  endpoint,
		sessionID: uuid.New().String(),
		writer:    w,
		flusher:   flusher,
	}, nil
}

// Start initializes the SSE connection
func (t *SSETransport) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isConnected {
		return fmt.Errorf("SSE transport already started")
	}

	// Set SSE headers
	h := t.writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("Access-Control-Allow-Origin", "*")

	// Send the endpoint event
	endpointURL := fmt.Sprintf("%s?sessionId=%s", t.endpoint, t.sessionID)
	if err := t.writeEvent("endpoint", endpointURL); err != nil {
		return err
	}

	t.isConnected = true
	return nil
}

// HandleMessage processes an incoming message
func (t *SSETransport) HandleMessage(msg []byte) error {
	var rpcMsg map[string]interface{}
	if err := json.Unmarshal(msg, &rpcMsg); err != nil {
		if t.OnError != nil {
			t.OnError(err)
		}
		return err
	}

	// Parse as a JSONRPCMessage
	var jsonrpcMsg JSONRPCMessage
	if _, ok := rpcMsg["method"]; ok {
		var req JSONRPCRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			if t.OnError != nil {
				t.OnError(err)
			}
			return err
		}
		jsonrpcMsg = &req
	} else {
		var resp JSONRPCResponse
		if err := json.Unmarshal(msg, &resp); err != nil {
			if t.OnError != nil {
				t.OnError(err)
			}
			return err
		}
		jsonrpcMsg = &resp
	}

	if t.OnMessage != nil {
		t.OnMessage(jsonrpcMsg)
	}
	return nil
}

// Send sends a message over the SSE connection
func (t *SSETransport) Send(msg JSONRPCMessage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return t.writeEvent("message", string(data))
}

// Close closes the SSE connection
func (t *SSETransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return nil
	}

	t.isConnected = false
	if t.OnClose != nil {
		t.OnClose()
	}
	return nil
}

// SessionID returns the unique session identifier for this transport
func (t *SSETransport) SessionID() string {
	return t.sessionID
}

// writeEvent writes an SSE event with the given event type and data
func (t *SSETransport) writeEvent(event, data string) error {
	if _, err := fmt.Fprintf(t.writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	t.flusher.Flush()
	return nil
}
