package main

import (
	"fmt"
	"github.com/metoro-io/mcp-golang/server"
	"github.com/metoro-io/mcp-golang/server/tools"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

type Content struct {
	Title       string  `json:"title" jsonschema:"description=The title to submit"`
	Description *string `json:"description,omitempty" jsonschema:"description=The description to submit"`
}
type MyFunctionsArguments struct {
	Submitter string  `json:"submitter" jsonschema:"description=The name of the thing calling this tool (openai, google, claude, etc)"`
	Content   Content `json:"content" jsonschema:"description=The content of the message"`
}

type ToggleLights struct {
	EntityID string `json:"entity_id,omitempty"`
}

func main() {
	done := make(chan struct{})

	s := server.NewServer(stdio.NewStdioServerTransport())
	err := s.RegisterTool("hello", "Say hello to a person", func(arguments MyFunctionsArguments) (*tools.ToolResponse, error) {
		return tools.NewToolReponse(tools.NewToolTextResponseContent(fmt.Sprintf("Hello, %s!", arguments.Submitter))), nil
	})
	if err != nil {
		panic(err)
	}

	err = s.Serve()
	if err != nil {
		panic(err)
	}

	<-done
}
