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
	"strings"

	"go.probo.inc/probo/pkg/validator"
)

const (
	tcCookieVersionBits       = 6
	tcCreatedBits             = 36
	tcLastUpdatedBits         = 36
	tcCmpIDBits               = 12
	tcSegmentTypeBits         = 3
	tcCookieVersion           = 2
	tcDisclosedVendorsSegment = 1
)

var errInvalidTCString = errors.New("invalid tc string")

func validateConsentTC(tcfEnabled bool, regulation *Regulation, tc *string) error {
	v := validator.New()

	if !tcfEnabled {
		if optionalNonEmptyString(tc) != nil {
			v.Check(tc, "tc", tcForbiddenWhenDisabled())
		}

		return v.Error()
	}

	if tcfRequiresTC(regulation) {
		v.Check(tc, "tc", validator.Required(), validator.NotEmpty())
	}

	if optionalNonEmptyString(tc) != nil {
		v.Check(tc, "tc", tcfStringFormat())
	}

	return v.Error()
}

func tcfServesGVL(regulation Regulation) bool {
	return regulation == RegulationGDPR || regulation == RegulationUKGDPR
}

func tcfRequiresTC(regulation *Regulation) bool {
	if regulation == nil {
		return false
	}

	return tcfServesGVL(*regulation)
}

func tcForbiddenWhenDisabled() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		if value == nil {
			return nil
		}

		return &validator.ValidationError{
			Code:    validator.ErrorCodeInvalidFormat,
			Message: "must be empty when TCF is not enabled",
		}
	}
}

func tcfStringFormat() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		if value == nil {
			return nil
		}

		encoded, ok := value.(string)
		if !ok {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeInvalidFormat,
				Message: "must be a string",
			}
		}

		if err := parseTCString(encoded); err != nil {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeInvalidFormat,
				Message: "must be a valid TCF 2.3 consent string",
			}
		}

		return nil
	}
}

func parseTCString(encoded string) error {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return errInvalidTCString
	}

	segments := strings.Split(encoded, ".")

	version, cmpID, err := parseTCCore(segments[0])
	if err != nil {
		return err
	}

	if version != tcCookieVersion {
		return errInvalidTCString
	}

	if cmpID != tcfCmpID {
		return errInvalidTCString
	}

	if !hasDisclosedVendorsSegment(segments[1:]) {
		return errInvalidTCString
	}

	return nil
}

func parseTCCore(segment string) (uint, uint, error) {
	data, err := decodeTCSegment(segment)
	if err != nil {
		return 0, 0, err
	}

	r := bitReader{data: data}

	version, err := r.read(tcCookieVersionBits)
	if err != nil {
		return 0, 0, err
	}

	if _, err := r.read(tcCreatedBits + tcLastUpdatedBits); err != nil {
		return 0, 0, err
	}

	cmpID, err := r.read(tcCmpIDBits)
	if err != nil {
		return 0, 0, err
	}

	return version, cmpID, nil
}

func hasDisclosedVendorsSegment(segments []string) bool {
	for _, segment := range segments {
		data, err := decodeTCSegment(segment)
		if err != nil {
			continue
		}

		r := bitReader{data: data}

		segmentType, err := r.read(tcSegmentTypeBits)
		if err != nil {
			continue
		}

		if segmentType == tcDisclosedVendorsSegment {
			return true
		}
	}

	return false
}

func decodeTCSegment(segment string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, errInvalidTCString
	}

	if len(data) == 0 {
		return nil, errInvalidTCString
	}

	return data, nil
}

type bitReader struct {
	data []byte
	pos  int
}

func (r *bitReader) read(n int) (uint, error) {
	if n <= 0 {
		return 0, errInvalidTCString
	}

	var value uint

	for range n {
		byteIdx := r.pos / 8
		if byteIdx >= len(r.data) {
			return 0, errInvalidTCString
		}

		bit := 7 - (r.pos % 8)

		value <<= 1
		if r.data[byteIdx]&(1<<bit) != 0 {
			value |= 1
		}

		r.pos++
	}

	return value, nil
}
