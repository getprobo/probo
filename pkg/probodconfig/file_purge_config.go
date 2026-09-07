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

package probodconfig

const (
	DefaultFilePurgeIntervalSeconds  = 3600
	DefaultFilePurgeRetentionSeconds = 2592000
	DefaultFilePurgeMaxPerTick       = 1000
)

// FilePurgeConfig configures the worker that permanently deletes
// soft-deleted files from object storage and the database.
// All durations are in seconds.
type FilePurgeConfig struct {
	// Interval is how often the worker scans for expired soft-deleted files.
	Interval int `json:"interval"`
	// Retention is how long a soft-deleted file is kept before purge.
	Retention int `json:"retention"`
	// MaxPerTick is the maximum number of files processed in one run.
	MaxPerTick int `json:"max-per-tick"`
}
