// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Access-review service actions.
// Format: access-review:<entity>:<action>
const (
	// Campaign actions
	ActionCampaignGet          = "access-review:campaign:get"
	ActionCampaignList         = "access-review:campaign:list"
	ActionCampaignCreate       = "access-review:campaign:create"
	ActionCampaignUpdate       = "access-review:campaign:update"
	ActionCampaignDelete       = "access-review:campaign:delete"
	ActionCampaignStart        = "access-review:campaign:start"
	ActionCampaignClose        = "access-review:campaign:close"
	ActionCampaignCancel       = "access-review:campaign:cancel"
	ActionCampaignAddSource    = "access-review:campaign:add-source"
	ActionCampaignRemoveSource = "access-review:campaign:remove-source"

	// Entry actions
	ActionEntryGet    = "access-review:entry:get"
	ActionEntryList   = "access-review:entry:list"
	ActionEntryDecide = "access-review:entry:decide"
	ActionEntryFlag   = "access-review:entry:flag"

	// Source actions
	ActionSourceGet    = "access-review:source:get"
	ActionSourceList   = "access-review:source:list"
	ActionSourceCreate = "access-review:source:create"
	ActionSourceUpdate = "access-review:source:update"
	ActionSourceDelete = "access-review:source:delete"
	ActionSourceSync   = "access-review:source:sync"

	// Driver catalog actions (global deployment-scoped catalog).
	ActionDriverCatalogList = "access-review:driver-catalog:list"
)
