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

package fields

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetOperatingFrequency_RejectsFlagsFromAnotherMode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		mode     string
		interval string
		event    string
	}{
		{name: "event with an interval", mode: "EVENT", interval: "P3M"},
		{name: "continuous with an interval", mode: "CONTINUOUS", interval: "P3M"},
		{name: "continuous with an event", mode: "CONTINUOUS", event: "when someone leaves"},
		{name: "periodic with an event", mode: "PERIODIC", interval: "P3M", event: "when someone leaves"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cmd := operatingCommand(t, tc.mode, tc.interval, tc.event)
			err := SetOperatingFrequency(cmd, map[string]any{}, tc.mode, tc.interval, tc.event)
			require.Error(t, err)
		})
	}
}

func TestSetOperatingFrequency_PeriodicKeepsTheInterval(t *testing.T) {
	t.Parallel()

	cmd := operatingCommand(t, "PERIODIC", "P3M", "")
	input := map[string]any{}
	require.NoError(t, SetOperatingFrequency(cmd, input, "PERIODIC", "P3M", ""))
	assert.Equal(t, map[string]any{"mode": "PERIODIC", "interval": "P3M"}, input["operatingFrequency"])
}

func operatingCommand(t *testing.T, mode, interval, event string) *cobra.Command {
	t.Helper()

	cmd := &cobra.Command{Use: "measure"}
	var code, controlType, nature, operatingMode, operatingFrequency, operatingEvent string
	var evidence, testing, status, owner, reviewer string
	AddMeasureFieldFlags(
		cmd,
		&code,
		&controlType,
		&nature,
		&operatingMode,
		&operatingFrequency,
		&operatingEvent,
		&evidence,
		&testing,
		&status,
		&owner,
		&reviewer,
	)
	require.NoError(t, cmd.Flags().Set("operating-mode", mode))
	if interval != "" {
		require.NoError(t, cmd.Flags().Set("operating-frequency", interval))
	}
	if event != "" {
		require.NoError(t, cmd.Flags().Set("operating-event", event))
	}

	return cmd
}
