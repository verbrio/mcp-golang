package main

import "github.com/metoro-io/mcp-golang"

type Content struct {
	Title       string  `json:"title" jsonschema:"description=The title to submit"`
	Description *string `json:"description,omitempty" jsonschema:"description=The description to submit"`
}
type MyFunctionsArguments struct {
	Submitter string  `json:"submitter" jsonschema:"description=The name of the person using the tool"`
	Content   Content `json:"content" jsonschema:"description=The content of the message"`
}

func main() {
	done := make(chan struct{})

	s := mcp.NewServer(mcp.NewStdioServerTransport())
	s.Tool("hello", "Say hello to a person", func(arguments MyFunctionsArguments) (mcp.ToolResponse, error) {
		// ... handle the tool logic
		return mcp.ToolResponse{Content: []mcp.Content{{Type: "text", Text: "Hello, " + arguments.Submitter + "!"}}}, nil
	})

	err := s.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}
