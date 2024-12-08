package mcp

import (
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// Here we define the actual MCP server that users will create and run
// A server can be passed a number of handlers to handle requests from clients
// Additionally it can be parametrized by a transport. This transport will be used to actually send and receive messages.
// So for example if the stdio transport is used, the server will read from stdin and write to stdout
// If the SSE transport is used, the server will send messages over an SSE connection and receive messages from HTTP POST requests.

// The interface that we're looking to support is something like [gin](https://github.com/gin-gonic/gin)s interface

type Server struct {
	transport          Transport
	Tools              map[string]*ToolType
	serverInstructions *string
	serverName         string
	serverVersion      string
}

type ToolType struct {
	Name            string
	Description     string
	Handler         func(CallToolRequestParamsArguments) (ToolResponse, error)
	ToolInputSchema *jsonschema.Schema
}

type ToolResponse struct {
	Result interface{} `json:"result"`
}

func NewServer(transport Transport) *Server {
	return &Server{
		transport: transport,
		Tools:     make(map[string]*ToolType),
	}
}

func dereferenceReflectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

// Tool registers a new tool with the server
func (s *Server) Tool(name string, description string, handler any) error {
	err := validateHandler(handler)
	if err != nil {
		return err
	}
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	argumentType := handlerType.In(0)

	reflector := jsonschema.Reflector{
		BaseSchemaID:               "",
		Anonymous:                  true,
		AssignAnchor:               false,
		AllowAdditionalProperties:  true,
		RequiredFromJSONSchemaTags: false,
		DoNotReference:             true,
		ExpandedStruct:             true,
		FieldNameTag:               "",
		IgnoredTypes:               nil,
		Lookup:                     nil,
		Mapper:                     nil,
		Namer:                      nil,
		KeyNamer:                   nil,
		AdditionalFields:           nil,
		CommentMap:                 nil,
	}

	inputSchema := reflector.ReflectFromType(argumentType)

	wrappedHandler := func(arguments CallToolRequestParamsArguments) (ToolResponse, error) {
		// We're going to json serialize the arguments and unmarshal them into the correct type
		jsonArgs, err := json.Marshal(arguments)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("failed to marshal arguments: %w", err)
		}

		// Instantiate a struct of the type of the arguments
		unmarshaledArguments := reflect.New(argumentType).Interface()

		// Unmarshal the JSON into the correct type
		err = json.Unmarshal(jsonArgs, &unmarshaledArguments)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}

		// Need to dereference the unmarshaled arguments
		unmarshaledArguments = reflect.ValueOf(unmarshaledArguments).Elem().Interface()

		// Call the handler with the typed arguments
		output := handlerValue.Call([]reflect.Value{reflect.ValueOf(unmarshaledArguments)})

		if len(output) != 2 {
			return ToolResponse{}, fmt.Errorf("tool handler must return exactly two values, got %d", len(output))
		}

		tool := output[0].Interface()
		errorOut := output[1].Interface()
		if errorOut == nil {
			return tool.(ToolResponse), nil
		}
		return tool.(ToolResponse), errorOut.(error)
	}

	s.Tools[name] = &ToolType{
		Name:            name,
		Description:     description,
		Handler:         wrappedHandler,
		ToolInputSchema: inputSchema,
	}

	return nil
}

