package mcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSEServerTransport(t *testing.T) {
	t.Run("basic message handling", func(t *testing.T) {
		w := httptest.NewRecorder()
		transport, err := NewSSEServerTransport("/messages", w)
		assert.NoError(t, err)

		var receivedMsg JSONRPCMessage
		transport.OnMessage(func(msg JSONRPCMessage) {
			receivedMsg = msg
		})

		err = transport.Start()
		assert.NoError(t, err)

		// Verify SSE headers
		headers := w.Header()
		assert.Equal(t, "text/event-stream", headers.Get("Content-Type"))
		assert.Equal(t, "no-cache", headers.Get("Cache-Control"))
		assert.Equal(t, "keep-alive", headers.Get("Connection"))

		// Verify endpoint event was sent
		body := w.Body.String()
		assert.Contains(t, body, "event: endpoint")
		assert.Contains(t, body, "/messages?sessionId=")

		// Test message handling
		msg := JSONRPCRequest{
			Jsonrpc: "2.0",
			Method:  "test",
			Id:      1,
		}
		msgBytes, err := json.Marshal(msg)
		assert.NoError(t, err)

		httpReq := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewReader(msgBytes))
		httpReq.Header.Set("Content-Type", "application/json")
		err = transport.HandlePostMessage(httpReq)
		assert.NoError(t, err)

		// Verify received message
		rpcReq, ok := receivedMsg.(*JSONRPCRequest)
		assert.True(t, ok)
		assert.Equal(t, "2.0", rpcReq.Jsonrpc)
		assert.Equal(t, RequestId(1), rpcReq.Id)

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("send message", func(t *testing.T) {
		w := httptest.NewRecorder()
		transport, err := NewSSEServerTransport("/messages", w)
		assert.NoError(t, err)

		err = transport.Start()
		assert.NoError(t, err)

		msg := JSONRPCResponse{
			Jsonrpc: "2.0",
			Result:  Result{AdditionalProperties: map[string]interface{}{"status": "ok"}},
			Id:      1,
		}

		err = transport.Send(msg)
		assert.NoError(t, err)

		// Verify output contains the message
		body := w.Body.String()
		assert.Contains(t, body, `event: message`)
		assert.Contains(t, body, `"result":{"AdditionalProperties":{"status":"ok"}}`)
	})

	t.Run("error handling", func(t *testing.T) {
		w := httptest.NewRecorder()
		transport, err := NewSSEServerTransport("/messages", w)
		assert.NoError(t, err)

		var receivedErr error
		transport.OnError(func(err error) {
			receivedErr = err
		})

		err = transport.Start()
		assert.NoError(t, err)

		// Test invalid JSON
		req := httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		err = transport.HandlePostMessage(req)
		assert.Error(t, err)
		assert.NotNil(t, receivedErr)
		assert.Contains(t, receivedErr.Error(), "invalid")

		// Test invalid content type
		req = httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "text/plain")
		err = transport.HandlePostMessage(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported content type")

		// Test invalid method
		req = httptest.NewRequest(http.MethodGet, "/messages", nil)
		err = transport.HandlePostMessage(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method not allowed")
	})
}
