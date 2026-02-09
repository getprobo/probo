// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package types

import "go.probo.inc/probo/pkg/coredata"

func NewSCIMBridge(bridge *coredata.SCIMBridge) *SCIMBridge {
	var connector *Connector
	if bridge.ConnectorID != nil {
		connector = &Connector{
			ID: *bridge.ConnectorID,
		}
	}

	return &SCIMBridge{
		ID:    bridge.ID,
		State: bridge.State,
		ScimConfiguration: &SCIMConfiguration{
			ID: bridge.ScimConfigurationID,
		},
		Connector:     connector,
		Type:          bridge.Type,
		ExcludedUserNames: bridge.ExcludedUserNames,
		CreatedAt:     bridge.CreatedAt,
		UpdatedAt:     bridge.UpdatedAt,
	}
}
