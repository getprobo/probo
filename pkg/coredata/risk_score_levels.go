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
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type (
	RiskScoreLevel string

	RiskScoreLevelMapping struct {
		Level    RiskScoreLevel `json:"level"`
		MinScore int            `json:"min_score"`
	}

	ScoreLevels []RiskScoreLevelMapping
)

const (
	RiskScoreLevelLow      RiskScoreLevel = "LOW"
	RiskScoreLevelMedium   RiskScoreLevel = "MEDIUM"
	RiskScoreLevelHigh     RiskScoreLevel = "HIGH"
	RiskScoreLevelCritical RiskScoreLevel = "CRITICAL"
)

func RiskScoreLevels() []RiskScoreLevel {
	return []RiskScoreLevel{
		RiskScoreLevelLow,
		RiskScoreLevelMedium,
		RiskScoreLevelHigh,
		RiskScoreLevelCritical,
	}
}

func (levels ScoreLevels) LevelFor(score int) RiskScoreLevel {
	var chosen RiskScoreLevel

	bestMin := -1
	for _, level := range levels {
		if level.MinScore <= score && level.MinScore >= bestMin {
			bestMin = level.MinScore
			chosen = level.Level
		}
	}

	return chosen
}

func (levels ScoreLevels) Spec() string {
	parts := make([]string, len(levels))
	for i, level := range levels {
		parts[i] = fmt.Sprintf("%s=%d", level.Level, level.MinScore)
	}

	return strings.Join(parts, ",")
}

func ParseScoreLevels(spec string) (ScoreLevels, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("score levels are required")
	}

	parts := strings.Split(spec, ",")
	levels := make(ScoreLevels, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		name, rawScore, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid score level %q", part)
		}

		level, err := parseRiskScoreLevel(name)
		if err != nil {
			return nil, err
		}

		minScore, err := strconv.Atoi(strings.TrimSpace(rawScore))
		if err != nil {
			return nil, fmt.Errorf("invalid score for %s", level)
		}

		levels = append(levels, RiskScoreLevelMapping{
			Level:    level,
			MinScore: minScore,
		})
	}

	return levels, nil
}

func parseRiskScoreLevel(value string) (RiskScoreLevel, error) {
	switch RiskScoreLevel(strings.ToUpper(strings.TrimSpace(value))) {
	case RiskScoreLevelLow, RiskScoreLevelMedium, RiskScoreLevelHigh, RiskScoreLevelCritical:
		return RiskScoreLevel(strings.ToUpper(strings.TrimSpace(value))), nil
	default:
		return "", fmt.Errorf("invalid score level %q", strings.TrimSpace(value))
	}
}

func (levels ScoreLevels) Value() (driver.Value, error) {
	if levels == nil {
		return nil, fmt.Errorf("score levels are required")
	}

	data, err := json.Marshal(levels)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal score levels: %w", err)
	}

	return data, nil
}

func (levels *ScoreLevels) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("cannot scan score levels: value is nil")
	}

	var data []byte

	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan score levels: unsupported type %T", value)
	}

	if len(data) == 0 || string(data) == "null" || string(data) == "{}" || string(data) == "[]" {
		return fmt.Errorf("cannot scan score levels: value is empty")
	}

	var decoded ScoreLevels
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("cannot unmarshal score levels: %w", err)
	}

	*levels = decoded

	return nil
}
