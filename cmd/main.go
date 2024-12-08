package main

import mcp "github.com/metoro-io/mcp-golang"

type HelloType struct {
	Hello string `json:"hello" mcp:"description:'This is hello, you need to pass it'"`
}
type MyFunctionsArguments struct {
	Foo string    `json:"foo" mcp:"description:'This is foo, you need to pass it'"`
	Bar HelloType `json:"bar" mcp:"description:'This is bar, you need to pass it'"`
}

func main() {
	s := mcp.NewServer(mcp.NewStdioServerTransport())
	s.Tool("test", "Test tool's description", func(arguments MyFunctionsArguments) (mcp.ToolResponse, error) {
		h := arguments.Bar.Hello
		// ... handle the tool logic
		println(arguments.Foo)
		return mcp.ToolResponse{Result: h}, nil
	})

	(*s.Tools["test"]).Handler(mcp.CallToolRequestParamsArguments{
		"Foo": "hello",
		"Bar": map[string]interface{}{
			"Hello": "world",
		},
	})
}
