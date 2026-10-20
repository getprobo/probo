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
)

func TestConnectorFilter_SQLArguments_EscapesLikeWildcards(t *testing.T) {
	t.Parallel()

	query := `a%b_c\`
	filter := coredata.NewConnectorListFilter(nil, &query, nil)
	args := filter.SQLArguments()

	assert.Equal(t, `a\%b\_c\\`, args["filter_query"])
	assert.Contains(t, filter.SQLFragment(), `ESCAPE '\'`)
	assert.Nil(t, args["filter_providers"])
}

func TestConnectorFilter_SQLArguments_SetsProvidersAndQueryProviders(t *testing.T) {
	t.Parallel()

	query := "amazon"
	filter := coredata.NewConnectorListFilter(
		[]coredata.ConnectorProvider{coredata.ConnectorProviderBrex},
		&query,
		[]coredata.ConnectorProvider{coredata.ConnectorProviderAWS},
	)
	args := filter.SQLArguments()

	assert.Equal(t, []string{"BREX"}, args["filter_providers"])
	assert.Equal(t, "amazon", args["filter_query"])
	assert.Equal(t, []string{"AWS"}, args["filter_query_providers"])
	assert.Nil(t, args["provider"])
}

func TestConnectorFilter_SQLArguments_LeavesBlankQueryUnset(t *testing.T) {
	t.Parallel()

	query := "   "
	provider := coredata.ConnectorProviderSlack
	filter := coredata.NewConnectorProviderFilter(&provider)
	list := coredata.NewConnectorListFilter(nil, &query, nil)

	assert.Equal(t, &provider, filter.SQLArguments()["provider"])
	assert.Nil(t, list.SQLArguments()["filter_query"])
	assert.Nil(t, list.SQLArguments()["provider"])
}
