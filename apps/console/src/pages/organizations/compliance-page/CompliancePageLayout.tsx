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
import { Badge, Button, IconBell2, IconCheckmark1, IconFolder2, IconMedal, IconPageTextLine, IconPencil, IconPeopleAdd, IconSettingsGear2, IconStore, PageHeader, TabLink, Tabs } from "@probo/ui";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePageLayoutQuery } from "#/__generated__/core/CompliancePageLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const compliancePageLayoutQuery = graphql`
  query CompliancePageLayoutQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        customDomain {
          domain
        }
        compliancePage: trustCenter {
          id
          active
        }
      }
    }
  }
`;

export function CompliancePageLayout(props: { queryRef: PreloadedQuery<CompliancePageLayoutQuery> }) {
  const { queryRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  usePageTitle(__("Compliance Page"));

  const { organization } = usePreloadedQuery<CompliancePageLayoutQuery>(compliancePageLayoutQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const compliancePageUrl = organization.compliancePage?.id
    ? organization.customDomain?.domain
      ? `https://${organization.customDomain.domain}`
      : `${window.location.origin}/trust/${organization.compliancePage.id}`
    : null;

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Compliance Page")}
        description={__(
          "Configure your public compliance page to showcase your security and compliance posture.",
        )}
      >
        <Badge variant={organization.compliancePage?.active ? "success" : "danger"}>
          {organization.compliancePage?.active ? __("Active") : __("Inactive")}
        </Badge>
        {organization.compliancePage?.active && compliancePageUrl && (
          <Button
            variant="secondary"
            onClick={() =>
              window.open(
                compliancePageUrl,
                "_blank",
                "noopener,noreferrer",
              )}
          >
            {__("Open")}
          </Button>
        )}
      </PageHeader>

      <Tabs>
        <TabLink to={`/organizations/${organizationId}/compliance-page`} end>
          <IconSettingsGear2 className="size-4" />
          {__("Overview")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/brand`}>
          <IconPencil className="size-4" />
          {__("Brand")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/domain`}>
          <IconStore className="size-4" />
          {__("Domain")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/references`}>
          <IconCheckmark1 className="size-4" />
          {__("References")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/audits`}>
          <IconMedal className="size-4" />
          {__("Audits")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/documents`}>
          <IconPageTextLine className="size-4" />
          {__("Documents")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/files`}>
          <IconFolder2 className="size-4" />
          {__("Files")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/third-parties`}>
          <IconStore className="size-4" />
          {__("Subprocessors")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/access`}>
          <IconPeopleAdd className="size-4" />
          {__("Access")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/mailing-list`}>
          <IconBell2 className="size-4" />
          {__("Mailing List")}
        </TabLink>
      </Tabs>

      <Outlet />
    </div>
  );
}
