// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package provider

import (
	"context"
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func segmentRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderSegment,
		DisplayName:    "Segment",
		SupportsAPIKey: true,
		// Segment authenticates with a Public API token as the default
		// Authorization: Bearer scheme, so no APIKeyHeader. The token is bound
		// to one workspace, but the workspace's region selects the API host
		// (US vs EU) and is not discoverable from the token, so it is captured
		// as an extra setting and resolved to a base URL (Pattern 3 + region);
		// there is nothing to pick.
		ExtraSettings: []ExtraSetting{
			{Key: "region", Label: "Region (US or EU)", Required: true},
		},
		BuildProbeURL: buildSegmentProbeURL,
		// No NewNameResolver: the Public API exposes no read-only workspace-name
		// endpoint on the token's scope, so the source keeps its generic name.
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.SegmentConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read segment connector settings: %w", err)
			}

			if s.BaseURL == "" {
				return nil, fmt.Errorf("cannot create segment driver: base URL is required")
			}

			return drivers.NewSegmentDriver(c, s.BaseURL), nil
		},
	}
}
