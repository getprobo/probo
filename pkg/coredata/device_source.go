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

// DeviceSource records what created a device row and therefore which
// lifecycle governs it. An AGENT row follows the Probo agent's enrollment and
// heartbeats. An MDM row is created and kept current by a device sync from an
// MDM connector, and never holds an agent credential.
type DeviceSource string

const (
	DeviceSourceAgent DeviceSource = "AGENT"
	DeviceSourceMDM   DeviceSource = "MDM"
)

var (
	_ fmt.Stringer             = DeviceSource("")
	_ encoding.TextMarshaler   = DeviceSource("")
	_ encoding.TextUnmarshaler = (*DeviceSource)(nil)
)

func DeviceSources() []DeviceSource {
	return []DeviceSource{
		DeviceSourceAgent,
		DeviceSourceMDM,
	}
}

func (v DeviceSource) IsValid() bool {
	switch v {
	case
		DeviceSourceAgent,
		DeviceSourceMDM:
		return true
	}

	return false
}

func (v DeviceSource) String() string {
	return string(v)
}

func (v DeviceSource) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DeviceSource) UnmarshalText(text []byte) error {
	val := DeviceSource(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DeviceSource value: %q", string(text))
	}

	*v = val

	return nil
}
