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

package esign

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/validator"
)

func TestAcceptSignature_EmptySignerFullName(t *testing.T) {
	t.Parallel()

	s := &Service{}
	_, err := s.AcceptSignature(
		context.Background(),
		nil,
		&AcceptSignatureRequest{SignerFullName: " \t "},
	)

	require.Error(t, err)
	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("signer_full_name"))
}

func TestCreateAndAcceptSignature_EmptySignerFullName(t *testing.T) {
	t.Parallel()

	s := &Service{}
	_, err := s.CreateAndAcceptSignature(
		context.Background(),
		nil,
		&CreateAndAcceptSignatureRequest{},
	)

	require.Error(t, err)
	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("signer_full_name"))
}

func TestRecordEvent_EmptyActorFullName(t *testing.T) {
	t.Parallel()

	s := &Service{}
	err := s.RecordEvent(
		context.Background(),
		nil,
		&RecordEventRequest{ActorFullName: " \t "},
	)

	require.Error(t, err)
	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("actor_full_name"))
}

func TestSignatureRequestsValidate_SignerFullName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		validate func() error
	}{
		{
			name: "accept signature",
			validate: func() error {
				return (AcceptSignatureRequest{SignerFullName: "Ada Lovelace"}).Validate()
			},
		},
		{
			name: "create and accept signature",
			validate: func() error {
				return (CreateAndAcceptSignatureRequest{SignerFullName: "Ada Lovelace"}).Validate()
			},
		},
		{
			name: "record event",
			validate: func() error {
				return (RecordEventRequest{ActorFullName: "Ada Lovelace"}).Validate()
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.NoError(t, tt.validate())
			},
		)
	}
}
