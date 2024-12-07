package mcp

// Role represents the sender or recipient of messages and data in a conversation
type Role string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

// LoggingLevel represents the severity of a log message
type LoggingLevel string

const (
	LogLevelEmergency LoggingLevel = "emergency"
	LogLevelAlert     LoggingLevel = "alert"
	LogLevelCritical  LoggingLevel = "critical"
	LogLevelError     LoggingLevel = "error"
	LogLevelWarning   LoggingLevel = "warning"
	LogLevelNotice    LoggingLevel = "notice"
	LogLevelInfo      LoggingLevel = "info"
	LogLevelDebug     LoggingLevel = "debug"
)

// RequestID represents a uniquely identifying ID for a request in JSON-RPC
type RequestID interface{} // can be string or integer

// ProgressToken represents a token used to associate progress notifications with the original request
type ProgressToken interface{} // can be string or integer

// Cursor represents an opaque token used for pagination
type Cursor string

// Implementation describes the name and version of an MCP implementation
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Annotations represents optional annotations for objects
type Annotations struct {
	Audience []Role  `json:"audience,omitempty"`
	Priority float64 `json:"priority,omitempty"`
}

// ModelHint provides hints for model selection
type ModelHint struct {
	Name string `json:"name,omitempty"`
}

// ModelPreferences represents the server's preferences for model selection
type ModelPreferences struct {
	CostPriority         float64     `json:"costPriority,omitempty"`
	IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
	SpeedPriority        float64     `json:"speedPriority,omitempty"`
	Hints                []ModelHint `json:"hints,omitempty"`
}

// ClientCapabilities represents capabilities a client may support
type ClientCapabilities struct {
	Experimental map[string]map[string]interface{} `json:"experimental,omitempty"`
	Roots        *struct {
		ListChanged bool `json:"listChanged"`
	} `json:"roots,omitempty"`
	Sampling map[string]interface{} `json:"sampling,omitempty"`
}

// ServerCapabilities represents capabilities that a server may support
type ServerCapabilities struct {
	Experimental map[string]map[string]interface{} `json:"experimental,omitempty"`
	Logging      map[string]interface{}            `json:"logging,omitempty"`
	Prompts      *struct {
		ListChanged bool `json:"listChanged"`
	} `json:"prompts,omitempty"`
	Resources *struct {
		ListChanged bool `json:"listChanged"`
		Subscribe   bool `json:"subscribe"`
	} `json:"resources,omitempty"`
	Tools *struct {
		ListChanged bool `json:"listChanged"`
	} `json:"tools,omitempty"`
}

// Content interfaces and implementations
type Content interface {
	GetType() string
	GetAnnotations() *Annotations
}

// TextContent represents text provided to or from an LLM
type TextContent struct {
	Type        string       `json:"type"`
	Text        string       `json:"text"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

func (t TextContent) GetType() string              { return t.Type }
func (t TextContent) GetAnnotations() *Annotations { return t.Annotations }

// ImageContent represents an image provided to or from an LLM
type ImageContent struct {
	Type        string       `json:"type"`
	Data        string       `json:"data"`
	MimeType    string       `json:"mimeType"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

func (i ImageContent) GetType() string              { return i.Type }
func (i ImageContent) GetAnnotations() *Annotations { return i.Annotations }

// ResourceContents represents the contents of a specific resource
type ResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
}

// TextResourceContents represents text-based resource contents
type TextResourceContents struct {
	ResourceContents
	Text string `json:"text"`
}

// BlobResourceContents represents binary resource contents
type BlobResourceContents struct {
	ResourceContents
	Blob string `json:"blob"`
}

// EmbeddedResource represents resource contents embedded in a prompt or tool call
type EmbeddedResource struct {
	Type        string       `json:"type"`
	Resource    interface{}  `json:"resource"` // can be TextResourceContents or BlobResourceContents
	Annotations *Annotations `json:"annotations,omitempty"`
}

// Message types
type SamplingMessage struct {
	Content Content `json:"content"`
	Role    Role    `json:"role"`
}

type PromptMessage struct {
	Content interface{} `json:"content"` // can be TextContent, ImageContent, or EmbeddedResource
	Role    Role        `json:"role"`
}

// Resource represents a known resource that the server can read
type Resource struct {
	Name        string       `json:"name"`
	URI         string       `json:"uri"`
	MimeType    string       `json:"mimeType,omitempty"`
	Description string       `json:"description,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

// ResourceTemplate represents a template for resources available on the server
type ResourceTemplate struct {
	Name        string       `json:"name"`
	URITemplate string       `json:"uriTemplate"`
	MimeType    string       `json:"mimeType,omitempty"`
	Description string       `json:"description,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

// Root represents a root directory or file that the server can operate on
type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

// Tool represents a tool definition that the client can call
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// PromptArgument describes an argument that a prompt can accept
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Prompt represents a prompt or prompt template that the server offers
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// Reference types
type PromptReference struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ResourceReference struct {
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// Request/Response structs
type MetaParams struct {
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}

type BaseRequest struct {
	Method string     `json:"method"`
	Params MetaParams `json:"params,omitempty"`
}

type BaseResult struct {
	Meta map[string]interface{} `json:"_meta,omitempty"`
}

// Initialize specific types
type InitializeRequest struct {
	Method string `json:"method"`
	Params struct {
		Capabilities    ClientCapabilities `json:"capabilities"`
		ClientInfo      Implementation     `json:"clientInfo"`
		ProtocolVersion string             `json:"protocolVersion"`
	} `json:"params"`
}

type InitializeResult struct {
	BaseResult
	Capabilities    ServerCapabilities `json:"capabilities"`
	Instructions    string             `json:"instructions,omitempty"`
	ProtocolVersion string             `json:"protocolVersion"`
	ServerInfo      Implementation     `json:"serverInfo"`
}

// CreateMessage specific types
type CreateMessageRequest struct {
	Method string `json:"method"`
	Params struct {
		Messages         []SamplingMessage      `json:"messages"`
		MaxTokens        int                    `json:"maxTokens"`
		Temperature      float64                `json:"temperature,omitempty"`
		StopSequences    []string               `json:"stopSequences,omitempty"`
		SystemPrompt     string                 `json:"systemPrompt,omitempty"`
		ModelPreferences *ModelPreferences      `json:"modelPreferences,omitempty"`
		IncludeContext   string                 `json:"includeContext,omitempty"`
		Metadata         map[string]interface{} `json:"metadata,omitempty"`
	} `json:"params"`
}

type CreateMessageResult struct {
	BaseResult
	Content    Content `json:"content"`
	Model      string  `json:"model"`
	Role       Role    `json:"role"`
	StopReason string  `json:"stopReason,omitempty"`
}

// Notification types
type CancelledNotification struct {
	Method string `json:"method"`
	Params struct {
		RequestID RequestID `json:"requestId"`
		Reason    string    `json:"reason,omitempty"`
	} `json:"params"`
}

type ProgressNotification struct {
	Method string `json:"method"`
	Params struct {
		ProgressToken ProgressToken `json:"progressToken"`
		Progress      float64       `json:"progress"`
		Total         float64       `json:"total,omitempty"`
	} `json:"params"`
}

type ResourceUpdatedNotification struct {
	Method string `json:"method"`
	Params struct {
		URI string `json:"uri"`
	} `json:"params"`
}

type LoggingMessageNotification struct {
	Method string `json:"method"`
	Params struct {
		Level  LoggingLevel `json:"level"`
		Data   interface{}  `json:"data"`
		Logger string       `json:"logger,omitempty"`
	} `json:"params"`
}
