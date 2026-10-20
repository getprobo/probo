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
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/timespan"
)

type InternalControlFieldFlags struct {
	Code                 string
	ControlType          string
	Nature               string
	OperatingMode        string
	OperatingFrequency   string
	OperatingEvent       string
	EvidenceCadence      string
	TestingCadence       string
	ImplementationStatus string
	OwnerID              string
	ReviewerID           string
}

func AddInternalControlFieldFlags(
	cmd *cobra.Command,
	code *string,
	controlType *string,
	nature *string,
	operatingMode *string,
	operatingFrequency *string,
	operatingEvent *string,
	evidenceCadence *string,
	testingCadence *string,
	implementationStatus *string,
	ownerID *string,
	reviewerID *string,
) {
	cmd.Flags().StringVar(code, "code", "", "Stable reference, for example IC-ACCESS-01")
	cmd.Flags().StringVar(controlType, "control-type", "", "Control type: PREVENTIVE, DETECTIVE, CORRECTIVE")
	cmd.Flags().StringVar(nature, "nature", "", "Nature: MANUAL")
	cmd.Flags().StringVar(operatingMode, "operating-mode", "", "Operating mode: CONTINUOUS, EVENT, or PERIODIC. Empty clears it")
	cmd.Flags().StringVar(operatingFrequency, "operating-frequency", "", "ISO-8601 duration for PERIODIC mode, for example P3M")
	cmd.Flags().StringVar(operatingEvent, "operating-event", "", "Event that runs the control, for example when someone leaves")
	cmd.Flags().StringVar(evidenceCadence, "evidence-cadence", "", "How often evidence is collected, as an ISO-8601 duration such as P1M")
	cmd.Flags().StringVar(testingCadence, "testing-cadence", "", "How often effectiveness is tested, as an ISO-8601 duration such as P3M")
	cmd.Flags().StringVar(implementationStatus, "implementation-status", "", "Status: NOT_IMPLEMENTED, IN_PROGRESS, IMPLEMENTED, OPERATING")
	cmd.Flags().StringVar(ownerID, "owner-id", "", "Owner profile ID")
	cmd.Flags().StringVar(reviewerID, "reviewer-id", "", "Reviewer profile ID")
}

func SetOperatingFrequency(cmd *cobra.Command, input map[string]any, mode, interval, event string) error {
	if !cmd.Flags().Changed("operating-mode") &&
		!cmd.Flags().Changed("operating-frequency") &&
		!cmd.Flags().Changed("operating-event") {
		return nil
	}

	if mode == "" {
		if interval != "" || event != "" {
			return fmt.Errorf("operating-mode is required")
		}

		input["operatingFrequency"] = nil

		return nil
	}

	if err := cmdutil.ValidateEnum("operating-mode", mode, []string{"CONTINUOUS", "EVENT", "PERIODIC"}); err != nil {
		return err
	}

	if err := rejectMismatchedOperatingFlags(mode, interval, event); err != nil {
		return err
	}

	value := map[string]any{"mode": mode}

	switch mode {
	case "PERIODIC":
		span, err := timespan.Parse(interval)
		if err != nil {
			return fmt.Errorf("operating-frequency: %w", err)
		}

		value["interval"] = span.String()
	case "EVENT":
		// An omitted event flag keeps the description the caller already
		// resolved. An explicit empty value clears it.
		if cmd.Flags().Changed("operating-event") {
			if event == "" {
				value["event"] = nil
			} else {
				value["event"] = event
			}
		} else if event != "" {
			value["event"] = event
		}
	}

	input["operatingFrequency"] = value

	return nil
}

func rejectMismatchedOperatingFlags(mode, interval, event string) error {
	switch mode {
	case "CONTINUOUS":
		if interval != "" {
			return fmt.Errorf("--operating-frequency is only valid with --operating-mode PERIODIC")
		}

		if event != "" {
			return fmt.Errorf("--operating-event is only valid with --operating-mode EVENT")
		}
	case "EVENT":
		if interval != "" {
			return fmt.Errorf("--operating-frequency is only valid with --operating-mode PERIODIC")
		}
	case "PERIODIC":
		if event != "" {
			return fmt.Errorf("--operating-event is only valid with --operating-mode EVENT")
		}

		if interval == "" {
			return fmt.Errorf("operating-frequency is required for PERIODIC")
		}
	}

	return nil
}

func setOptionalID(cmd *cobra.Command, input map[string]any, flag, key, raw string) {
	if !cmd.Flags().Changed(flag) {
		return
	}

	if raw == "" {
		input[key] = nil
		return
	}

	input[key] = raw
}

func setDuration(cmd *cobra.Command, input map[string]any, flag, key, raw string) error {
	if !cmd.Flags().Changed(flag) {
		return nil
	}

	if raw == "" {
		input[key] = nil
		return nil
	}

	span, err := timespan.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", flag, err)
	}

	input[key] = span.String()

	return nil
}

func SetInternalControlFields(cmd *cobra.Command, input map[string]any, flags InternalControlFieldFlags) error {
	if cmd.Flags().Changed("code") {
		input["code"] = flags.Code
	}

	if cmd.Flags().Changed("control-type") {
		if err := cmdutil.ValidateEnum("control-type", flags.ControlType, []string{"PREVENTIVE", "DETECTIVE", "CORRECTIVE"}); err != nil {
			return err
		}

		input["controlType"] = flags.ControlType
	}

	if cmd.Flags().Changed("nature") {
		if err := cmdutil.ValidateEnum("nature", flags.Nature, []string{"MANUAL"}); err != nil {
			return err
		}

		input["nature"] = flags.Nature
	}

	if err := SetOperatingFrequency(cmd, input, flags.OperatingMode, flags.OperatingFrequency, flags.OperatingEvent); err != nil {
		return err
	}

	if err := setDuration(cmd, input, "evidence-cadence", "evidenceCadence", flags.EvidenceCadence); err != nil {
		return err
	}

	if err := setDuration(cmd, input, "testing-cadence", "testingCadence", flags.TestingCadence); err != nil {
		return err
	}

	if cmd.Flags().Changed("implementation-status") {
		if err := cmdutil.ValidateEnum(
			"implementation-status",
			flags.ImplementationStatus,
			[]string{"NOT_IMPLEMENTED", "IN_PROGRESS", "IMPLEMENTED", "OPERATING"},
		); err != nil {
			return err
		}

		input["implementationStatus"] = flags.ImplementationStatus
	}

	setOptionalID(cmd, input, "owner-id", "ownerId", flags.OwnerID)
	setOptionalID(cmd, input, "reviewer-id", "reviewerId", flags.ReviewerID)

	return nil
}
