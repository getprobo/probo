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

package cookiebanner

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/validator"
)

const (
	// Minted by packages/cookie-banner-tcf encodeTCString (CMP ID 4095, 2.3).
	validTCStringV23 = "CQqvPYAQqvPYA__ABBENAqFgAAAAAAAAAAAAAAAAAAAA.IAaQAQAaAAAA.YAAAAAAAAAAA"
	coreOnlyTCString = "CPzqA4APzqA4AEsAAAENAwCAAAAAAAAAAAAAAAAAAAAA"
)

func TestValidateConsentTC_Scenario(t *testing.T) {
	t.Parallel()

	gdpr := RegulationGDPR
	ukGDPR := RegulationUKGDPR
	ccpa := RegulationCCPA
	wrongCmpID := mintTCString(tcCookieVersion, 2, true)
	customCmpID := mintTCString(tcCookieVersion, 123, true)

	tests := []struct {
		name       string
		tcfEnabled bool
		regulation *Regulation
		tc         *string
		cmpID      int
		wantField  string
		wantCode   validator.ErrorCode
	}{
		{
			name:       "allows empty tc when tcf is off",
			tcfEnabled: false,
		},
		{
			name:       "rejects tc when tcf is off",
			tcfEnabled: false,
			tc:         new(validTCStringV23),
			wantField:  "tc",
			wantCode:   validator.ErrorCodeInvalidFormat,
		},
		{
			name:       "requires tc for gdpr when tcf is on",
			tcfEnabled: true,
			regulation: &gdpr,
			wantField:  "tc",
			wantCode:   validator.ErrorCodeRequired,
		},
		{
			name:       "requires tc for uk gdpr when tcf is on",
			tcfEnabled: true,
			regulation: &ukGDPR,
			wantField:  "tc",
			wantCode:   validator.ErrorCodeRequired,
		},
		{
			name:       "allows empty tc for ccpa when tcf is on",
			tcfEnabled: true,
			regulation: &ccpa,
		},
		{
			name:       "allows empty tc when regulation is unset",
			tcfEnabled: true,
		},
		{
			name:       "rejects garbage",
			tcfEnabled: true,
			regulation: &gdpr,
			tc:         new("not-a-tc-string"),
			wantField:  "tc",
			wantCode:   validator.ErrorCodeInvalidFormat,
		},
		{
			name:       "rejects core-only string",
			tcfEnabled: true,
			regulation: &gdpr,
			tc:         new(coreOnlyTCString),
			wantField:  "tc",
			wantCode:   validator.ErrorCodeInvalidFormat,
		},
		{
			name:       "rejects wrong cmp id",
			tcfEnabled: true,
			regulation: &gdpr,
			tc:         &wrongCmpID,
			wantField:  "tc",
			wantCode:   validator.ErrorCodeInvalidFormat,
		},
		{
			name:       "accepts a 2.3 string for gdpr",
			tcfEnabled: true,
			regulation: &gdpr,
			tc:         new(validTCStringV23),
		},
		{
			name:       "accepts a string minted for a custom cmp id",
			tcfEnabled: true,
			regulation: &gdpr,
			cmpID:      123,
			tc:         &customCmpID,
		},
		{
			name:       "rejects the default cmp id when the instance uses another",
			tcfEnabled: true,
			regulation: &gdpr,
			cmpID:      123,
			tc:         new(validTCStringV23),
			wantField:  "tc",
			wantCode:   validator.ErrorCodeInvalidFormat,
		},
		{
			name:       "accepts a 2.3 string for ccpa when present",
			tcfEnabled: true,
			regulation: &ccpa,
			tc:         new(validTCStringV23),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				err := validateConsentTC(tt.tcfEnabled, tt.regulation, tt.tc, tt.cmpID)
				if tt.wantField == "" {
					assert.NoError(t, err)
					return
				}

				require.Error(t, err)

				validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
				require.True(t, ok)

				fieldErrors := validationErrors.ByField(tt.wantField)
				require.NotEmpty(t, fieldErrors)
				assert.Equal(t, tt.wantCode, fieldErrors[0].Code)
			},
		)
	}
}

func TestParseTCString_AcceptsEncoderFixture(t *testing.T) {
	t.Parallel()

	assert.NoError(t, parseTCString(validTCStringV23, DefaultTCFCmpID))
}

func mintTCString(version, cmpID uint, disclosed bool) string {
	var core bitWriter
	core.write(version, tcCookieVersionBits)
	core.write(0, tcCreatedBits)
	core.write(0, tcLastUpdatedBits)
	core.write(cmpID, tcCmpIDBits)

	encoded := core.encode()
	if !disclosed {
		return encoded
	}

	var segment bitWriter
	segment.write(tcDisclosedVendorsSegment, tcSegmentTypeBits)

	return encoded + "." + segment.encode()
}

type bitWriter struct {
	bits []bool
}

func (w *bitWriter) write(value uint, n int) {
	for i := n - 1; i >= 0; i-- {
		w.bits = append(w.bits, value&(1<<uint(i)) != 0)
	}
}

func (w *bitWriter) encode() string {
	data := make([]byte, (len(w.bits)+7)/8)
	for i, bit := range w.bits {
		if bit {
			data[i/8] |= 1 << (7 - (i % 8))
		}
	}

	return base64.RawURLEncoding.EncodeToString(data)
}
