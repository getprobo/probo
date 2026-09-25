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

import type { GovernanceNavPanelQuery } from "#/__generated__/iam/GovernanceNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelItem } from "./NavPanelItem";
import { NavPanelQuery } from "./NavPanelQuery";
import type { NavPanelBodyProps } from "./navPanels";

const governanceNavPanelQuery = graphql`
  query GovernanceNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canListTasks: permission(action: "core:task:list")
        canListMeasures: permission(action: "core:measure:list")
        canListFrameworks: permission(action: "core:framework:list")
        canListAudits: permission(action: "core:audit:list")
        canListFindings: permission(action: "core:finding:list")
        canListDocuments: permission(action: "core:document:list")
        canListStatementsOfApplicability: permission(action: "core:statement-of-applicability:list")
      }
    }
  }
`;

export function GovernanceNavPanel({ group }: NavPanelBodyProps) {
  return (
    <NavPanelQuery<GovernanceNavPanelQuery> query={governanceNavPanelQuery}>
      {queryRef => <GovernanceNavPanelInner queryRef={queryRef} group={group} />}
    </NavPanelQuery>
  );
}

interface GovernanceNavPanelInnerProps extends NavPanelBodyProps {
  queryRef: PreloadedQuery<GovernanceNavPanelQuery>;
}

function GovernanceNavPanelInner({ queryRef, group }: GovernanceNavPanelInnerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<GovernanceNavPanelQuery>(governanceNavPanelQuery, queryRef);
  const { organization } = data;
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  return (
    <>
      {organization.canListTasks && (
        <NavPanelItem
          label={t("nav.tasks")}
          to={navHref(organizationId, group, "tasks")}
        />
      )}
      {organization.canListMeasures && (
        <NavPanelItem
          label={t("nav.measures")}
          to={navHref(organizationId, group, "measures")}
        />
      )}
      {organization.canListFrameworks && (
        <NavPanelItem
          label={t("nav.frameworks")}
          to={navHref(organizationId, group, "frameworks")}
        />
      )}
      {organization.canListAudits && (
        <NavPanelItem
          label={t("nav.audits")}
          to={navHref(organizationId, group, "audits")}
        />
      )}
      {organization.canListFindings && (
        <NavPanelItem
          label={t("nav.findings")}
          to={navHref(organizationId, group, "findings")}
        />
      )}
      {organization.canListDocuments && (
        <NavPanelItem
          label={t("nav.documents")}
          to={navHref(organizationId, group, "documents")}
        />
      )}
      {organization.canListStatementsOfApplicability && (
        <NavPanelItem
          label={t("nav.statementsOfApplicability")}
          to={navHref(organizationId, group, "statements-of-applicability")}
        />
      )}
    </>
  );
}
