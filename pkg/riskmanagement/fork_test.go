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

package riskmanagement

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestOrderBoundaries_ParentFirst(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	rootID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	childID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	grandchildID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)

	grandchild := &coredata.RiskAnalysisBoundary{ID: grandchildID, ParentBoundaryID: &childID}
	child := &coredata.RiskAnalysisBoundary{ID: childID, ParentBoundaryID: &rootID}
	root := &coredata.RiskAnalysisBoundary{ID: rootID}

	ordered, err := orderBoundaries([]*coredata.RiskAnalysisBoundary{grandchild, child, root})
	require.NoError(t, err)
	require.Len(t, ordered, 3)
	assert.Equal(t, rootID, ordered[0].ID)
	assert.Equal(t, childID, ordered[1].ID)
	assert.Equal(t, grandchildID, ordered[2].ID)
}

func TestOrderBoundaries_Cycle(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	aID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)
	bID := gid.New(tenantID, coredata.RiskAnalysisBoundaryEntityType)

	a := &coredata.RiskAnalysisBoundary{ID: aID, ParentBoundaryID: &bID}
	b := &coredata.RiskAnalysisBoundary{ID: bID, ParentBoundaryID: &aID}

	_, err := orderBoundaries([]*coredata.RiskAnalysisBoundary{a, b})
	require.Error(t, err)
}
