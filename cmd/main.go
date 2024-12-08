package main

import "github.com/metoro-io/mcp-golang"

type Content struct {
	Title       string  `json:"hello" jsonschema:"description=The title to submit"`
	Description *string `json:"world,omitempty" jsonschema:"description=The description to submit"`
}
type MyFunctionsArguments struct {
	Submitter string  `json:"foo" jsonschema:"description=The name of the person using the tool"`
	Content   Content `json:"bar" jsonschema:"description=The content of the message"`
}

func main() {
	done := make(chan struct{})

	s := mcp.NewServer(mcp.NewStdioServerTransport())
	s.Tool("hello", "Say hello to a person", func(arguments MyFunctionsArguments) (mcp.ToolResponse, error) {
		// ... handle the tool logic
		return mcp.ToolResponse{Result: "Submitted " + arguments.Content.Title}, nil
	})

	//(*s.Tools["test"]).Handler(mcp.CallToolRequestParamsArguments{
	//	"Foo": "hello",
	//	"Bar": map[string]interface{}{
	//		"Hello": "world",
	//	},
	//})

	err := s.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}
