package tools

type Role string

const RoleAssistant Role = "assistant"
const RoleUser Role = "user"

type ContentAnnotations struct {
	// Describes who the intended customer of this object or data is.
	//
	// It can include multiple entries to indicate ToolResponse useful for multiple
	// audiences (e.g., `["user", "assistant"]`).
	Audience []Role `json:"audience,omitempty" yaml:"audience,omitempty" mapstructure:"audience,omitempty"`

	// Describes how important this data is for operating the server.
	//
	// A value of 1 means "most important," and indicates that the data is
	// effectively required, while 0 means "least important," and indicates that
	// the data is entirely optional.
	Priority *float64 `json:"priority,omitempty" yaml:"priority,omitempty" mapstructure:"priority,omitempty"`
}

// Text provided to or from an LLM.
type TextContent struct {
	// The text ToolResponse of the message.
	Text string `json:"text" yaml:"text" mapstructure:"text"`
}

// An image provided to or from an LLM.
type ImageContent struct {
	// The base64-encoded image data.
	Data string `json:"data" yaml:"data" mapstructure:"data"`

	// The MIME type of the image. Different providers may support different image
	// types.
	MimeType string `json:"mimeType" yaml:"mimeType" mapstructure:"mimeType"`
}

type EmbeddedResourceType string

const (
	EmbeddedResourceTypeBlob EmbeddedResourceType = "blob"
	EmbeddedResourceTypeText EmbeddedResourceType = "text"
)

type BlobResourceContents struct {
	// A base64-encoded string representing the binary data of the item.
	Blob string `json:"blob" yaml:"blob" mapstructure:"blob"`

	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// The URI of this resource.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

type TextResourceContents struct {
	// The MIME type of this resource, if known.
	MimeType *string `json:"mimeType,omitempty" yaml:"mimeType,omitempty" mapstructure:"mimeType,omitempty"`

	// The text of the item. This must only be set if the item can actually be
	// represented as text (not binary data).
	Text string `json:"text" yaml:"text" mapstructure:"text"`

	// The URI of this resource.
	Uri string `json:"uri" yaml:"uri" mapstructure:"uri"`
}

// The contents of a resource, embedded into a prompt or tool call result.
//
// It is up to the client how best to render embedded resources for the benefit
// of the LLM and/or the user.
type EmbeddedResource struct {
	EmbeddedResourceType EmbeddedResourceType
	TextResourceContents *TextResourceContents
	BlobResourceContents *BlobResourceContents
}

type ContentType string

const (
	ContentTypeText             ContentType = "text"
	ContentTypeImage            ContentType = "image"
	ContentTypeEmbeddedResource ContentType = "embedded-resource"
)

// This is a union type of all the different ToolResponse that can be sent back to the client.
// We allow creation through constructors only to make sure that the ToolResponse is valid.
type ToolResponse struct {
	Content []ToolResponseContent
	Error   error
}

func NewToolReponse(content ...ToolResponseContent) *ToolResponse {
	return &ToolResponse{
		Content: content,
	}
}

type ToolResponseContent struct {
	Type             ContentType
	TextContent      *TextContent
	ImageContent     *ImageContent
	EmbeddedResource *EmbeddedResource
	Annotations      *ContentAnnotations
}

// Custom JSON marshaling for ToolResponse.

func (c *ToolResponseContent) WithAnnotations(annotations ContentAnnotations) *ToolResponseContent {
	c.Annotations = &annotations
	return c
}

// NewToolError creates a new ToolResponse that represents an error.
// This is used to create a result that will be returned to the client as an error for a tool call.
func NewToolError(err error) *ToolResponse {
	return &ToolResponse{
		Error: err,
	}
}

// NewToolImageResponseContent creates a new ToolResponse that is an image.
// The given data is base64-encoded
func NewToolImageResponseContent(base64EncodedStringData string, mimeType string) *ToolResponseContent {
	return &ToolResponseContent{
		Type:         ContentTypeImage,
		ImageContent: &ImageContent{Data: base64EncodedStringData, MimeType: mimeType},
	}
}

// NewToolTextResponseContent creates a new ToolResponse that is a simple text string.
// The client will render this as a single string.
func NewToolTextResponseContent(content string) *ToolResponseContent {
	return &ToolResponseContent{
		Type:        ContentTypeText,
		TextContent: &TextContent{Text: content},
	}
}

// NewToolBlobResourceResponseContent creates a new ToolResponse that is a blob of binary data.
// The given data is base64-encoded; the client will decode it.
// The client will render this as a blob; it will not be human-readable.
func NewToolBlobResourceResponseContent(uri string, base64EncodedData string, mimeType string) *ToolResponseContent {
	return &ToolResponseContent{
		Type: ContentTypeEmbeddedResource,
		EmbeddedResource: &EmbeddedResource{
			EmbeddedResourceType: EmbeddedResourceTypeBlob,
			BlobResourceContents: &BlobResourceContents{
				Blob:     base64EncodedData,
				MimeType: &mimeType,
				Uri:      uri,
			}},
	}
}

// NewToolTextResourceResponseContent creates a new ToolResponse that is an embedded resource of type "text".
// The given text is embedded in the response as a TextResourceContents, which
// contains the given MIME type and URI. The text is not base64-encoded.
func NewToolTextResourceResponseContent(uri string, text string, mimeType string) *ToolResponseContent {
	return &ToolResponseContent{
		Type: ContentTypeEmbeddedResource,
		EmbeddedResource: &EmbeddedResource{
			EmbeddedResourceType: EmbeddedResourceTypeText,
			TextResourceContents: &TextResourceContents{
				MimeType: &mimeType, Text: text, Uri: uri,
			}},
	}
}
