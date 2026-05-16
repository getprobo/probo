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

package lib0

// WriteVarUint writes x using the lib0 unsigned variable-length integer
// encoding: little-endian 7-bit groups with the continuation bit (0x80) set
// on every byte except the last.
func WriteVarUint(e *Encoder, x uint64) {
	for x >= 0x80 {
		_ = e.WriteByte(byte(x) | 0x80)
		x >>= 7
	}
	_ = e.WriteByte(byte(x))
}

// ReadVarUint reads a lib0-encoded unsigned variable-length integer.
//
// Returns ErrUnexpectedEOF if the input is exhausted before the
// continuation bit is cleared, and ErrVarUintOverflow if the encoded value
// would not fit in a uint64.
func ReadVarUint(d *Decoder) (uint64, error) {
	var (
		v     uint64
		shift uint
	)

	for {
		b, err := d.ReadByte()
		if err != nil {
			return 0, err
		}

		if shift >= 64 || (shift == 63 && b > 1) {
			return 0, ErrVarUintOverflow
		}

		v |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return v, nil
		}

		shift += 7
	}
}

// WriteVarInt writes x using the lib0 signed variable-length integer
// encoding. The first byte carries the sign in bit 6 and a continuation
// flag in bit 7; bits 0..5 hold the lowest six bits of |x|. Subsequent
// bytes follow the unsigned varint format with 7-bit groups.
func WriteVarInt(e *Encoder, x int64) {
	var (
		isNegative = x < 0
		u          uint64
	)

	if isNegative {
		// uint64(-x) for x = math.MinInt64 produces 1<<63 in Go, which
		// is the correct unsigned magnitude.
		u = uint64(-x)
	} else {
		u = uint64(x)
	}

	first := byte(u & 0x3f)
	if isNegative {
		first |= 0x40
	}
	if u > 0x3f {
		first |= 0x80
	}
	_ = e.WriteByte(first)

	u >>= 6
	for u > 0 {
		b := byte(u & 0x7f)
		if u > 0x7f {
			b |= 0x80
		}
		_ = e.WriteByte(b)
		u >>= 7
	}
}

// ReadVarInt reads a lib0-encoded signed variable-length integer.
func ReadVarInt(d *Decoder) (int64, error) {
	first, err := d.ReadByte()
	if err != nil {
		return 0, err
	}

	var (
		num        = uint64(first & 0x3f)
		isNegative = first&0x40 != 0
		hasMore    = first&0x80 != 0
		mult       = uint64(64)
	)

	for hasMore {
		b, err := d.ReadByte()
		if err != nil {
			return 0, err
		}

		// (b&0x7f) * mult overflows uint64 once the encoded value
		// exceeds 64 bits; the divide round-trip detects that without
		// risking a panic (mult is always non-zero — see the shift
		// guard below).
		add := uint64(b&0x7f) * mult
		if add/mult != uint64(b&0x7f) {
			return 0, ErrVarIntOverflow
		}
		// num + add cannot wrap: by construction num occupies the low
		// 6 + 7(k-1) bits before iteration k, and add only contributes
		// bits at positions ≥ 6 + 7(k-1), so the two never collide and
		// their sum stays within uint64 (which the previous check also
		// guarantees).
		num += add

		hasMore = b&0x80 != 0
		if hasMore {
			// Compute the next multiplier (mult * 128) with an
			// overflow guard so that mult stays strictly positive.
			next := mult << 7
			if next>>7 != mult {
				return 0, ErrVarIntOverflow
			}
			mult = next
		}
	}

	if isNegative {
		// |num| can equal 1<<63, which represents math.MinInt64.
		if num > 1<<63 {
			return 0, ErrVarIntOverflow
		}
		if num == 1<<63 {
			return -1 << 63, nil
		}
		return -int64(num), nil
	}

	if num > 1<<63-1 {
		return 0, ErrVarIntOverflow
	}
	return int64(num), nil
}
