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

import { ErrorLayout } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { RedirectToEmployeePortalQuery } from "#/__generated__/core/RedirectToEmployeePortalQuery.graphql";
import { ExternalRedirect } from "#/components/ExternalRedirect";
import { employeePortalHref } from "#/lib/employeePortalHref";

export const redirectToEmployeePortalQuery = graphql`
  query RedirectToEmployeePortalQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        employeePortals(
          first: 1
          orderBy: { field: CREATED_AT, direction: ASC }
        ) {
          edges {
            node {
              id
            }
          }
        }
      }
    }
  }
`;

export interface RedirectToEmployeePortalProps {
  queryRef: PreloadedQuery<RedirectToEmployeePortalQuery>;
}

export function RedirectToEmployeePortal({ queryRef }: RedirectToEmployeePortalProps) {
  const { t } = useTranslation();
  const { organization } = usePreloadedQuery<RedirectToEmployeePortalQuery>(
    redirectToEmployeePortalQuery,
    queryRef,
  );

  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for organization node");
  }

  const employeePortalId = organization.employeePortals?.edges[0]?.node.id ?? null;

  if (employeePortalId == null) {
    return (
      <ErrorLayout
        title={t("redirectToEmployeePortal.missing.title")}
        description={t("redirectToEmployeePortal.missing.description")}
      />
    );
  }

  return <ExternalRedirect to={employeePortalHref(employeePortalId)} />;
}
