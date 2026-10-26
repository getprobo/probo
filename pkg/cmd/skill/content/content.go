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

// Package content is the agent skills embedded in the prb binary.
package content

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed prb-auth prb-records using-prb
var files embed.FS

// Skill is one embedded agent skill.
type Skill struct {
	Name        string
	Description string
}

// List returns the embedded skills, sorted by name.
func List() ([]Skill, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("cannot read embedded skills: %w", err)
	}

	skills := make([]Skill, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skill, err := load(entry.Name())
		if err != nil {
			return nil, err
		}

		skills = append(skills, skill)
	}

	slices.SortFunc(skills, func(a, b Skill) int {
		return strings.Compare(a.Name, b.Name)
	})

	return skills, nil
}

// Read returns a file from a skill. An empty rel reads SKILL.md.
func Read(name, rel string) ([]byte, error) {
	if _, err := load(name); err != nil {
		return nil, err
	}

	if rel == "" {
		rel = "SKILL.md"
	}

	rel = path.Clean(rel)
	if !fs.ValidPath(rel) {
		return nil, fmt.Errorf("invalid skill path %q", rel)
	}

	data, err := files.ReadFile(path.Join(name, rel))
	if err != nil {
		return nil, fmt.Errorf("cannot read skill %s file %s: %w", name, rel, err)
	}

	return data, nil
}

// Install copies every embedded skill into dir. Existing files with the
// same contents are left in place. Different contents are replaced only
// when force is set.
func Install(dir string, force bool) ([]string, error) {
	skills, err := List()
	if err != nil {
		return nil, err
	}

	written := make([]string, 0, len(skills))
	for _, skill := range skills {
		paths, err := installSkill(dir, skill.Name, force)
		if err != nil {
			return nil, err
		}

		written = append(written, paths...)
	}

	return written, nil
}

func installSkill(dir, name string, force bool) ([]string, error) {
	var (
		relPaths []string
		payloads [][]byte
	)

	err := fs.WalkDir(files, name, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		prefix := name + "/"
		if !strings.HasPrefix(p, prefix) {
			return fmt.Errorf("cannot resolve skill path %s", p)
		}

		rel := strings.TrimPrefix(p, prefix)

		if !fs.ValidPath(rel) {
			return fmt.Errorf("invalid skill path %q", rel)
		}

		data, err := files.ReadFile(p)
		if err != nil {
			return fmt.Errorf("cannot read skill file %s: %w", p, err)
		}

		relPaths = append(relPaths, rel)
		payloads = append(payloads, data)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot walk skill %s: %w", name, err)
	}

	destinations := make([]string, len(relPaths))
	for i, rel := range relPaths {
		dst := filepath.Join(dir, name, filepath.FromSlash(rel))
		destinations[i] = dst

		existing, err := os.ReadFile(dst)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("cannot read %s: %w", dst, err)
		}

		if string(existing) == string(payloads[i]) {
			continue
		}

		if !force {
			return nil, fmt.Errorf("skill file %s already exists; pass --force to replace it", dst)
		}
	}

	for i, dst := range destinations {
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, fmt.Errorf("cannot create skill directory: %w", err)
		}

		if err := os.WriteFile(dst, payloads[i], 0o644); err != nil {
			return nil, fmt.Errorf("cannot write skill file %s: %w", dst, err)
		}
	}

	return destinations, nil
}

func load(name string) (Skill, error) {
	data, err := files.ReadFile(path.Join(name, "SKILL.md"))
	if err != nil {
		return Skill{}, fmt.Errorf("unknown skill %q", name)
	}

	front, ok := frontmatter(string(data))
	if !ok {
		return Skill{}, fmt.Errorf("skill %s is missing frontmatter", name)
	}

	fields := map[string]string{}
	for line := range strings.SplitSeq(front, "\n") {
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}

		fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	if fields["name"] == "" || fields["description"] == "" {
		return Skill{}, fmt.Errorf("skill %s is missing name or description", name)
	}

	if fields["name"] != name {
		return Skill{}, fmt.Errorf("skill directory %s declares name %s", name, fields["name"])
	}

	return Skill{
		Name:        fields["name"],
		Description: fields["description"],
	}, nil
}

func frontmatter(raw string) (string, bool) {
	const sep = "---\n"

	if !strings.HasPrefix(raw, sep) {
		return "", false
	}

	rest := strings.TrimPrefix(raw, sep)
	front, _, ok := strings.Cut(rest, "\n---\n")

	return front, ok
}
