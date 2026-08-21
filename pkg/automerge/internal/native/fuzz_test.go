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

package native

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	officialChangeFixture   = "hW9Kg5nDjoUBzgEAEAECAwQFBgcICQoLDA0ODxABAYDiz6oGF29mZmljaWFsIHNjYWxhciBmaXh0dXJlAAoBBAIHEQYTBxU4NAJCCFYQVxtwAgALBAAACwILfgwNAAx/AAACAAt8AAx0AHUDbmlsAm5vA3llcwR1aW50A2ludAVmbG9hdAR0ZXh0BWJ5dGVzBHdoZW4FY291bnQEbGlzdAAECwQKAX8CAgQCAXYAAQITFIUBVjdpGAMAAhYqeQAAAAAAAPg/aGVsbG8A/wf70JX/vDEJYWIPAA=="
	officialDocumentFixture = "hW9Kg3tNcOoAogIBEAECAwQFBgcICQoLDA0ODxABmcOOhfOq6K9fyRtQMpEkw5nRGiPrg0/hSLI3KA5LqKcHAQIDAhMCIwY1GUACVgIMAQQCBxEGEwcVOCECIw80AkIKVhJXG4ABAn8AfwF/D3+A4s+qBn8Xb2ZmaWNpYWwgc2NhbGFyIGZpeHR1cmV/AH8HAAsEAAALAgt+DA0ADH8AAAIAC3wADHQAdQVieXRlcwVjb3VudAVmbG9hdANpbnQEbGlzdANuaWwCbm8EdGV4dAR1aW50BHdoZW4DeWVzAAQPAHQIAnx/BnYBBX0FegkDAQsEBAF/AgYBAgQCAXw3GIUBFAIAewFWE2kCAgACFgD/BwkAAAAAAAD4P3loZWxsbyr70JX/vDFhYg8AAA=="
)

