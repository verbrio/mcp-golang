package tools

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/sjson"
)

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

// Custom JSON marshaling for EmbeddedResource
func (c EmbeddedResource) MarshalJSON() ([]byte, error) {
	switch c.EmbeddedResourceType {
	case EmbeddedResourceTypeBlob:
		return json.Marshal(c.BlobResourceContents)
	case EmbeddedResourceTypeText:
		return json.Marshal(c.TextResourceContents)
	default:
		return nil, fmt.Errorf("unknown embedded resource type: %s", c.EmbeddedResourceType)
	}
}

type ContentType string

const (
	// The value is the value of the "type" field in the ToolResponseContent so do not change
	ContentTypeText             ContentType = "text"
	ContentTypeImage            ContentType = "image"
	ContentTypeEmbeddedResource ContentType = "resource"
)

// This is a union type of all the different ToolResponse that can be sent back to the client.
// We allow creation through constructors only to make sure that the ToolResponse is valid.
type ToolResponse struct {
	Content []*ToolResponseContent
}

type ToolResponseSent struct {
	*ToolResponse
	Error error
}

// Custom JSON marshaling for ToolResponse
func (c ToolResponseSent) MarshalJSON() ([]byte, error) {
	if c.Error != nil {
		errorText := c.Error.Error()
		c.Content = []*ToolResponseContent{NewToolTextResponseContent(errorText)}
	}
	return json.Marshal(struct {
		Content []*ToolResponseContent `json:"content" yaml:"content" mapstructure:"content"`
		IsError bool                   `json:"isError" yaml:"isError" mapstructure:"isError"`
	}{
		Content: c.Content,
		IsError: c.Error != nil,
	})
}

func NewToolReponse(content ...*ToolResponseContent) *ToolResponse {
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

// Custom JSON marshaling for ToolResponse Content
func (c ToolResponseContent) MarshalJSON() ([]byte, error) {
	rawJson := []byte{}

	switch c.Type {
	case ContentTypeText:
		j, err := json.Marshal(c.TextContent)
		if err != nil {
			return nil, err
		}
		rawJson = j
	case ContentTypeImage:
		j, err := json.Marshal(c.ImageContent)
		if err != nil {
			return nil, err
		}
		rawJson = j
	case ContentTypeEmbeddedResource:
		j, err := json.Marshal(c.EmbeddedResource)
		if err != nil {
			return nil, err
		}
		rawJson = j
	default:
		return nil, fmt.Errorf("unknown content type: %s", c.Type)
	}

	// Add the type
	rawJson, err := sjson.SetBytes(rawJson, "type", string(c.Type))
	if err != nil {
		return nil, err
	}
	// Add the annotations
	if c.Annotations != nil {
		marshal, err := json.Marshal(c.Annotations)
		if err != nil {
			return nil, err
		}
		rawJson, err = sjson.SetBytes(rawJson, "annotations", marshal)
		if err != nil {
			return nil, err
		}
	}
	return rawJson, nil
}

func (c *ToolResponseContent) WithAnnotations(annotations ContentAnnotations) *ToolResponseContent {
	c.Annotations = &annotations
	return c
}

// NewToolResponseSentError creates a new ToolResponse that represents an error.
// This is used to create a result that will be returned to the client as an error for a tool call.
func NewToolResponseSentError(err error) *ToolResponseSent {
	return &ToolResponseSent{
		Error: err,
	}
}

// NewToolResponseSent creates a new ToolResponseSent
func NewToolResponseSent(response *ToolResponse) *ToolResponseSent {
	return &ToolResponseSent{
		ToolResponse: response,
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