func (s *Server) Serve() error {
	protocol := NewProtocol(nil)

	protocol.SetRequestHandler("initialize", func(req JSONRPCRequest, _ RequestHandlerExtra) (interface{}, error) {
		return InitializeResult{
			Meta:            nil,
			Capabilities:    s.generateCapabilities(),
			Instructions:    s.serverInstructions,
			ProtocolVersion: "2024-11-05",
			ServerInfo: Implementation{
				Name:    s.serverName,
				Version: s.serverVersion,
			},
		}, nil
	})

	// Definition for a tool the client can call.
	type ToolRetType struct {
		// A human-readable description of the tool.
		Description *string `json:"description,omitempty" yaml:"description,omitempty" mapstructure:"description,omitempty"`

		// A JSON Schema object defining the expected parameters for the tool.
		InputSchema interface{} `json:"inputSchema" yaml:"inputSchema" mapstructure:"inputSchema"`

		// The name of the tool.
		Name string `json:"name" yaml:"name" mapstructure:"name"`
	}

	protocol.SetRequestHandler("tools/list", func(req JSONRPCRequest, _ RequestHandlerExtra) (interface{}, error) {
		return map[string]interface{}{
			"tools": func() []ToolRetType {
				var tools []ToolRetType
				for _, tool := range s.Tools {
					tools = append(tools, ToolRetType{
						Name:        tool.Name,
						Description: &tool.Description,
						InputSchema: tool.ToolInputSchema,
					})
				}
				return tools
			}(),
		}, nil
	})

	protocol.SetRequestHandler("tools/call", func(req JSONRPCRequest, extra RequestHandlerExtra) (interface{}, error) {
		println("here")
		// We need to marshal the arguments into JSON and unmarshal them into the correct type
		jsonArgs, err := json.Marshal(req.Params)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("failed to marshal arguments: %w", err)
		}
		println(string(jsonArgs))

		// Marshal and print extra
		jsonExtra, err := json.Marshal(extra)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("failed to marshal arguments: %w", err)
		}
		println(string(jsonExtra))

		params := CallToolRequestParams{}
		// Instantiate a struct of the type of the arguments
		err = json.Unmarshal(jsonArgs, &params)
		if err != nil {
			return ToolResponse{}, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}

		for name, tool := range s.Tools {
			if name != params.Name {
				continue
			}
			println("Calling tool: " + name)
			return tool.Handler(params.Arguments)
		}
		return ToolResponse{}, fmt.Errorf("unknown tool: %s", req.Method)
	})

	return protocol.Connect(s.transport)
}

func (s *Server) generateCapabilities() ServerCapabilities {
	f := false
	return ServerCapabilities{
		Tools: func() *ServerCapabilitiesTools {
			if s.Tools == nil {
				return nil
			}
			return &ServerCapabilitiesTools{
				ListChanged: &f,
			}
		}(),
	}
}

func validateHandler(handler any) error {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	if handlerType.NumIn() != 1 {
		return fmt.Errorf("handler must take exactly one argument, got %d", handlerType.NumIn())
	}

	if handlerType.NumOut() != 2 {
		return fmt.Errorf("handler must return exactly two values, got %d", handlerType.NumOut())
	}

	// Check that the output type is ToolResponse
	if handlerType.Out(0) != reflect.TypeOf(ToolResponse{}) {
		return fmt.Errorf("handler must return mcp.ToolResponse, got %s", handlerType.Out(0).Name())
	}

	// Check that the output type is error
	if handlerType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("handler must return error, got %s", handlerType.Out(1).Name())
	}

	return nil
}

// validateStruct validates a struct based on its mcp tags
func validateStruct(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("mcp")
		if tag == "" {
			continue
		}

		// Parse the tag
		tagMap := parseTag(tag)

		// Get validation rules
		if validation, ok := tagMap["validation"]; ok {
			if strings.Contains(validation, "maxLength") {
				length := extractMaxLength(validation)
				fieldVal := val.Field(i)
				if fieldVal.Kind() == reflect.String && fieldVal.Len() > length {
					return fmt.Errorf("field %s exceeds maximum length of %d", field.Name, length)
				}
			}
		}

		// If it's a struct, recursively validate
		if field.Type.Kind() == reflect.Struct {
			if err := validateStruct(val.Field(i).Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseTag parses an mcp tag into a map of key-value pairs
func parseTag(tag string) map[string]string {
	result := make(map[string]string)
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.Trim(kv[1], "'")
		}
	}
	return result
}

// extractMaxLength extracts the maximum length from a maxLength validation rule
func extractMaxLength(validation string) int {
	re := regexp.MustCompile(`maxLength\((\d+)\)`)
	matches := re.FindStringSubmatch(validation)
	if len(matches) == 2 {
		length, _ := strconv.Atoi(matches[1])
		return length
	}
	return 0
}