var officialStorageFixtures = map[string]string{
	"64bit_obj_id_change.automerge":              "hW9Kg2J1YNYBPwAQ2gpUVEJDSYSFAKXr4azZTQGAgICAgIABwb7Eg+EwCAFoYW5nZSAxAAUVAzQBQgJWAnACfwFhAX8AfwB/AA==",
	"64bit_obj_id_doc.automerge":                 "hW9Kg1QlwfAAiAEBENoKVFRCQ0mEhQCl6+Gs2U0BYnVg1rgzMb6KmqiBrHdIhUya1snH32TnNnXqIPdqicoHAQIDAhMIIwc1CkACVgIHFQMhAiMINAFCAlYCgAECfwB/AX+AgICAgIABf8G+xIPhMH8IAWhhbmdlIDF/AH8HAAsEAAALAgt+DA0ADH8AAAIAC3wADHQAdQVieXRlcwVjb3VudAVmbG9hdANpbnQEbGlzdANuaWwCbm8EdGV4dAR1aW50BHdoZW4DeWVzAAQPAHQIAnx/BnYBBX0FegkDAQsEBAF/AgYBAgQCAXw3GIUBFAIAewFWE2kCAgACFgD/BwkAAAAAAAD4P3loZWxsbyr70JX/vDFhYg8AAA==",
	"counter_value_has_incorrect_meta.automerge": "hW9Kgz5jZeYBNQAQiwZtoyQvRmChZG+IqRHxlAEBtLbS0OIwAAAGFQM0AUICVgJXAnACfwFhAX8BfygQf38A",
	"counter_value_is_ok.automerge":              "hW9Kg9Rz2qYBNQAQ/LFH/soQTf6flKoCf2h7awEBvvfR0OIwAAAGFQM0AUICVgJXAnACfwFhAX8BfyjQD38A",
	"counter_value_is_overlong.automerge":        "hW9Kg2/N3H0BNQAQiwZtoyQvRmChZG+IqRHxlAEBtLbS0OIwAAAGFQM0AUICVgJXAnACfwFhAX8BfyjQf38A",
	"two_change_chunks.automerge":                "hW9Kg5rD1zABOQAQ2gpUVEJDSYSFAKXr4azZTQEBwb7Eg+EwCGNoYW5nZSAxAAUVAzQBQgJWAnACfwFhAX8AfwB/AIVvSoOn5yfVAWQBmsPXMPJi2jXnHbWRuegVDemwKCba91AG8imJFq3sbgsQ2gpUVEJDSYSFAKXr4azZTQICwb7Eg+EwCGNoYW5nZSAyAAgBAgICFQM0AUICVgJXAXACfwB/AX8BYQF/AX8WYn8A",
	"two_change_chunks_compressed.automerge":     "hW9Kg5rD1zACPmIQuMUVEuLk7NnSyrD09cM1N30ZGQ/uO9L80IAjOSMxLz1VwZCBVZTZhNGJKYypgKmeMZGxngEEAQEAAP//hW9Kg6fnJ9UCbgBkAJv/AZrD1zDyYto15x21kbnoFQ3psCgm2vdQBvIpiRat7G4LENoKVFRCQ0mEhQCl6+Gs2U0CAsG+xIPhMAhjaGFuZ2UgMgAIAQICAhUDNAFCAlYCVwFwAn8AfwF/AWEBfwF/FmJ/AAEAAP//",
	"two_change_chunks_out_of_order.automerge":   "hW9Kg6fnJ9UBZAGaw9cw8mLaNecdtZG56BUN6bAoJtr3UAbyKYkWrexuCxDaClRUQkNJhIUApevhrNlNAgLBvsSD4TAIY2hhbmdlIDIACAECAgIVAzQBQgJWAlcBcAJ/AH8BfwFhAX8BfxZifwCFb0qDmsPXMAE5ABDaClRUQkNJhIUApevhrNlNAQHBvsSD4TAIY2hhbmdlIDEABRUDNAFCAlYCcAJ/AWEBfwB/AH8A",
	"fuzz-action-is-48":                          "hW9Kg818x5kBMAAQMDAwMDAwMDAwMDAwMDAwMDAwMAAABhUDNAFCAlYCYQJwAjABMDABMH8G0A9/AA==",
	"fuzz-empty-crash":                           "hW9Kg5ailtIAAA==",
	"fuzz-incorrect-max-op":                      "hW9Kg/IrF9QAdAEQAlGmcMDRT1KAagbMI3V0owG3hG4vm1xsqt7I1lr4Yc0pMEkeiGUjKJAdUqx8qMyyJAYBAgMCEwIjAkACVgIIFQYhAiMCNAFCAlYCVwSAAQJ/AH8BfwB/AH8Afwd/BG8BfwF/AH8AAX8Bf0ZvAHBzfwAA",
	"fuzz-invalid-deflate":                       "hW9KgzAwMDAAcQEQMDAwMDAwMDAwMDAwMDAwMAEwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMAYBAgMCIAIwAjECMQIIIAQhAjACMQExAjkCVwOAAQJ/AH8BfwF/AH8Afwd/AmZ6fwB/AQF/AX8277+9fwAA",
	"fuzz-missing-actor":                         "hW9Kgwdz5dgAdAEQ77C8VImLQtKPhfeZlnIU6AGYGdAqe3cyAAAAAAAAACH9xtoZ+f//AAuWa10o81nHmwYBAgMCEwIjAkACVgIIFQYhAiQCNAFCAlYCVwSAAQJ/BH8BfwF/AHcAfwd/BG8AcHN/AH8BAX8DOEZvb3DbfwAA",
	"fuzz-overflow-length":                       "hW9Kgw1aCmMAqwEBAAAQAAAAAAAAAAEAAADj4+Pj4+PjhW9K4+PjhW9Kg+Pj4+Ph4+Nw1nBwcHBwg+MdGOPjL+HjSoPj4+Pj4ePWcHBwcHCD4x0Y4+Mv4+Pj//////8n////////////////////////AAAAAAAAAAAAAAD/////AAAAAAgAAAAAAAAAAAABAAAAAAQAAgEHXf/////////j4+PjBHBwcHBwAQABAAACAgddAQA=",
	"fuzz-too-many-deps":                         "hW9Kg51nWyAAfAEQ77C8VImLQtKPhfeZcHIU6AGYGdAqe3fDi6Us1PDrRyH9xtoZAvksQgOWa10o81nHmwYBAgMCEwIjAkAKVgIIFQYhAiMCNAFCAlYCVwSAAQJ/AH8BfwF/AH/q6urq6urq6gB/B38EbwBwc38AfwEBuwF/Rm9vcHN/AAA=",
	"fuzz-too-many-ops":                          "hW9Kg1XPQM0AfAEQ77C8VImbQtKPhfeZcHIU6AGYGdAqe3fDi6Us1PDrRyH9xtoZAvksQgOWa10o81nHmwYBAgMCEwIjAkACVgIIFQYhAiMCNAFCAlYCVwSAAQp/AH8BfwF/AH8Afwd/BG8AcHN/AH8BAX8Bf0Zvb3Bzf52dnZ2dnZ2dAAA=",
}

func FuzzDecode(f *testing.F) {
	change, err := base64.StdEncoding.DecodeString(officialChangeFixture)
	require.NoError(f, err)
	document, err := base64.StdEncoding.DecodeString(officialDocumentFixture)
	require.NoError(f, err)

	f.Add(change)
	f.Add(document)
	f.Add([]byte{0x85, 0x6f, 0x4a, 0x83})

	for _, encoded := range officialStorageFixtures {
		data, err := base64.StdEncoding.DecodeString(encoded)
		require.NoError(f, err)
		f.Add(data)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024*1024 {
			t.Skip()
		}

		_, _ = Decode(data)
	})
}

func FuzzParseSyncMessage(f *testing.F) {
	message, err := (SyncMessage{
		Version: SyncMessageVersion2,
		Heads:   [][32]byte{{1}},
		Need:    [][32]byte{{2}},
		Have: []SyncHave{
			{
				LastSync: [][32]byte{{3}},
				Bloom:    []byte{4, 5, 6},
			},
		},
		Changes: [][]byte{{7, 8, 9}},
	}).Encode()
	require.NoError(f, err)

	f.Add(message)
	f.Add([]byte{})
	f.Add([]byte{byte(SyncMessageVersion1)})
	f.Add([]byte{byte(SyncMessageVersion2)})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1024*1024 {
			t.Skip()
		}

		parsed, err := ParseSyncMessage(data)
		if err != nil {
			return
		}

		encoded, err := parsed.Encode()
		if err != nil {
			return
		}

		_, _ = ParseSyncMessage(encoded)
	})
}
