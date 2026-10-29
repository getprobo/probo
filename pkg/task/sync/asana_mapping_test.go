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

package tasksync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/task/sync/asana"
	"go.probo.inc/probo/pkg/validator"
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

func TestBuildAsanaMappingRequiresOneInboundStateForASharedSection(t *testing.T) {
	t.Parallel()

	sections := []asana.Section{{GID: "section-a", Name: "Doing"}}
	states := []AsanaStateInput{
		{State: coredata.TaskStateTodo, SectionGID: "section-a"},
		{State: coredata.TaskStateInProgress, SectionGID: "section-a"},
	}

	_, err := BuildAsanaMapping(states, "", nil, sections, nil)
	require.Error(t, err)

	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)

	states[1].Inbound = true
	saved, err := BuildAsanaMapping(states, "", nil, sections, nil)
	require.NoError(t, err)
	assert.Equal(t, coredata.TaskStateInProgress.String(), saved.SectionInbound["section-a"])
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
