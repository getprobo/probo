// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
