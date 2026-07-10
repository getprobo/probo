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

import { safeOpenUrl } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconBell2, IconCheckmark1, IconFolder2, IconMedal, IconPageTextLine, IconPencil, IconPeopleAdd, IconSettingsGear2, IconShield, IconStore, PageHeader, TabLink, Tabs } from "@probo/ui";
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
        compliancePage: trustCenter {
          id
          active
          publicUrl
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

  const compliancePageUrl = organization.compliancePage?.publicUrl || null;

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
            onClick={() => safeOpenUrl(compliancePageUrl)}
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
        <TabLink to={`/organizations/${organizationId}/compliance-page/references`}>
          <IconCheckmark1 className="size-4" />
          {__("References")}
        </TabLink>
        <TabLink to={`/organizations/${organizationId}/compliance-page/commitments`}>
          <IconShield className="size-4" />
          {__("Commitments")}
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
