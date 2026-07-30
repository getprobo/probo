package mcp

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MustUnmarshalSchema unmarshals a JSON schema string into a jsonschema.Schema
// Panics if unmarshaling fails, providing compile-time safety for schema definitions
func MustUnmarshalSchema(schemaJSON string) *jsonschema.Schema {
	var schema jsonschema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		panic("invalid schema JSON: " + err.Error())
	}
	return &schema
}

// PromptHandlerFor is a typed prompt handler that accepts structured arguments.
// Similar to mcp.ToolHandlerFor, this allows prompts to work with typed Go structs
// instead of raw map[string]string.
//
// The Args type parameter must be a struct or map type. Arguments will be automatically
// unmarshaled from the prompt request's Arguments map into the Args type.
//
// Example:
//
//	type TaskArgs struct {
//	    Topic    string `json:"topic"`
//	    Detailed bool   `json:"detailed"`
//	}
//
//	func (r *Resolver) TaskHelpPrompt(ctx context.Context, req *mcp.GetPromptRequest, args TaskArgs) (*mcp.GetPromptResult, error) {
//	    // args.Topic and args.Detailed are already parsed
//	    return &mcp.GetPromptResult{...}, nil
//	}
type PromptHandlerFor[Args any] func(context.Context, *mcp.GetPromptRequest, Args) (*mcp.GetPromptResult, error)

// AddPrompt is a generic wrapper around Server.AddPrompt that provides type-safe argument handling.
// It automatically converts the prompt arguments from map[string]string into the typed Args parameter.
//
// This matches the ergonomics of mcp.AddTool for a consistent API experience across tools and prompts.
//
// The Args type must be a struct with string fields or map[string]string. If it's a struct, the fields
// will be populated from the arguments map based on their json tags.
//
// Example:
//
//	type HelpArgs struct {
//	    Topic string `json:"topic"`
//	}
//
//	mcp.AddPrompt(server, &mcp.Prompt{
//	    Name: "help",
//	    Description: "Get help",
//	}, resolver.HelpPrompt) // HelpPrompt receives typed HelpArgs
func AddPrompt[Args any](s *mcp.Server, p *mcp.Prompt, h PromptHandlerFor[Args]) {
	s.AddPrompt(p, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		var args Args

		// Convert map[string]string to typed Args using JSON as the intermediary.
		// This properly handles json tags and field mapping.
		if len(req.Params.Arguments) > 0 {
			argsBytes, err := json.Marshal(req.Params.Arguments)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(argsBytes, &args); err != nil {
				return nil, err
			}
		}

		return h(ctx, req, args)
	})
}
