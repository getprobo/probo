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

package timespan

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sosodev/duration"
)

const (
	microsecondsPerSecond = 1_000_000
	microsecondsPerMinute = 60 * microsecondsPerSecond
	microsecondsPerHour   = 60 * microsecondsPerMinute
	microsecondsPerDay    = 24 * microsecondsPerHour
	daysPerMonth          = 30
	microsecondsPerMonth  = daysPerMonth * microsecondsPerDay
	monthsPerYear         = 12
	daysPerWeek           = 7
)

type TimeSpan struct {
	Months       int32
	Days         int32
	Microseconds int64
}

var (
	Zero TimeSpan

	_ pgtype.IntervalValuer  = TimeSpan{}
	_ pgtype.IntervalScanner = (*TimeSpan)(nil)
)

func FromDuration(d time.Duration) TimeSpan {
	return TimeSpan{Microseconds: d.Microseconds()}
}

func (ts TimeSpan) IsZero() bool {
	return ts == Zero
}

func Parse(s string) (TimeSpan, error) {
	parsed, err := parseISO(s)
	if err != nil {
		return Zero, fmt.Errorf("cannot parse timespan %q: %w", s, err)
	}

	ts, err := fromISO(parsed)
	if err != nil {
		return Zero, fmt.Errorf("cannot convert timespan %q: %w", s, err)
	}

	return ts, nil
}

func parseISO(s string) (*duration.Duration, error) {
	const (
		parsingPeriod = iota
		parsingTime
	)

	state := parsingPeriod
	parsed := &duration.Duration{}
	num := ""
	seenComponent := false
	seenTime := false

	switch {
	case strings.HasPrefix(s, "P"):
	case strings.HasPrefix(s, "-P"):
		parsed.Negative = true
		s = strings.TrimPrefix(s, "-")
	default:
		return nil, fmt.Errorf("cannot parse timespan: unexpected input")
	}

	var err error

	for _, char := range s {
		switch char {
		case 'P':
			if state != parsingPeriod {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}
		case 'T':
			if state == parsingTime {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			state = parsingTime
		case 'Y':
			if state != parsingPeriod {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			parsed.Years, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse years: %w", err)
			}

			num = ""
			seenComponent = true
		case 'M':
			switch state {
			case parsingPeriod:
				parsed.Months, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot parse months: %w", err)
				}
			case parsingTime:
				parsed.Minutes, err = strconv.ParseFloat(num, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot parse minutes: %w", err)
				}

				seenTime = true
			}

			num = ""
			seenComponent = true
		case 'W':
			if state != parsingPeriod {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			parsed.Weeks, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse weeks: %w", err)
			}

			num = ""
			seenComponent = true
		case 'D':
			if state != parsingPeriod {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			parsed.Days, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse days: %w", err)
			}

			num = ""
			seenComponent = true
		case 'H':
			if state != parsingTime {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			parsed.Hours, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse hours: %w", err)
			}

			num = ""
			seenComponent = true
			seenTime = true
		case 'S':
			if state != parsingTime {
				return nil, fmt.Errorf("cannot parse timespan: unexpected input")
			}

			parsed.Seconds, err = strconv.ParseFloat(num, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse seconds: %w", err)
			}

			num = ""
			seenComponent = true
			seenTime = true
		default:
			if unicode.IsNumber(char) || char == '.' || (char == '-' && num == "") {
				num += string(char)
				continue
			}

			return nil, fmt.Errorf("cannot parse timespan: unexpected input")
		}
	}

	if num != "" || !seenComponent || (state == parsingTime && !seenTime) {
		return nil, fmt.Errorf("cannot parse timespan: incomplete expression")
	}

	return parsed, nil
}

