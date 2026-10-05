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

package toolselect

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Tool struct {
		Name        string `yaml:"name"`
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	}

	specFile struct {
		Tools []Tool `yaml:"tools"`
	}
)

func LoadSpec(path string) ([]Tool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read MCP specification: %w", err)
	}

	var spec specFile
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("cannot parse MCP specification: %w", err)
	}

	if len(spec.Tools) == 0 {
		return nil, fmt.Errorf("cannot parse MCP specification: no tools")
	}

	return spec.Tools, nil
}
