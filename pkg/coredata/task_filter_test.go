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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestTaskFilter_SQLArguments_EscapesLikeWildcards(t *testing.T) {
	t.Parallel()

	query := `a%b_c\`
	filter := coredata.NewTaskFilter(&query, nil, nil, nil)
	args := filter.SQLArguments()

	assert.Equal(t, `a\%b\_c\\`, args["filter_query"])
	assert.Contains(t, filter.SQLFragment(), `ESCAPE '\'`)
}

func TestTaskFilter_SQLArguments_LeavesPlainQueryUnchanged(t *testing.T) {
	t.Parallel()

	query := "access"
	filter := coredata.NewTaskFilter(&query, nil, nil, nil)
	args := filter.SQLArguments()

	assert.Equal(t, "access", args["filter_query"])
}

func TestTaskFilter_SQLArguments_SetsAssignedToID(t *testing.T) {
	t.Parallel()

	assignedToID := gid.New(gid.NilTenant, 1)
	filter := coredata.NewTaskFilter(nil, nil, &assignedToID, nil)
	args := filter.SQLArguments()

	assert.Equal(t, assignedToID, args["filter_assigned_to_id"])
	assert.Nil(t, args["filter_query"])
	assert.Nil(t, args["filter_state"])
	assert.Nil(t, args["filter_internal_control_id"])
	assert.Contains(t, filter.SQLFragment(), "assigned_to_profile_id = @filter_assigned_to_id::text")
}

func TestTaskFilter_SQLArguments_SetsInternalControlID(t *testing.T) {
	t.Parallel()

	internalControlID := gid.New(gid.NilTenant, 2)
	filter := coredata.NewTaskFilter(nil, nil, nil, &internalControlID)
	args := filter.SQLArguments()

	assert.Equal(t, internalControlID, args["filter_internal_control_id"])
	assert.Nil(t, args["filter_query"])
	assert.Nil(t, args["filter_state"])
	assert.Nil(t, args["filter_assigned_to_id"])
	assert.Contains(t, filter.SQLFragment(), "filter_ict.internal_control_id = @filter_internal_control_id")
	assert.Contains(t, filter.SQLFragment(), "filter_ict.tenant_id = @tenant_id")
}
