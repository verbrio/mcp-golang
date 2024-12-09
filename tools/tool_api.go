package tools

type Role string

const RoleAssistant Role = "assistant"
const RoleUser Role = "user"

type ContentAnnotations struct {
	// Describes who the intended customer of this object or data is.
	//
	// It can include multiple entries to indicate Content useful for multiple
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
	// The text Content of the message.
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
	ContentTypeError            ContentType = "error"
)

// This is a union type of all the different Content that can be sent back to the client.
// We allow creation through constructors only to make sure that the Content is valid.
type Content struct {
	Type             ContentType
	TextContent      *TextContent
	ImageContent     *ImageContent
	EmbeddedResource *EmbeddedResource
	Annotations      *ContentAnnotations
	Error            error
}

func (c *Content) WithAnnotations(annotations ContentAnnotations) *Content {
	c.Annotations = &annotations
	return c
}

// NewToolError creates a new Content that represents an error.
// This is used to create a result that will be returned to the client as an error for a tool call.
func NewToolError(err error) *Content {
	return &Content{
		Type:  ContentTypeError,
		Error: err,
	}
}

// NewImageContent creates a new Content that is an image.
// The given data is base64-encoded
func NewImageContent(base64EncodedStringData string, mimeType string) *Content {
	return &Content{
		Type:         ContentTypeImage,
		ImageContent: &ImageContent{Data: base64EncodedStringData, MimeType: mimeType},
	}
}

// NewTextContent creates a new Content that is a simple text string.
// The client will render this as a single string.
func NewTextContent(content string) *Content {
	return &Content{
		Type:        ContentTypeText,
		TextContent: &TextContent{Text: content},
	}
}

// NewBlobResource creates a new Content that is a blob of binary data.
// The given data is base64-encoded; the client will decode it.
// The client will render this as a blob; it will not be human-readable.
func NewBlobResource(uri string, base64EncodedData string, mimeType string) *Content {
	return &Content{
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

// NewTextResource creates a new Content that is an embedded resource of type "text".
// The given text is embedded in the response as a TextResourceContents, which
// contains the given MIME type and URI. The text is not base64-encoded.
func NewTextResource(uri string, text string, mimeType string) *Content {
	return &Content{
		Type: ContentTypeEmbeddedResource,
		EmbeddedResource: &EmbeddedResource{
			EmbeddedResourceType: EmbeddedResourceTypeText,
			TextResourceContents: &TextResourceContents{
				MimeType: &mimeType, Text: text, Uri: uri,
			}},
	}
}