func fromISO(d *duration.Duration) (TimeSpan, error) {
	months, fracMonths := math.Modf(d.Years*monthsPerYear + d.Months)
	days, fracDays := math.Modf(d.Weeks*daysPerWeek + d.Days + fracMonths*daysPerMonth)
	microseconds := math.Round(
		d.Hours*float64(microsecondsPerHour) +
			d.Minutes*float64(microsecondsPerMinute) +
			d.Seconds*float64(microsecondsPerSecond) +
			fracDays*float64(microsecondsPerDay),
	)

	if d.Negative {
		months = -months
		days = -days
		microseconds = -microseconds
	}

	if months < math.MinInt32 || months > math.MaxInt32 {
		return Zero, fmt.Errorf("cannot parse timespan: months overflow")
	}

	if days < math.MinInt32 || days > math.MaxInt32 {
		return Zero, fmt.Errorf("cannot parse timespan: days overflow")
	}

	if math.IsInf(microseconds, 0) || math.IsNaN(microseconds) ||
		microseconds < math.MinInt64 || microseconds >= 1<<63 {
		return Zero, fmt.Errorf("cannot parse timespan: microseconds overflow")
	}

	return TimeSpan{
		Months:       int32(months),
		Days:         int32(days),
		Microseconds: int64(microseconds),
	}.normalizeClock(), nil
}

func (ts TimeSpan) String() string {
	ts = ts.normalizeClock()
	if ts.IsZero() {
		return "PT0S"
	}

	prefix := "P"
	months := int64(ts.Months)
	days := int64(ts.Days)
	microseconds := absInt64(ts.Microseconds)
	clockNegative := false

	if ts.Months <= 0 && ts.Days <= 0 && ts.Microseconds <= 0 {
		prefix = "-P"
		months = int64(absInt32(ts.Months))
		days = int64(absInt32(ts.Days))
	} else if ts.Microseconds < 0 {
		clockNegative = true
	}

	var b strings.Builder
	b.WriteString(prefix)

	years := months / monthsPerYear
	months = months % monthsPerYear

	if years != 0 {
		fmt.Fprintf(&b, "%dY", years)
	}

	if months != 0 {
		fmt.Fprintf(&b, "%dM", months)
	}

	if days != 0 {
		fmt.Fprintf(&b, "%dD", days)
	}

	if microseconds != 0 {
		b.WriteByte('T')

		hours := microseconds / microsecondsPerHour
		microseconds %= microsecondsPerHour
		minutes := microseconds / microsecondsPerMinute
		microseconds %= microsecondsPerMinute
		seconds := microseconds / microsecondsPerSecond
		frac := microseconds % microsecondsPerSecond

		if hours != 0 {
			if clockNegative {
				b.WriteByte('-')
			}

			fmt.Fprintf(&b, "%dH", hours)
		}

		if minutes != 0 {
			if clockNegative {
				b.WriteByte('-')
			}

			fmt.Fprintf(&b, "%dM", minutes)
		}

		if seconds != 0 || frac != 0 {
			if clockNegative {
				b.WriteByte('-')
			}

			if frac == 0 {
				fmt.Fprintf(&b, "%dS", seconds)
			} else {
				fmt.Fprintf(&b, "%d.%sS", seconds, strings.TrimRight(fmt.Sprintf("%06d", frac), "0"))
			}
		}
	}

	if b.String() == "P" || b.String() == "-P" {
		return "PT0S"
	}

	return b.String()
}

func absInt32(v int32) uint64 {
	if v >= 0 {
		return uint64(v)
	}

	if v == math.MinInt32 {
		return 1 << 31
	}

	return uint64(-v)
}

func absInt64(v int64) uint64 {
	if v >= 0 {
		return uint64(v)
	}

	if v == math.MinInt64 {
		return 1 << 63
	}

	return uint64(-v)
}

func (ts TimeSpan) ClockDuration() (time.Duration, bool) {
	if ts.Months != 0 {
		return 0, false
	}

	dayNs, ok := mulInt64(int64(ts.Days), int64(24*time.Hour))
	if !ok {
		return 0, false
	}

	usNs, ok := mulInt64(ts.Microseconds, int64(time.Microsecond))
	if !ok {
		return 0, false
	}

	total, ok := addInt64(dayNs, usNs)
	if !ok {
		return 0, false
	}

	return time.Duration(total), true
}

