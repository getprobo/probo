// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.probo.inc/probo/pkg/probodconfig"
	"sigs.k8s.io/yaml"
)

type Format string

const (
	FormatYAML Format = "yaml"
	FormatJSON Format = "json"
)

func WriteConfig(cfg *probodconfig.FullConfig, path string, format Format) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	var tree any
	if err := yaml.Unmarshal(data, &tree); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	pruned := pruneEmptyStrings(tree)

	switch format {
	case FormatJSON:
		data, err = json.MarshalIndent(pruned, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal pruned config as json: %w", err)
		}
	case FormatYAML:
		data, err = yaml.Marshal(pruned)
		if err != nil {
			return fmt.Errorf("marshal pruned config as yaml: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config format: %q", format)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func pruneEmptyStrings(value any) any {
	switch v := value.(type) {
	case map[string]any:
		for key, child := range v {
			if s, ok := child.(string); ok && s == "" {
				delete(v, key)

				continue
			}

			v[key] = pruneEmptyStrings(child)
		}

		return v
	case []any:
		for i, child := range v {
			v[i] = pruneEmptyStrings(child)
		}

		return v
	default:
		return v
	}
}
