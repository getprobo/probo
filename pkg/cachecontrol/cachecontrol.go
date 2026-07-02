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

// Package cachecontrol parses HTTP Cache-Control header values as defined in
// RFC 9111 Section 5.2.
//
// The API and parsing approach are adapted from github.com/lestrrat-go/httpcc
// (MIT license, https://github.com/lestrrat-go/httpcc).
package cachecontrol

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxAge          = "max-age"
	MaxStale        = "max-stale"
	MinFresh        = "min-fresh"
	NoCache         = "no-cache"
	NoStore         = "no-store"
	NoTransform     = "no-transform"
	OnlyIfCached    = "only-if-cached"
	MustRevalidate  = "must-revalidate"
	Public          = "public"
	Private         = "private"
	ProxyRevalidate = "proxy-revalidate"
	SMaxAge         = "s-maxage"
)

type (
	TokenPair struct {
		Name  string
		Value string
	}

	TokenValuePolicy int

	directiveValidator interface {
		Validate(name string) TokenValuePolicy
	}

	directiveValidatorFn func(string) TokenValuePolicy
)

const (
	NoArgument TokenValuePolicy = iota
	TokenOnly
	QuotedStringOnly
	AnyTokenValue
)

func (fn directiveValidatorFn) Validate(name string) TokenValuePolicy {
	return fn(name)
}

func responseDirectiveValidator(name string) TokenValuePolicy {
	switch name {
	case MustRevalidate, NoStore, NoTransform, Public, ProxyRevalidate:
		return NoArgument
	case NoCache, Private:
		return QuotedStringOnly
	case MaxAge, SMaxAge:
		return TokenOnly
	default:
		return AnyTokenValue
	}
}

func requestDirectiveValidator(name string) TokenValuePolicy {
	switch name {
	case MaxAge, MaxStale, MinFresh:
		return TokenOnly
	case NoCache, NoStore, NoTransform, OnlyIfCached:
		return NoArgument
	default:
		return AnyTokenValue
	}
}

// ParseRequestDirective parses a single Cache-Control directive from a request.
func ParseRequestDirective(raw string) (*TokenPair, error) {
	return parseDirective(raw, directiveValidatorFn(requestDirectiveValidator))
}

// ParseResponseDirective parses a single Cache-Control directive from a response.
func ParseResponseDirective(raw string) (*TokenPair, error) {
	return parseDirective(raw, directiveValidatorFn(responseDirectiveValidator))
}

// ParseRequestDirectives parses Cache-Control directives from a request header.
func ParseRequestDirectives(header string) ([]*TokenPair, error) {
	return parseDirectives(header, ParseRequestDirective)
}

// ParseResponseDirectives parses Cache-Control directives from a response header.
func ParseResponseDirectives(header string) ([]*TokenPair, error) {
	return parseDirectives(header, ParseResponseDirective)
}

// ParseRequest parses the Cache-Control header value of an HTTP request.
func ParseRequest(header string) (*RequestDirective, error) {
	tokens, err := ParseRequestDirectives(header)
	if err != nil {
		return nil, fmt.Errorf("cannot parse request cache-control: %w", err)
	}

	dir := &RequestDirective{
		extensions: make(map[string]string),
	}

	for _, token := range tokens {
		name := strings.ToLower(token.Name)

		switch name {
		case MaxAge:
			seconds, err := parseDeltaSeconds(token.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse max-age: %w", err)
			}

			dir.maxAge = &seconds
		case MaxStale:
			if token.Value == "" {
				dir.maxStaleUnbounded = true
				break
			}

			seconds, err := parseDeltaSeconds(token.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse max-stale: %w", err)
			}

			dir.maxStale = &seconds
		case MinFresh:
			seconds, err := parseDeltaSeconds(token.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse min-fresh: %w", err)
			}

			dir.minFresh = &seconds
		case NoCache:
			dir.noCache = true
		case NoStore:
			dir.noStore = true
		case NoTransform:
			dir.noTransform = true
		case OnlyIfCached:
			dir.onlyIfCached = true
		default:
			dir.extensions[token.Name] = token.Value
		}
	}

	return dir, nil
}

