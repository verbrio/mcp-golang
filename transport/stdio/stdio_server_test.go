package stdio

import (
	"bytes"
	"context"
	"github.com/metoro-io/mcp-golang/transport"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStdioServerTransport(t *testing.T) {
	t.Run("basic message handling", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		tr := NewStdioServerTransportWithIO(in, out)

		var receivedMsg transport.JSONRPCMessage
		var wg sync.WaitGroup
		wg.Add(1)

		tr.SetMessageHandler(func(msg *transport.BaseMessage) {
			receivedMsg = msg
			wg.Done()
		})

		ctx := context.Background()
		err := transport.Start(ctx)
		assert.NoError(t, err)

		// Write a test message to the input buffer
		testMsg := `{"jsonrpc": "2.0", "method": "test", "params": {}, "id": 1}` + "\n"
		_, err = in.Write([]byte(testMsg))
		assert.NoError(t, err)

		// Wait for message processing with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}

		// Verify received message
		req, ok := receivedMsg.(*JSONRPCRequest)
		assert.True(t, ok)
		assert.Equal(t, "test", req.Method)
		assert.Equal(t, mcp.RequestId(1), req.Id)

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("double start error", func(t *testing.T) {
		transport := NewStdioServerTransport()
		ctx := context.Background()
		err := transport.Start(ctx)
		assert.NoError(t, err)

		err = transport.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already started")

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("send message", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		transport := NewStdioServerTransportWithIO(in, out)

		msg := &JSONRPCResponse{
			Jsonrpc: "2.0",
			Result:  Result{AdditionalProperties: map[string]interface{}{"status": "ok"}},
			Id:      1,
		}

		err := transport.Send(msg)
		assert.NoError(t, err)

		// Verify output contains the message and newline
		assert.Contains(t, out.String(), `{"id":1,"jsonrpc":"2.0","result":{"AdditionalProperties":{"status":"ok"}}}`)
		assert.Contains(t, out.String(), "\n")
	})

	t.Run("error handling", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		transport := NewStdioServerTransportWithIO(in, out)

		var receivedErr error
		var wg sync.WaitGroup
		wg.Add(1)

		transport.SetErrorHandler(func(err error) {
			receivedErr = err
			wg.Done()
		})

		ctx := context.Background()
		err := transport.Start(ctx)
		assert.NoError(t, err)

		// Write invalid JSON to trigger error
		_, err = in.Write([]byte(`{"invalid json`))
		assert.NoError(t, err)

		// Write newline to complete the message
		_, err = in.Write([]byte("\n"))
		assert.NoError(t, err)

		// Wait for error handling with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}

		assert.NotNil(t, receivedErr)
		assert.Contains(t, receivedErr.Error(), "unexpected end of JSON input")

		err = transport.Close()
		assert.NoError(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		in := &bytes.Buffer{}
		out := &bytes.Buffer{}
		transport := NewStdioServerTransportWithIO(in, out)

		ctx, cancel := context.WithCancel(context.Background())
		err := transport.Start(ctx)
		assert.NoError(t, err)

		var closed bool
		transport.SetCloseHandler(func() {
			closed = true
		})

		// Cancel context and wait for close
		cancel()
		time.Sleep(100 * time.Millisecond)

		assert.True(t, closed, "transport should be closed after context cancellation")
	})
}
