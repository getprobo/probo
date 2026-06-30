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
	"fmt"
	"os"
	"path/filepath"

	"go.probo.inc/probo/pkg/probodconfig"
	"sigs.k8s.io/yaml"
)

func WriteConfig(cfg *probodconfig.FullConfig, path string) error {
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

	data, err = yaml.Marshal(pruned)
	if err != nil {
		return fmt.Errorf("marshal pruned config: %w", err)
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
