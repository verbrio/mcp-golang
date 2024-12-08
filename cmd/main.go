package main

import mcp "github.com/metoro-io/mcp-golang"

type HelloType struct {
	Hello string `mcp:"description:'description',validation:maxLength(10)"`
}
type MyFunctionsArguments struct {
	Foo string    `mcp:"description:'description',validation:maxLength(10)"`
	Bar HelloType `mcp:"description:'description'"`
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