func mulInt64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}

	if a == math.MinInt64 || b == math.MinInt64 {
		if a == 1 {
			return b, true
		}

		if b == 1 {
			return a, true
		}

		return 0, false
	}

	result := a * b
	if result/a != b {
		return 0, false
	}

	return result, true
}

func addInt64(a, b int64) (int64, bool) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, false
	}

	if b < 0 && a < math.MinInt64-b {
		return 0, false
	}

	return a + b, true
}

// Compare reports whether ts is less than, equal to, or greater than other.
// Ordering treats one month as 30 days and one day as 24 hours. That is only
// an ordering rule; a calendar month is still not a fixed elapsed duration.
func (ts TimeSpan) Compare(other TimeSpan) int {
	return ts.linearMicroseconds().Cmp(other.linearMicroseconds())
}

// CompareDuration reports whether ts is less than, equal to, or greater than d.
// The span is ordered at microsecond precision against the full nanosecond
// bound so leftover nanoseconds in d are not discarded.
func (ts TimeSpan) CompareDuration(d time.Duration) int {
	return ts.linearNanoseconds().Cmp(big.NewInt(d.Nanoseconds()))
}

func (ts TimeSpan) linearNanoseconds() *big.Int {
	ns := ts.linearMicroseconds()
	ns.Mul(ns, big.NewInt(int64(time.Microsecond)))

	return ns
}

func (ts TimeSpan) linearMicroseconds() *big.Int {
	months := new(big.Int).SetInt64(int64(ts.Months))
	months.Mul(months, big.NewInt(microsecondsPerMonth))

	days := new(big.Int).SetInt64(int64(ts.Days))
	days.Mul(days, big.NewInt(microsecondsPerDay))

	months.Add(months, days)
	months.Add(months, new(big.Int).SetInt64(ts.Microseconds))

	return months
}

func (ts TimeSpan) IntervalValue() (pgtype.Interval, error) {
	ts = ts.normalizeClock()

	return pgtype.Interval{
		Months:       ts.Months,
		Days:         ts.Days,
		Microseconds: ts.Microseconds,
		Valid:        true,
	}, nil
}

func (ts *TimeSpan) ScanInterval(v pgtype.Interval) error {
	if !v.Valid {
		return fmt.Errorf("cannot scan NULL into TimeSpan")
	}

	*ts = TimeSpan{
		Months:       v.Months,
		Days:         v.Days,
		Microseconds: v.Microseconds,
	}.normalizeClock()

	return nil
}

// normalizeClock borrows between days and microseconds so the clock fields
// share a sign and |microseconds| < 24 hours. Months are left untouched: a
// calendar month is not a fixed number of days.
func (ts TimeSpan) normalizeClock() TimeSpan {
	total := new(big.Int).SetInt64(int64(ts.Days))
	total.Mul(total, big.NewInt(microsecondsPerDay))
	total.Add(total, new(big.Int).SetInt64(ts.Microseconds))

	dayUs := big.NewInt(microsecondsPerDay)
	days, us := new(big.Int).QuoRem(total, dayUs, new(big.Int))

	if !days.IsInt64() || !us.IsInt64() {
		return ts
	}

	normalizedDays := days.Int64()
	if normalizedDays < math.MinInt32 || normalizedDays > math.MaxInt32 {
		return ts
	}

	ts.Days = int32(normalizedDays)
	ts.Microseconds = us.Int64()

	return ts
}

func (ts TimeSpan) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(ts.String())
	if err != nil {
		return nil, fmt.Errorf("cannot marshal timespan: %w", err)
	}

	return data, nil
}

func (ts *TimeSpan) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*ts = Zero
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("cannot unmarshal timespan: %w", err)
	}

	parsed, err := Parse(s)
	if err != nil {
		return fmt.Errorf("cannot parse timespan: %w", err)
	}

	*ts = parsed

	return nil
}
