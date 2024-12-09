package transport

import "encoding/json"

type JSONRPCMessage interface{}

type RequestId int

type BaseJSONRPCRequest struct {
	// Id corresponds to the JSON schema field "id".
	Id RequestId `json:"id" yaml:"id" mapstructure:"id"`

	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	// It is stored as a []byte to enable efficient marshaling and unmarshaling into custom types later on in the protocol
	Params json.RawMessage `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type BaseJSONRPCNotification struct {
	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	// It is stored as a []byte to enable efficient marshaling and unmarshaling into custom types later on in the protocol
	Params json.RawMessage `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type BaseJSONRPCResponse struct {
}

type BaseMessageType string

const (
	BaseMessageTypeJSONRPCRequestType      BaseMessageType = "request"
	BaseMessageTypeJSONRPCNotificationType BaseMessageType = "notification"
	BaseMessgeTypeJSONRPCResponseType      BaseMessageType = "response"
)

type BaseJsonRpcMessage struct {
	Type                BaseMessageType
	JsonRpcRequest      *BaseJSONRPCRequest
	JsonRpcNotification *BaseJSONRPCNotification
	JsonRpcResponse     *BaseJSONRPCResponse
}

func NewBaseMessageNotification(notification BaseJSONRPCNotification) *BaseJsonRpcMessage {
	return &BaseJsonRpcMessage{
		Type:                BaseMessageTypeJSONRPCNotificationType,
		JsonRpcNotification: &notification,
	}
}

func NewBaseMessageRequest(request BaseJSONRPCRequest) *BaseJsonRpcMessage {
	return &BaseJsonRpcMessage{
		Type:           BaseMessageTypeJSONRPCRequestType,
		JsonRpcRequest: &request,
	}
}

func NewBaseMessageResponse(response BaseJSONRPCRequest) *BaseJsonRpcMessage {
	return &BaseJsonRpcMessage{
		Type:           BaseMessgeTypeJSONRPCResponseType,
		JsonRpcRequest: &response,
	}
}
