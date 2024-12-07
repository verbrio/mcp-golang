// Package mcp provides a Go implementation of the MCP (Metoro Communication Protocol).
//
// This file implements the stdio transport layer for JSON-RPC communication.
// It provides functionality to read and write JSON-RPC messages over standard input/output
// streams, similar to the TypeScript implementation in @typescript-sdk/src/shared/stdio.ts.
//
// Key Components:
//
// 1. ReadBuffer:
//    - Buffers continuous stdio stream into discrete JSON-RPC messages
//    - Thread-safe with mutex protection
//    - Handles message framing using newline delimiters
//    - Methods: Append (add data), ReadMessage (read complete message), Clear (reset buffer)
//
// 2. StdioTransport:
//    - Implements the Transport interface using stdio
//    - Uses bufio.Reader for efficient buffered reading
//    - Thread-safe with mutex protection
//    - Supports:
//      * Asynchronous message reading
//      * Message sending with newline framing
//      * Proper cleanup on close
//      * Event handlers for close, error, and message events
//
// 3. Message Handling:
//    - Deserializes JSON-RPC messages into appropriate types:
//      * JSONRPCRequest: Messages with ID and method
//      * JSONRPCNotification: Messages with method but no ID
//      * JSONRPCError: Error responses with ID
//      * Generic responses: Success responses with ID
//    - Serializes messages to JSON with newline termination
//
// Thread Safety:
//    - All public methods are thread-safe
//    - Uses sync.Mutex for state protection
//    - Safe for concurrent reading and writing
//
// Usage:
//    transport := NewStdioTransport()
//    transport.SetMessageHandler(func(msg interface{}) {
//        // Handle incoming message
//    })
//    transport.Start()
//    defer transport.Close()
//
//    // Send a message
//    transport.Send(map[string]interface{}{
//        "jsonrpc": "2.0",
//        "method": "test",
//        "params": map[string]interface{}{},
//    })
//
// Error Handling:
//    - All methods return meaningful errors
//    - Transport supports error handler for async errors
//    - Proper cleanup on error conditions
//
// For more details, see the test file stdio_test.go.
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// ReadBuffer buffers a continuous stdio stream into discrete JSON-RPC messages.
type ReadBuffer struct {
	mu     sync.Mutex
	buffer []byte
}

// NewReadBuffer creates a new ReadBuffer.
func NewReadBuffer() *ReadBuffer {
	return &ReadBuffer{}
}

// Append adds a chunk of data to the buffer.
func (rb *ReadBuffer) Append(chunk []byte) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.buffer == nil {
		rb.buffer = chunk
	} else {
		rb.buffer = append(rb.buffer, chunk...)
	}
}

// ReadMessage reads a complete JSON-RPC message from the buffer.
// Returns nil if no complete message is available.
func (rb *ReadBuffer) ReadMessage() (interface{}, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.buffer == nil {
		return nil, nil
	}

	// Find newline
	for i := 0; i < len(rb.buffer); i++ {
		if rb.buffer[i] == '\n' {
			// Extract line
			line := string(rb.buffer[:i])
			rb.buffer = rb.buffer[i+1:]
			return deserializeMessage(line)
		}
	}

	return nil, nil
}

// Clear clears the buffer.
func (rb *ReadBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buffer = nil
}

// deserializeMessage deserializes a JSON-RPC message from a string.
func deserializeMessage(line string) (interface{}, error) {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC message: %w", err)
	}

	// Check if it's a request, response, or notification
	if _, hasID := msg["id"]; hasID {
		if _, hasMethod := msg["method"]; hasMethod {
			var req JSONRPCRequest
			if err := json.Unmarshal([]byte(line), &req); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON-RPC request: %w", err)
			}
			return &req, nil
		}
		if _, hasError := msg["error"]; hasError {
			var err JSONRPCError
			if err := json.Unmarshal([]byte(line), &err); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON-RPC error: %w", err)
			}
			return &err, nil
		}
		// Must be a response
		return msg, nil
	}
	// Must be a notification
	var notif JSONRPCNotification
	if err := json.Unmarshal([]byte(line), &notif); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON-RPC notification: %w", err)
	}
	return &notif, nil
}

// serializeMessage serializes a JSON-RPC message to a string.
func serializeMessage(message interface{}) (string, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON-RPC message: %w", err)
	}
	return string(bytes) + "\n", nil
}

// StdioTransport implements Transport interface using stdio.
type StdioTransport struct {
	reader     *bufio.Reader
	writer     io.Writer
	readBuffer *ReadBuffer

	closeHandler    func()
	errorHandler   func(error)
	messageHandler func(interface{})

	closed bool
	mu     sync.RWMutex
	wg     sync.WaitGroup
}

// NewStdioTransport creates a new StdioTransport.
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader:     bufio.NewReader(os.Stdin),
		writer:     os.Stdout,
		readBuffer: NewReadBuffer(),
	}
}

// Start starts reading from stdin.
func (t *StdioTransport) Start() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.Unlock()

	t.wg.Add(1)
	go t.readLoop()
	return nil
}

// Send sends a message to stdout.
func (t *StdioTransport) Send(message interface{}) error {
	t.mu.RLock()
	if t.closed {
		t.mu.RUnlock()
		return fmt.Errorf("transport is closed")
	}
	t.mu.RUnlock()

	serialized, err := serializeMessage(message)
	if err != nil {
		return err
	}

	_, err = t.writer.Write([]byte(serialized))
	return err
}

// Close closes the transport.
func (t *StdioTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()

	if t.closeHandler != nil {
		t.closeHandler()
	}

	t.wg.Wait()
	return nil
}

// SetCloseHandler sets the close handler.
func (t *StdioTransport) SetCloseHandler(handler func()) {
	t.mu.Lock()
	t.closeHandler = handler
	t.mu.Unlock()
}

// SetErrorHandler sets the error handler.
func (t *StdioTransport) SetErrorHandler(handler func(error)) {
	t.mu.Lock()
	t.errorHandler = handler
	t.mu.Unlock()
}

// SetMessageHandler sets the message handler.
func (t *StdioTransport) SetMessageHandler(handler func(interface{})) {
	t.mu.Lock()
	t.messageHandler = handler
	t.mu.Unlock()
}

// readLoop reads messages from stdin continuously.
func (t *StdioTransport) readLoop() {
	defer t.wg.Done()

	buf := make([]byte, 4096)
	for {
		t.mu.RLock()
		if t.closed {
			t.mu.RUnlock()
			return
		}
		t.mu.RUnlock()

		n, err := t.reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				t.mu.RLock()
				if t.errorHandler != nil {
					t.errorHandler(fmt.Errorf("read error: %w", err))
				}
				t.mu.RUnlock()
			}
			return
		}

		t.readBuffer.Append(buf[:n])

		for {
			msg, err := t.readBuffer.ReadMessage()
			if err != nil {
				t.mu.RLock()
				if t.errorHandler != nil {
					t.errorHandler(fmt.Errorf("failed to read message: %w", err))
				}
				t.mu.RUnlock()
				break
			}
			if msg == nil {
				break
			}

			t.mu.RLock()
			if t.messageHandler != nil {
				t.messageHandler(msg)
			}
			t.mu.RUnlock()
		}
	}
}
