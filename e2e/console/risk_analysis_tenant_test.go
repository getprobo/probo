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

package console_test

import (
	"testing"

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestRiskAnalysis_TenantIsolation(t *testing.T) {
	t.Parallel()

	owner1 := testutil.NewClient(t, testutil.RoleOwner)
	owner2 := testutil.NewClient(t, testutil.RoleOwner)
	raID := factory.CreateRiskAnalysis(owner1)

	var result struct {
		Node *struct {
			ID string `json:"id"`
		} `json:"node"`
	}

	err := owner2.Execute(`
		query($id: ID!) { node(id: $id) { ... on RiskAnalysis { id } } }
	`, map[string]any{"id": raID}, &result)
	testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "RiskAnalysis")
}
