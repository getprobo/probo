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

package llm

//go:generate go run go.probo.inc/probo/internal/cmd/genmodels

import (
	"regexp"
	"strings"
	"sync"
)

type (
	// ModelDefinition describes a model's capabilities and limits.
	ModelDefinition struct {
		ID              string
		Name            string
		ContextLength   int
		MaxOutputTokens int
		Supports        SupportedParameters
	}

	// SupportedParameters tracks which request parameters a model accepts.
	SupportedParameters struct {
		Temperature       bool
		TopP              bool
		TopK              bool
		FrequencyPenalty  bool
		PresencePenalty   bool
		Stop              bool
		Seed              bool
		MaxTokens         bool
		ToolChoice        bool
		ParallelToolCalls bool
		ResponseFormat    bool
		StructuredOutputs bool
		Reasoning         bool
	}

	// Registry provides model capability lookups.
	Registry struct {
		byID map[string]*ModelDefinition
	}
)

var (
	defaultRegistry     *Registry
	defaultRegistryOnce sync.Once
)

// NewRegistry builds a registry from the given model definitions.
func NewRegistry(models map[string]ModelDefinition) *Registry {
	r := &Registry{byID: make(map[string]*ModelDefinition, len(models)*3)}
	for id, m := range models {
		m.ID = id
		r.index(&m)
	}

	return r
}

// DefaultRegistry returns the cached registry built from generated model data.
func DefaultRegistry() *Registry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry(generatedModels)
	})

	return defaultRegistry
}

// Lookup finds a model by ID. It accepts both provider-prefixed IDs
// ("anthropic/claude-opus-4.6") and bare provider IDs ("claude-opus-4-6",
// "gpt-5.4"). Dated provider snapshots ("gpt-5-nano-2025-08-07") fall
// back to their undated base model ("gpt-5-nano"). Returns false if the
// model is not in the registry.
func (r *Registry) Lookup(modelID string) (ModelDefinition, bool) {
	if m, ok := r.byID[modelID]; ok {
		return *m, true
	}

	if m, ok := r.byID[normalizeModelID(modelID)]; ok {
		return *m, true
	}

	if base := stripModelDateSuffix(modelID); base != modelID {
		if m, ok := r.byID[base]; ok {
			return *m, true
		}

		if m, ok := r.byID[normalizeModelID(base)]; ok {
			return *m, true
		}
	}

	return ModelDefinition{}, false
}

// Provider returns the provider prefix from the model ID (e.g. "anthropic"
// from "anthropic/claude-opus-4.6").
func (m *ModelDefinition) Provider() string {
	provider, _, _ := strings.Cut(m.ID, "/")
	return provider
}

func (r *Registry) index(m *ModelDefinition) {
	r.byID[m.ID] = m

	if idx := strings.IndexByte(m.ID, '/'); idx >= 0 {
		r.byID[m.ID[idx+1:]] = m
	}

	normalized := normalizeModelID(m.ID)
	if normalized != m.ID {
		r.byID[normalized] = m
	}
}

func normalizeModelID(id string) string {
	if idx := strings.IndexByte(id, '/'); idx >= 0 {
		id = id[idx+1:]
	}

	return strings.ReplaceAll(id, ".", "-")
}

// modelDateSuffix matches a trailing provider snapshot date such as the
// "-2025-08-07" in "gpt-5-nano-2025-08-07".
var modelDateSuffix = regexp.MustCompile(`-\d{4}-\d{2}-\d{2}$`)

// stripModelDateSuffix removes a trailing dated-snapshot suffix from a
// model ID, leaving the base model ID unchanged when none is present.
func stripModelDateSuffix(id string) string {
	return modelDateSuffix.ReplaceAllString(id, "")
}
