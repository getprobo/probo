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

import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { PrivacyNavPanelQuery } from "#/__generated__/iam/PrivacyNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelItem } from "./NavPanelItem";
import { NavPanelQuery } from "./NavPanelQuery";
import type { NavPanelBodyProps } from "./navPanels";

const privacyNavPanelQuery = graphql`
  query PrivacyNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canListRightsRequests: permission(action: "core:rights-request:list")
        canListProcessingActivities: permission(action: "core:processing-activity:list")
        canListDataProtectionImpactAssessments: permission(action: "core:data-protection-impact-assessment:list")
        canListTransferImpactAssessments: permission(action: "core:transfer-impact-assessment:list")
      }
    }
  }
`;

export function PrivacyNavPanel({ group }: NavPanelBodyProps) {
  return (
    <NavPanelQuery<PrivacyNavPanelQuery> query={privacyNavPanelQuery}>
      {queryRef => <PrivacyNavPanelInner queryRef={queryRef} group={group} />}
    </NavPanelQuery>
  );
}

interface PrivacyNavPanelInnerProps extends NavPanelBodyProps {
  queryRef: PreloadedQuery<PrivacyNavPanelQuery>;
}

function PrivacyNavPanelInner({ queryRef, group }: PrivacyNavPanelInnerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<PrivacyNavPanelQuery>(privacyNavPanelQuery, queryRef);
  const { organization } = data;
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <>
      {organization.canListRightsRequests && (
        <NavPanelItem
          label={t("nav.rightsRequests")}
          to={navHref(organizationId, group, "rights-requests")}
        />
      )}
      {organization.canListProcessingActivities && (
        <NavPanelItem
          label={t("nav.processingActivities")}
          to={navHref(organizationId, group, "processing-activities")}
        />
      )}
      {organization.canListDataProtectionImpactAssessments && (
        <NavPanelItem
          label={t("nav.dataProtectionImpactAssessments")}
          to={navHref(organizationId, group, "dpias")}
        />
      )}
      {organization.canListTransferImpactAssessments && (
        <NavPanelItem
          label={t("nav.transferImpactAssessments")}
          to={navHref(organizationId, group, "tias")}
        />
      )}
    </>
  );
}
