// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mcputils

import (
	"fmt"
	"reflect"
	"slices"
	"unsafe"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisteredToolNames returns the tools currently registered on server.
//
// The MCP SDK keeps that set unexported. Call this only while the server is
// still being built, before it serves requests.
func RegisteredToolNames(server *mcp.Server) ([]string, error) {
	if server == nil {
		return nil, fmt.Errorf("cannot list MCP tools: server is nil")
	}

	toolsField := reflect.ValueOf(server).Elem().FieldByName("tools")
	if !toolsField.IsValid() {
		return nil, fmt.Errorf("cannot list MCP tools: server has no tools field")
	}

	tools := reflect.NewAt(toolsField.Type(), unsafe.Pointer(toolsField.UnsafeAddr())).Elem()
	if tools.Kind() != reflect.Pointer || tools.IsNil() {
		return nil, fmt.Errorf("cannot list MCP tools: tool set is nil")
	}

	featuresField := tools.Elem().FieldByName("features")
	if !featuresField.IsValid() {
		return nil, fmt.Errorf("cannot list MCP tools: tool set has no features field")
	}

	features := reflect.NewAt(featuresField.Type(), unsafe.Pointer(featuresField.UnsafeAddr())).Elem()
	if features.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot list MCP tools: features are not a map")
	}

	names := make([]string, 0, features.Len())
	for _, key := range features.MapKeys() {
		name := key.String()
		if name == "" {
			return nil, fmt.Errorf("cannot list MCP tools: empty tool name")
		}

		names = append(names, name)
	}

	slices.Sort(names)

	return names, nil
}

// RestrictTools removes every registered tool for which allow returns false.
func RestrictTools(server *mcp.Server, allow func(name string) bool) error {
	if allow == nil {
		return fmt.Errorf("cannot restrict MCP tools: allow is nil")
	}

	names, err := RegisteredToolNames(server)
	if err != nil {
		return err
	}

	drop := make([]string, 0, len(names))
	for _, name := range names {
		if !allow(name) {
			drop = append(drop, name)
		}
	}

	server.RemoveTools(drop...)

	return nil
}
