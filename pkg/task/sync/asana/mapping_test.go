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

package asana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

func TestInboundStateUsesCompletedUntilASectionIsMapped(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.TaskStateDone, InboundState("", true, coredata.TaskStateTodo, nil))
	assert.Equal(t, coredata.TaskStateCanceled, InboundState("", true, coredata.TaskStateCanceled, nil))
	assert.Equal(t, coredata.TaskStateTodo, InboundState("", false, coredata.TaskStateDone, nil))
	assert.Equal(t, coredata.TaskStateInProgress, InboundState("", false, coredata.TaskStateInProgress, nil))
}

func TestInboundStatePrefersTheMappedSection(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		States: map[coredata.TaskState]string{
			coredata.TaskStateTodo: "section-todo",
		},
		SectionInbound: map[string]coredata.TaskState{
			"section-doing": coredata.TaskStateInProgress,
		},
	}

	assert.Equal(
		t,
		coredata.TaskStateInProgress,
		InboundState("section-doing", true, coredata.TaskStateTodo, mapping),
	)
	assert.Equal(
		t,
		coredata.TaskStateDone,
		InboundState("section-other", true, coredata.TaskStateTodo, mapping),
	)
}

func TestBuildAsanaMappingUsesTheFirstStateForASharedSection(t *testing.T) {
	t.Parallel()

	sections := []client.Section{{GID: "section-a", Name: "Doing"}}
	states := []AsanaStateInput{
		{State: coredata.TaskStateInProgress, SectionGID: "section-a", Inbound: true},
		{State: coredata.TaskStateTodo, SectionGID: "section-a"},
	}

	saved, err := BuildAsanaMapping(states, "", nil, sections, nil)
	require.NoError(t, err)
	assert.Equal(t, coredata.TaskStateTodo.String(), saved.SectionInbound["section-a"])
}

func TestOutboundPriorityWritesOnlyAMappedOption(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		PriorityFieldGID: "field",
		Priorities: map[coredata.TaskPriority]string{
			coredata.TaskPriorityHigh: "option-high",
		},
	}

	option, ok := OutboundPriorityOption(coredata.TaskPriorityHigh, mapping)
	assert.True(t, ok)
	assert.Equal(t, "option-high", option)

	_, ok = OutboundPriorityOption(coredata.TaskPriorityLow, mapping)
	assert.False(t, ok)

	_, ok = OutboundPriorityOption(coredata.TaskPriorityHigh, nil)
	assert.False(t, ok)
}

func TestInboundStateKeepsAnUnmappedLocalState(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		States: map[coredata.TaskState]string{
			coredata.TaskStateInProgress: "section-doing",
		},
		SectionInbound: map[string]coredata.TaskState{
			"section-doing": coredata.TaskStateInProgress,
		},
	}

	assert.Equal(
		t,
		coredata.TaskStateBacklog,
		InboundState("section-doing", false, coredata.TaskStateBacklog, mapping),
	)
	assert.Equal(
		t,
		coredata.TaskStateDone,
		InboundState("section-doing", true, coredata.TaskStateBacklog, mapping),
	)
	assert.Equal(
		t,
		coredata.TaskStateInProgress,
		InboundState("section-doing", false, coredata.TaskStateInProgress, mapping),
	)
	assert.Equal(
		t,
		coredata.TaskStateInProgress,
		inboundStateFromSection("section-doing", false, coredata.TaskStateTodo, mapping),
	)
}

func TestInboundPriorityKeepsAnUnmappedLocalPriority(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		PriorityFieldGID: "field",
		Priorities: map[coredata.TaskPriority]string{
			coredata.TaskPriorityHigh: "option-high",
			coredata.TaskPriorityLow:  "option-low",
		},
		OptionInbound: map[string]coredata.TaskPriority{
			"option-high": coredata.TaskPriorityHigh,
			"option-low":  coredata.TaskPriorityLow,
		},
	}

	priority, ok := InboundPriority("option-high", coredata.TaskPriorityMedium, mapping)
	assert.False(t, ok)
	assert.Equal(t, coredata.TaskPriorityMedium, priority)

	priority, ok = InboundPriority("option-high", coredata.TaskPriorityLow, mapping)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityHigh, priority)

	priority, ok = inboundPriorityFromOption("option-high", coredata.TaskPriorityMedium, mapping)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityHigh, priority)
}

func TestInboundStateKeepsTheCurrentStateOnTheSameSection(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		States: map[coredata.TaskState]string{
			coredata.TaskStateTodo:       "section-open",
			coredata.TaskStateInProgress: "section-open",
		},
		SectionInbound: map[string]coredata.TaskState{
			"section-open": coredata.TaskStateTodo,
		},
	}

	assert.Equal(
		t,
		coredata.TaskStateInProgress,
		InboundState("section-open", false, coredata.TaskStateInProgress, mapping),
	)
}

func TestInboundStateUsesTheDesignatedStateWhenCompletionDiffers(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		States: map[coredata.TaskState]string{
			coredata.TaskStateTodo:       "section-open",
			coredata.TaskStateInProgress: "section-open",
		},
		SectionInbound: map[string]coredata.TaskState{
			"section-open": coredata.TaskStateTodo,
		},
	}

	assert.Equal(
		t,
		coredata.TaskStateTodo,
		InboundState("section-open", true, coredata.TaskStateInProgress, mapping),
	)
}

func TestInboundStateUsesTheDesignatedStateForADifferentSection(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		States: map[coredata.TaskState]string{
			coredata.TaskStateTodo:       "section-todo",
			coredata.TaskStateInProgress: "section-doing",
		},
		SectionInbound: map[string]coredata.TaskState{
			"section-todo":  coredata.TaskStateTodo,
			"section-doing": coredata.TaskStateInProgress,
		},
	}

	assert.Equal(
		t,
		coredata.TaskStateInProgress,
		InboundState("section-doing", false, coredata.TaskStateTodo, mapping),
	)
}

func TestInboundPriorityKeepsTheCurrentPriorityOnTheSameOption(t *testing.T) {
	t.Parallel()

	mapping := &AsanaProjectMapping{
		PriorityFieldGID: "field",
		Priorities: map[coredata.TaskPriority]string{
			coredata.TaskPriorityHigh:   "option-shared",
			coredata.TaskPriorityLow:    "option-shared",
			coredata.TaskPriorityMedium: "option-medium",
		},
		OptionInbound: map[string]coredata.TaskPriority{
			"option-shared": coredata.TaskPriorityHigh,
			"option-medium": coredata.TaskPriorityMedium,
		},
	}

	priority, ok := InboundPriority("option-shared", coredata.TaskPriorityLow, mapping)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityLow, priority)

	priority, ok = InboundPriority("option-medium", coredata.TaskPriorityLow, mapping)
	require.True(t, ok)
	assert.Equal(t, coredata.TaskPriorityMedium, priority)
}
