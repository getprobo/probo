// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package saml

import (
	"errors"
	"testing"
	"time"

	"github.com/crewjam/saml"
	"go.probo.inc/probo/pkg/coredata"
)

func TestValidateAssertionRejectsEmptySAMLSubject(t *testing.T) {
	t.Parallel()

	s := &Service{}
	config := &coredata.SAMLConfiguration{IdPEntityID: "https://idp.example"}
	now := time.Now()

	tests := []struct {
		name  string
		value string
	}{
		{name: "empty", value: ""},
		{name: "whitespace", value: "   "},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				err := s.validateAssertion(
					&saml.Assertion{
						ID: "assertion-1",
						Issuer: saml.Issuer{
							Value: config.IdPEntityID,
						},
						Subject: &saml.Subject{
							NameID: &saml.NameID{Value: tt.value},
						},
						Conditions: &saml.Conditions{
							NotOnOrAfter: now.Add(time.Hour),
						},
					},
					config,
					now,
				)
				if err == nil {
					t.Fatal("expected error for empty SAML subject")
				}

				if _, ok := errors.AsType[*ErrSAMLSubjectRequired](err); !ok {
					t.Fatalf("expected *ErrSAMLSubjectRequired, got %T: %v", err, err)
				}

				if got := err.Error(); got != "NameID value is required" {
					t.Fatalf("unexpected error: %v", err)
				}
			},
		)
	}
}

func TestHasSAMLSubject(t *testing.T) {
	t.Parallel()

	empty := ""
	whitespace := "  "
	value := "user@example.com"

	tests := []struct {
		name     string
		subject  *string
		expected bool
	}{
		{name: "nil", subject: nil, expected: false},
		{name: "empty", subject: &empty, expected: false},
		{name: "whitespace", subject: &whitespace, expected: false},
		{name: "present", subject: &value, expected: true},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				identity := &coredata.Identity{SAMLSubject: tt.subject}

				if got := hasSAMLSubject(identity); got != tt.expected {
					t.Fatalf("hasSAMLSubject() = %v, want %v", got, tt.expected)
				}
			},
		)
	}
}
