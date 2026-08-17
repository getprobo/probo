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
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { RegistriesNavPanel_organization$key } from "#/__generated__/iam/RegistriesNavPanel_organization.graphql";
import type { RegistriesNavPanelQuery } from "#/__generated__/iam/RegistriesNavPanelQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navHref } from "#/pages/iam/organizations/_lib/navigation";

import { NavPanelItem } from "./NavPanelItem";
import { NavPanelQuery } from "./NavPanelQuery";
import type { NavPanelBodyProps } from "./navPanels";

const registriesNavPanelQuery = graphql`
  query RegistriesNavPanelQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...RegistriesNavPanel_organization
      }
    }
  }
`;

const registriesNavPanelFragment = graphql`
  fragment RegistriesNavPanel_organization on Organization {
    canListData: permission(action: "core:datum:list")
    canListAssets: permission(action: "core:asset:list")
    canListBusinessFunctions: permission(action: "core:business-function:list")
    canListAiSystems: permission(action: "core:ai-system:list")
    canListObligations: permission(action: "core:obligation:list")
  }
`;

export function RegistriesNavPanel({ group }: NavPanelBodyProps) {
  return (
    <NavPanelQuery<RegistriesNavPanelQuery> query={registriesNavPanelQuery}>
      {queryRef => <RegistriesNavPanelInner queryRef={queryRef} group={group} />}
    </NavPanelQuery>
  );
}

interface RegistriesNavPanelInnerProps extends NavPanelBodyProps {
  queryRef: PreloadedQuery<RegistriesNavPanelQuery>;
}

function RegistriesNavPanelInner({ queryRef, group }: RegistriesNavPanelInnerProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<RegistriesNavPanelQuery>(registriesNavPanelQuery, queryRef);
  if (data.organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }
  const organization = useFragment<RegistriesNavPanel_organization$key>(
    registriesNavPanelFragment,
    data.organization,
  );

  return (
    <>
      {organization.canListData && (
        <NavPanelItem
          label={t("nav.data")}
          to={navHref(organizationId, group, "data")}
        />
      )}
      {organization.canListAssets && (
        <NavPanelItem
          label={t("nav.assets")}
          to={navHref(organizationId, group, "assets")}
        />
      )}
      {organization.canListBusinessFunctions && (
        <NavPanelItem
          label={t("nav.businessFunctions")}
          to={navHref(organizationId, group, "business-functions")}
        />
      )}
      {organization.canListAiSystems && (
        <NavPanelItem
          label={t("nav.aiSystems")}
          to={navHref(organizationId, group, "ai-systems")}
        />
      )}
      {organization.canListObligations && (
        <NavPanelItem
          label={t("nav.obligations")}
          to={navHref(organizationId, group, "obligations")}
        />
      )}
    </>
  );
}
