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
      }}
      />
    </div>
  );
}
