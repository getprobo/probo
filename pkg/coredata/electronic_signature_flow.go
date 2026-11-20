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

package coredata

import (
	"encoding"
	"fmt"
)

type ElectronicSignatureFlow string

const (
	// ElectronicSignatureFlowSignature is the original flow: a single signer
	// receives a stamped copy of the PDF and completes the signature manually.
	ElectronicSignatureFlowSignature ElectronicSignatureFlow = "SIGNATURE"

	// ElectronicSignatureFlowClickwrap is the high-volume flow: the signer
	// clicks a checkbox on a shared version PDF. No per-signature PDF copy is
	// produced; the version PDF is referenced directly.
	ElectronicSignatureFlowClickwrap ElectronicSignatureFlow = "CLICKWRAP"
)

var (
	_ fmt.Stringer             = ElectronicSignatureFlow("")
	_ encoding.TextMarshaler   = ElectronicSignatureFlow("")
	_ encoding.TextUnmarshaler = (*ElectronicSignatureFlow)(nil)
)

func ElectronicSignatureFlows() []ElectronicSignatureFlow {
	return []ElectronicSignatureFlow{
		ElectronicSignatureFlowSignature,
		ElectronicSignatureFlowClickwrap,
	}
}

func (v ElectronicSignatureFlow) IsValid() bool {
	switch v {
	case ElectronicSignatureFlowSignature, ElectronicSignatureFlowClickwrap:
		return true
	}

	return false
}

func (v ElectronicSignatureFlow) String() string {
	return string(v)
}

func (v ElectronicSignatureFlow) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ElectronicSignatureFlow) UnmarshalText(text []byte) error {
	val := ElectronicSignatureFlow(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ElectronicSignatureFlow value: %q", string(text))
	}

	*v = val

	return nil
}
