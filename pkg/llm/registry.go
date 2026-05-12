// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package llm

//go:generate go run go.probo.inc/probo/internal/cmd/genmodels

import (
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
// "gpt-5.4"). Returns false if the model is not in the registry.
func (r *Registry) Lookup(modelID string) (ModelDefinition, bool) {
	if m, ok := r.byID[modelID]; ok {
		return *m, true
	}
	if m, ok := r.byID[normalizeModelID(modelID)]; ok {
		return *m, true
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
