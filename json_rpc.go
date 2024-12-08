package mcp

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
	Params []byte `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type BaseJSONRPCNotification struct {
	// Jsonrpc corresponds to the JSON schema field "jsonrpc".
	Jsonrpc string `json:"jsonrpc" yaml:"jsonrpc" mapstructure:"jsonrpc"`

	// Method corresponds to the JSON schema field "method".
	Method string `json:"method" yaml:"method" mapstructure:"method"`

	// Params corresponds to the JSON schema field "params".
	// It is stored as a []byte to enable efficient marshaling and unmarshaling into custom types later on in the protocol
	Params []byte `json:"params,omitempty" yaml:"params,omitempty" mapstructure:"params,omitempty"`
}

type BaseMessageType int

const (
	BaseMessageTypeJSONRPCRequestType      BaseMessageType = 1
	BaseMessageTypeJSONRPCNotificationType BaseMessageType = 2
)

type BaseMessage struct {
	Type            BaseMessageType
	RpcMessage      *BaseJSONRPCRequest
	RpcNotification *BaseJSONRPCNotification
}

func NewBaseMessageNotification(notification BaseJSONRPCNotification) *BaseMessage {
	return &BaseMessage{
		Type:            BaseMessageTypeJSONRPCNotificationType,
		RpcNotification: &notification,
	}
}

func NewBaseMessageRequest(request BaseJSONRPCRequest) *BaseMessage {
	return &BaseMessage{
		Type:       BaseMessageTypeJSONRPCRequestType,
		RpcMessage: &request,
	}
}

type BaseCallToolRequestParams struct {
	// Arguments corresponds to the JSON schema field "arguments".
	// It is stored as a []byte to enable efficient marshaling and unmarshaling into custom types later on in the protocol
	Arguments []byte `json:"arguments" yaml:"arguments" mapstructure:"arguments"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}
