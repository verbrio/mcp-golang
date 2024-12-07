package mcp

import (
	"bufio"
	"bytes"
	"testing"
)

// TestReadBuffer tests the buffering functionality for JSON-RPC messages.
// The ReadBuffer is crucial for handling streaming input and properly framing messages.
// It verifies:
// 1. Empty buffer handling returns nil message
// 2. Incomplete messages are properly buffered
// 3. Complete messages are correctly extracted
// 4. Multiple message fragments are handled correctly
// 5. Buffer clearing works as expected
// This is a critical test as message framing is fundamental to the protocol.
func TestReadBuffer(t *testing.T) {
	rb := NewReadBuffer()

	// Test empty buffer
	msg, err := rb.ReadMessage()
	if err != nil {
		t.Errorf("ReadMessage failed: %v", err)
	}
	if msg != nil {
		t.Errorf("Expected nil message, got %v", msg)
	}

	// Test incomplete message
	rb.Append([]byte(`{"jsonrpc": "2.0", "method": "test"`))
	msg, err = rb.ReadMessage()
	if err != nil {
		t.Errorf("ReadMessage failed: %v", err)
	}
	if msg != nil {
		t.Errorf("Expected nil message, got %v", msg)
	}

	// Test complete message
	rb.Append([]byte(`, "params": {}}`))
	rb.Append([]byte("\n"))
	msg, err = rb.ReadMessage()
	if err != nil {
		t.Errorf("ReadMessage failed: %v", err)
	}
	if msg == nil {
		t.Error("Expected message, got nil")
	}

	// Test clear
	rb.Clear()
	msg, err = rb.ReadMessage()
	if err != nil {
		t.Errorf("ReadMessage failed: %v", err)
	}
	if msg != nil {
		t.Errorf("Expected nil message, got %v", msg)
	}
}

// TestStdioTransport tests the stdio-based transport implementation.
// The stdio transport is the primary means of communication for the protocol.
// It ensures:
// 1. Messages can be sent correctly with proper framing
// 2. Transport properly handles closing
// 3. Sending after close returns appropriate error
// 4. Output includes required newline termination
// This test is essential as the transport layer must be reliable for the protocol to function.
func TestStdioTransport(t *testing.T) {
	// Create a transport with a buffer instead of actual stdin/stdout
	var input bytes.Buffer
	var output bytes.Buffer
	transport := NewStdioTransport()
	transport.reader = bufio.NewReader(bytes.NewReader(input.Bytes()))
	transport.writer = &output

	// Test sending a message
	message := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "test",
		"params":  map[string]interface{}{},
	}
	if err := transport.Send(message); err != nil {
		t.Errorf("Send failed: %v", err)
	}

	// Verify output
	outputStr := output.String()
	if outputStr == "" {
		t.Error("Expected output, got empty string")
	}
	if outputStr[len(outputStr)-1] != '\n' {
		t.Error("Expected message to end with newline")
	}

	// Test closing
	if err := transport.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Test sending after close
	if err := transport.Send(message); err == nil {
		t.Error("Expected error when sending after close")
	}
}

// TestMessageDeserialization tests the parsing of different JSON-RPC message types.
// Proper message type detection and parsing is critical for protocol operation.
// It tests:
// 1. Request messages (with method and ID)
// 2. Notification messages (with method, no ID)
// 3. Error responses (with error object)
// 4. Success responses (with result)
// Each message type must be correctly identified and parsed to maintain protocol integrity.
func TestMessageDeserialization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType interface{}
	}{
		{
			name:     "request",
			input:    `{"jsonrpc": "2.0", "method": "test", "params": {}, "id": 1}`,
			wantType: &JSONRPCRequest{},
		},
		{
			name:     "notification",
			input:    `{"jsonrpc": "2.0", "method": "test", "params": {}}`,
			wantType: &JSONRPCNotification{},
		},
		{
			name:     "error",
			input:    `{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": 1}`,
			wantType: &JSONRPCError{},
		},
		{
			name:     "response",
			input:    `{"jsonrpc": "2.0", "result": {}, "id": 1}`,
			wantType: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := deserializeMessage(tt.input)
			if err != nil {
				t.Errorf("deserializeMessage failed: %v", err)
			}
			if msg == nil {
				t.Error("Expected message, got nil")
			}
			switch tt.wantType.(type) {
			case *JSONRPCRequest:
				if _, ok := msg.(*JSONRPCRequest); !ok {
					t.Errorf("Expected *JSONRPCRequest, got %T", msg)
				}
			case *JSONRPCNotification:
				if _, ok := msg.(*JSONRPCNotification); !ok {
					t.Errorf("Expected *JSONRPCNotification, got %T", msg)
				}
			case *JSONRPCError:
				if _, ok := msg.(*JSONRPCError); !ok {
					t.Errorf("Expected *JSONRPCError, got %T", msg)
				}
			case map[string]interface{}:
				if _, ok := msg.(map[string]interface{}); !ok {
					t.Errorf("Expected map[string]interface{}, got %T", msg)
				}
			}
		})
	}
}
