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

package deviceagent

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedEvents struct {
	info []string
	warn []string
	errs []string
	fail bool
}

func (r *recordedEvents) Info(_ uint32, msg string) error {
	if r.fail {
		return errors.New("full")
	}

	r.info = append(r.info, msg)

	return nil
}

func (r *recordedEvents) Warning(_ uint32, msg string) error {
	r.warn = append(r.warn, msg)

	return nil
}

func (r *recordedEvents) Error(_ uint32, msg string) error {
	r.errs = append(r.errs, msg)

	return nil
}

func TestEventLogWriter_ReportsLevels(t *testing.T) {
	t.Parallel()

	events := &recordedEvents{}
	writer := newEventLogWriter(events)

	_, err := writer.Write([]byte("{\"level\":\"INFO\",\"msg\":\"started\"}\n"))
	require.NoError(t, err)
	_, err = writer.Write([]byte("{\"level\":\"WARN\",\"msg\":\"slow\"}\n{\"level\":\"ERROR\",\"msg\":\"down\"}\n"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	assert.Equal(t, []string{"{\"level\":\"INFO\",\"msg\":\"started\"}"}, events.info)
	assert.Equal(t, []string{"{\"level\":\"WARN\",\"msg\":\"slow\"}"}, events.warn)
	assert.Equal(t, []string{"{\"level\":\"ERROR\",\"msg\":\"down\"}"}, events.errs)
}

func TestEventLogWriter_FlushesAPartialLineOnClose(t *testing.T) {
	t.Parallel()

	events := &recordedEvents{}
	writer := newEventLogWriter(events)

	_, err := writer.Write([]byte("{\"level\":\"INFO\",\"msg\":\"partial\"}"))
	require.NoError(t, err)
	assert.Empty(t, events.info)

	require.NoError(t, writer.Close())
	assert.Equal(t, []string{"{\"level\":\"INFO\",\"msg\":\"partial\"}"}, events.info)
}

func TestEventLogWriter_ReturnsTheReportError(t *testing.T) {
	t.Parallel()

	writer := newEventLogWriter(&recordedEvents{fail: true})

	_, err := writer.Write([]byte("{\"level\":\"INFO\",\"msg\":\"started\"}\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot write Windows event log")
}
