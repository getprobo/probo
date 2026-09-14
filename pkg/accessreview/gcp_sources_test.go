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

package accessreview_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCreateGCPSources_Empty(t *testing.T) {
	t.Parallel()

	_, _, err := (*accessreview.Service)(nil).CreateGCPSources(
		context.Background(),
		nil,
		gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		nil,
	)
	require.ErrorIs(t, err, accessreview.ErrGCPSourcesEmpty)
}

func TestCreateGCPSources_TooMany(t *testing.T) {
	t.Parallel()

	projects := make([]accessreview.GCPSourceProject, accessreview.MaxGCPSourceProjects+1)
	_, _, err := (*accessreview.Service)(nil).CreateGCPSources(
		context.Background(),
		nil,
		gid.New(gid.NewTenantID(), coredata.OrganizationEntityType),
		projects,
	)
	require.ErrorIs(t, err, accessreview.ErrGCPSourcesTooMany)
	assert.Contains(t, err.Error(), "100")
}
