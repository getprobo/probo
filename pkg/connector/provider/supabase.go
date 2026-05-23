// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func supabaseRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderSupabase,
		DisplayName:    "Supabase",
		ProbeURL:       "https://api.supabase.com/v1/organizations",
		SupportsAPIKey: true,
		ExtraSettings: []ExtraSetting{
			{Key: "organizationSlug", Label: "Organization Slug", Required: true},
		},
		MarshalSettings: func(in *SettingsInput) (json.RawMessage, error) {
			if in == nil || in.SupabaseOrganizationSlug == nil || *in.SupabaseOrganizationSlug == "" {
				return nil, fmt.Errorf("cannot create supabase connector: supabaseOrganizationSlug is required")
			}

			return json.Marshal(&coredata.SupabaseConnectorSettings{OrganizationSlug: *in.SupabaseOrganizationSlug})
		},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.SupabaseConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read supabase connector settings: %w", err)
			}

			if s.OrganizationSlug == "" {
				return nil, fmt.Errorf("cannot create supabase driver: organization_slug is required")
			}

			return drivers.NewSupabaseDriver(c, s.OrganizationSlug), nil
		},
		NewNameResolver: func(ctx context.Context, _ *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.SupabaseConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read supabase connector settings", log.Error(err))
				return nil
			}

			return drivers.NewSupabaseNameResolver(s.OrganizationSlug)
		},
	}
}
