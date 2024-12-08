package mcp

// Here we define the actual MCP server that users will create and run
// A server can be passed a number of handlers to handle requests from clients
// Additionally it can be parametrized by a transport. This transport will be used to actually send and receive messages.
// So for example if the stdio transport is used, the server will read from stdin and write to stdout
// If the SSE transport is used, the server will send messages over an SSE connection and receive messages from HTTP POST requests.

// The interface that we're looking to support is something like [gin](https://github.com/gin-gonic/gin)s interface
// Example use would be:

// func main() {
//     transport := mcp.NewStdioTransport()
//     server := mcp.NewServer(transport)
//

//       type HelloType struct {
//           Hello string `mcp:"description:'description',valudation:maxLength(10)`"`
//       }
//       type MyFunctionsArguments struct {
//           Foo string `mcp:"description:'description',validation:maxLength(10)`"`
//           Bar HelloType `mcp:"description:'description'`"`
//       }
//     server.Tool("test", "Test tool's description", MyFunctionsArguments{}, func(argument MyFunctionsArguments) (ToolResponse, error) {
//        let h := argument.Bar.Hello
//     })
//
//       arguments := NewObject(new Map[string]Argument]{
//		   "foo", NewString("description", NewStringValidation(Required, MaxLength(10))),
//		   "bar", NewObject(new Map[string]Argument]{
//		     "hello", NewStringValidation(Required, MaxLength(10))),
//		   }, NewObjectValidation(Required))
//		 )
//     server.Tool("test", "Test tool's description", arguments, func(argument Object) (ToolResponse, error) {
//        let bar, err := argument.GetString("bar")
//        if err != nil {
//            return nil, err
//        }
//        let h, err := bar.GetString("hello")
//		  if err != nil {
//			  return nil, err
//		  }
//     })
//
//
//     // Send a message
//     transport.Send(map[string]interface{}{
//         "jsonrpc": "2.0",
//         "method": "test",
//         "params": map[string]interface{}{},
//     })
// }