// ParseResponse parses the Cache-Control header value of an HTTP response.
// When multiple max-age directives are present, the minimum value is kept
// per RFC 7234 Section 4.2.3.
func ParseResponse(header string) (*ResponseDirective, error) {
	tokens, err := ParseResponseDirectives(header)
	if err != nil {
		return nil, fmt.Errorf("cannot parse response cache-control: %w", err)
	}

	dir := &ResponseDirective{
		extensions: make(map[string]string),
	}

	for _, token := range tokens {
		name := strings.ToLower(token.Name)

		switch name {
		case MaxAge:
			seconds, err := parseDeltaSeconds(token.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse max-age: %w", err)
			}

			setMinimumUint64(&dir.maxAge, seconds)
		case MustRevalidate:
			dir.mustRevalidate = true
		case NoCache:
			dir.noCache = appendFields(dir.noCache, token.Value)
		case NoStore:
			dir.noStore = true
		case NoTransform:
			dir.noTransform = true
		case Public:
			dir.public = true
		case Private:
			dir.private = appendFields(dir.private, token.Value)
		case ProxyRevalidate:
			dir.proxyRevalidate = true
		case SMaxAge:
			seconds, err := parseDeltaSeconds(token.Value)
			if err != nil {
				return nil, fmt.Errorf("cannot parse s-maxage: %w", err)
			}

			setMinimumUint64(&dir.sMaxAge, seconds)
		default:
			dir.extensions[token.Name] = token.Value
		}
	}

	return dir, nil
}

func parseDirective(raw string, validator directiveValidator) (*TokenPair, error) {
	raw = strings.TrimSpace(raw)

	idx := strings.IndexByte(raw, '=')
	if idx == -1 {
		return &TokenPair{Name: raw}, nil
	}

	pair := &TokenPair{
		Name: strings.TrimSpace(raw[:idx]),
	}

	if len(raw) <= idx {
		return pair, nil
	}

	value := strings.TrimSpace(raw[idx+1:])

	switch validator.Validate(strings.ToLower(pair.Name)) {
	case TokenOnly:
		if value != "" && value[0] == '"' {
			return nil, fmt.Errorf("invalid value for %s: quoted string not allowed", pair.Name)
		}
	case QuotedStringOnly:
		if value == "" {
			break
		}

		if value[0] != '"' {
			return nil, fmt.Errorf("invalid value for %s: bare token not allowed", pair.Name)
		}

		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: malformed quoted string", pair.Name)
		}

		value = unquoted
	case AnyTokenValue:
		if value != "" && value[0] == '"' {
			unquoted, err := strconv.Unquote(value)
			if err != nil {
				return nil, fmt.Errorf("invalid value for %s: malformed quoted string", pair.Name)
			}

			value = unquoted
		}
	case NoArgument:
		if value != "" {
			return nil, fmt.Errorf("received argument to directive %s", pair.Name)
		}
	}

	pair.Value = value

	return pair, nil
}

func parseDirectives(header string, parse func(string) (*TokenPair, error)) ([]*TokenPair, error) {
	scanner := bufio.NewScanner(strings.NewReader(header))
	scanner.Split(scanCommaSeparatedWords)

	var tokens []*TokenPair

	for scanner.Scan() {
		token, err := parse(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("cannot parse directive #%d: %w", len(tokens)+1, err)
		}

		tokens = append(tokens, token)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("cannot scan cache-control directives: %w", err)
	}

	return tokens, nil
}

func appendFields(fields []string, raw string) []string {
	scanner := bufio.NewScanner(strings.NewReader(raw))
	scanner.Split(scanCommaSeparatedWords)

	for scanner.Scan() {
		fields = append(fields, scanner.Text())
	}

	return fields
}

func setMinimumUint64(target **uint64, value uint64) {
	if *target == nil || value < **target {
		v := value
		*target = &v
	}
}

func parseDeltaSeconds(raw string) (uint64, error) {
	if raw == "" {
		return 0, fmt.Errorf("empty delta-seconds")
	}

	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid delta-seconds %q", raw)
		}
	}

	seconds, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid delta-seconds %q: %w", raw, err)
	}

	return seconds, nil
}

func secondsToDuration(seconds uint64) time.Duration {
	const maxSeconds = uint64(math.MaxInt64 / int64(time.Second))
	if seconds > maxSeconds {
		return time.Duration(math.MaxInt64)
	}

	return time.Duration(seconds) * time.Second
}

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}

		return false
	}

	if '\u2000' <= r && r <= '\u200a' {
		return true
	}

	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}

	return false
}

func scanCommaSeparatedWords(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := 0

	for width := 0; start < len(data); start += width {
		var r rune

		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}

	var ws int

	inQuotes := false

	for width, i := 0, start; i < len(data); i += width {
		var r rune

		r, width = utf8.DecodeRune(data[i:])

		switch {
		case r == '"':
			inQuotes = !inQuotes
			ws = 0
		case isSpace(r) && !inQuotes:
			ws++
		case r == ',' && !inQuotes:
			return i + width, data[start : i-ws], nil
		default:
			ws = 0
		}
	}

	if atEOF && len(data) > start {
		return len(data), data[start : len(data)-ws], nil
	}

	return start, nil, nil
}
