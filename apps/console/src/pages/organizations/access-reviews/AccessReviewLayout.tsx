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

import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  IconFolder2,
  IconKey,
  PageHeader,
  TabLink,
  Tabs,
} from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { AccessReviewLayoutQuery } from "#/__generated__/core/AccessReviewLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const accessReviewLayoutQuery = graphql`
  query AccessReviewLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        id
        canCreateSource: permission(action: "core:access-source:create")
        canCreateCampaign: permission(action: "core:access-review-campaign:create")
        connectorProviderInfos {
          provider
          displayName
          oauthConfigured
          apiKeySupported
          clientCredentialsSupported
          extraSettings {
            key
            label
            required
          }
        }
        ...AccessReviewCampaignsTabFragment
        ...AccessReviewSourcesTabFragment
      }
    }
  }
`;

export default function AccessReviewLayout({
  queryRef,
}: {
  queryRef: PreloadedQuery<AccessReviewLayoutQuery>;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Access Reviews"));

  const { organization } = usePreloadedQuery(accessReviewLayoutQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("Organization not found");
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Access Reviews")}
        description={__(
          "Review and manage user access across your organization's systems and applications.",
        )}
      />

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/access-reviews`} end>
          <IconKey className="size-4" />
          {__("Campaigns")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/access-reviews/sources`}>
          <IconFolder2 className="size-4" />
          {__("Sources")}
        </TabLink>
      </Tabs>

      <Outlet context={{
        organizationRef: organization,
        canCreateSource: organization.canCreateSource,
        canCreateCampaign: organization.canCreateCampaign,
        connectorProviderInfos: organization.connectorProviderInfos,
      }}
      />
    </div>
  );
}
