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
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

type Hit struct {
	Name    string
	Score   float64
	Matches int
}

// Rank orders tools for a natural-language task. useDescription includes tool
// descriptions; the name-only mode is the baseline that ignores them.
func Rank(tools []Tool, task string, useDescription bool) []Hit {
	taskTokens := tokenize(task)
	docs := make([]map[string]struct{}, len(tools))
	names := make([]map[string]struct{}, len(tools))
	df := map[string]int{}

	for i, tool := range tools {
		names[i] = tokenSet(tokenize(tool.Name + " " + tool.Title))
		body := tool.Name + " " + tool.Title
		if useDescription {
			body += " " + tool.Description
		}

		docs[i] = tokenSet(tokenize(body))
		for token := range docs[i] {
			df[token]++
		}
	}

	n := float64(len(tools))
	idf := make(map[string]float64, len(df))
	for token, count := range df {
		idf[token] = math.Log((n+1)/(float64(count)+1)) + 1
	}

	hits := make([]Hit, 0, len(tools))
	for i, tool := range tools {
		matched, total := score(taskTokens, names[i], docs[i], idf)
		hits = append(hits, Hit{
			Name:    tool.Name,
			Score:   total,
			Matches: matched,
		})
	}

	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Matches != hits[j].Matches {
			return hits[i].Matches > hits[j].Matches
		}

		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}

		return hits[i].Name < hits[j].Name
	})

	return hits
}

func score(taskTokens []string, nameTokens, docTokens map[string]struct{}, idf map[string]float64) (int, float64) {
	seen := make(map[string]struct{}, len(taskTokens))
	total := 0.0
	matched := 0

	for _, token := range taskTokens {
		if _, ok := seen[token]; ok {
			continue
		}

		seen[token] = struct{}{}

		weight := 0.0
		if _, ok := nameTokens[token]; ok {
			weight += 0.25
		}

		if _, ok := docTokens[token]; ok {
			weight += 1
			matched++
		} else if weight > 0 {
			matched++
		}

		if weight == 0 {
			continue
		}

		total += weight * idf[token]
	}

	return matched, total
}

func tokenSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}

	return set
}

var camelBoundary = regexp.MustCompile(`([a-z0-9])([A-Z])`)

func tokenize(value string) []string {
	value = camelBoundary.ReplaceAllString(value, "${1} ${2}")
	value = strings.ToLower(value)

	fields := strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 3 || stopwords[field] {
			continue
		}

		tokens = append(tokens, field)
	}

	return tokens
}

var stopwords = map[string]bool{
	"and": true, "are": true, "but": true, "call": true, "for": true,
	"from": true, "has": true, "have": true, "into": true, "its": true,
	"not": true, "one": true, "only": true, "pass": true, "that": true,
	"the": true, "they": true, "this": true, "use": true, "when": true,
	"with": true, "you": true, "your": true,
}
