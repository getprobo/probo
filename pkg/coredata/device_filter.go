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
	"github.com/jackc/pgx/v5"
)

type (
	DeviceFilter struct {
		query             *string
		states            []DeviceState
		platforms         []DevicePlatform
		agentVersions     []string
		hasPostureWarning *bool
	}
)

func NewDeviceFilter(
	query *string,
	states []DeviceState,
	platforms []DevicePlatform,
	agentVersions []string,
	hasPostureWarning *bool,
) *DeviceFilter {
	return &DeviceFilter{
		query:             query,
		states:            states,
		platforms:         platforms,
		agentVersions:     agentVersions,
		hasPostureWarning: hasPostureWarning,
	}
}

func (f *DeviceFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"query":                      f.query,
		"filter_states":              filterStrings(f.states),
		"filter_platforms":           filterStrings(f.platforms),
		"filter_agent_versions":      filterStrings(f.agentVersions),
		"filter_has_posture_warning": f.hasPostureWarning,
		"posture_status_fail":        DevicePostureStatusFail,
		"posture_status_unknown":     DevicePostureStatusUnknown,
	}
}

func filterStrings[T ~string](values []T) []string {
	if len(values) == 0 {
		return nil
	}

	encoded := make([]string, len(values))
	for i, value := range values {
		encoded[i] = string(value)
	}

	return encoded
}

func (f *DeviceFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @query::text IS NOT NULL AND @query::text != '' THEN
			EXISTS (
				SELECT 1
				FROM iam_membership_profiles p
				WHERE p.id = devices.owner_profile_id
					AND to_tsvector('simple', p.full_name) @@ (
						SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
						FROM unnest(
							regexp_split_to_array(
								trim(regexp_replace(@query, '-', ' ', 'g')),
								'\s+'
							)
						) AS lexeme
						WHERE lexeme <> ''
					)
			)
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_states::device_state[] IS NOT NULL THEN
			state = ANY(@filter_states::device_state[])
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_platforms::device_platform[] IS NOT NULL THEN
			platform = ANY(@filter_platforms::device_platform[])
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_agent_versions::text[] IS NOT NULL THEN
			agent_version = ANY(@filter_agent_versions::text[])
		ELSE TRUE
	END
)
AND
(
	CASE
		WHEN @filter_has_posture_warning::boolean IS TRUE THEN
			EXISTS (
				SELECT 1
				FROM (
					SELECT DISTINCT ON (check_key)
						status
					FROM device_postures
					WHERE
						device_id = devices.id
						AND tenant_id = devices.tenant_id
					ORDER BY check_key, observed_at DESC
				) latest
				WHERE latest.status IN (
					@posture_status_fail::device_posture_status,
					@posture_status_unknown::device_posture_status
				)
			)
		ELSE TRUE
	END
)`
}
