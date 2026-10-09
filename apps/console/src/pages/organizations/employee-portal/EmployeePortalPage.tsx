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

import { usePageTitle } from "@probo/hooks";
import { ErrorLayout } from "@probo/ui";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeePortalPageQuery } from "#/__generated__/core/EmployeePortalPageQuery.graphql";

import { EmployeePortalBrandingSection } from "./_components/EmployeePortalBrandingSection";
import { EmployeePortalCapabilitiesSection } from "./_components/EmployeePortalCapabilitiesSection";
import { employeePortalPage } from "./variants";

const NS = "organizations/employee-portal";

export const employeePortalPageQuery = graphql`
  query EmployeePortalPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        employeePortals(
          first: 1
          orderBy: { field: CREATED_AT, direction: ASC }
        ) {
          edges {
            node {
              ...EmployeePortalBrandingSection_employeePortal
              ...EmployeePortalCapabilitiesSection_employeePortal
            }
          }
        }
      }
    }
  }
`;

interface EmployeePortalPageProps {
  queryRef: PreloadedQuery<EmployeePortalPageQuery>;
}

export function EmployeePortalPage({ queryRef }: EmployeePortalPageProps) {
  const { t } = useTranslation(NS);
  const { root, header } = employeePortalPage();
  const title = t("employeePortalPage.title");
  usePageTitle(title);

  const { organization } = usePreloadedQuery<EmployeePortalPageQuery>(
    employeePortalPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const employeePortal = organization.employeePortals.edges[0]?.node;
  if (employeePortal == null) {
    return (
      <ErrorLayout
        title={t("missing.title")}
        description={t("missing.description")}
      />
    );
  }

  return (
    <div className={root()}>
      <div className={header()}>
        <Heading level={1} size={6} weight="medium" highContrast>
          {title}
        </Heading>
        <Text size={2} color="faint">
          {t("employeePortalPage.description")}
        </Text>
      </div>
      <EmployeePortalBrandingSection employeePortalKey={employeePortal} />
      <EmployeePortalCapabilitiesSection employeePortalKey={employeePortal} />
    </div>
  );
}
