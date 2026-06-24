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

package accessreview

import "go.probo.inc/probo/pkg/coredata"

const (
	ScopeV1AccessReviewRead coredata.OAuth2Scope = "v1:access-review:read"
	ScopeV1AccessReview     coredata.OAuth2Scope = "v1:access-review"
)

// OAuth2ScopeMappings maps OAuth2 scopes to access-review actions.
var OAuth2ScopeMappings = map[coredata.OAuth2Scope][]string{
	ScopeV1AccessReviewRead: {
		ActionCampaignGet,
		ActionCampaignList,
		ActionEntryGet,
		ActionEntryList,
		ActionSourceGet,
		ActionSourceList,
		ActionDriverCatalogList,
	},
	ScopeV1AccessReview: {
		ActionCampaignGet,
		ActionCampaignList,
		ActionEntryGet,
		ActionEntryList,
		ActionSourceGet,
		ActionSourceList,
		ActionDriverCatalogList,
		ActionCampaignCreate,
		ActionCampaignUpdate,
		ActionCampaignDelete,
		ActionCampaignStart,
		ActionCampaignClose,
		ActionCampaignCancel,
		ActionCampaignAddSource,
		ActionCampaignRemoveSource,
		ActionEntryDecide,
		ActionEntryFlag,
		ActionSourceCreate,
		ActionSourceUpdate,
		ActionSourceDelete,
		ActionSourceSync,
	},
}
