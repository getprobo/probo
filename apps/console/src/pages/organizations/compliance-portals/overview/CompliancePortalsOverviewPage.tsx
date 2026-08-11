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
import { dateFormat } from "@probo/i18n";
import { Badge, Button, Card, PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalsOverviewPageQuery } from "#/__generated__/core/CompliancePortalsOverviewPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CompliancePortalEmptyState } from "./_components/CompliancePortalEmptyState";

export const compliancePortalsOverviewPageQuery = graphql`
  query CompliancePortalsOverviewPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePortals(first: 50, orderBy: { field: CREATED_AT, direction: DESC })
          @connection(key: "CompliancePortalsOverviewPage_compliancePortals", filters: [])
          @required(action: THROW) {
          __id
          edges {
            node {
              id
              entityName
              active
              publicUrl
              createdAt
            }
          }
        }
      }
    }
  }
`;

interface CompliancePortalsOverviewPageProps {
  queryRef: PreloadedQuery<CompliancePortalsOverviewPageQuery>;
}

export function CompliancePortalsOverviewPage({ queryRef }: CompliancePortalsOverviewPageProps) {
  const { t, i18n } = useTranslation("organizations/compliance-portals");
  const organizationId = useOrganizationId();

  usePageTitle(t("overviewPage.title"));

  const { organization } = usePreloadedQuery<CompliancePortalsOverviewPageQuery>(
    compliancePortalsOverviewPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const portals = organization.compliancePortals.edges.map(e => e.node);
  const newPortalHref = `/organizations/${organizationId}/compliance-portals/new`;

  if (portals.length === 0) {
    return (
      <div className="space-y-6">
        <PageHeader
          title={t("overviewPage.title")}
          description={t("overviewPage.description")}
        />
        <CompliancePortalEmptyState>
          <Button to={newPortalHref}>{t("overviewPage.actions.createFirst")}</Button>
        </CompliancePortalEmptyState>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("overviewPage.title")}
        description={t("overviewPage.description")}
      />

      <div className="space-y-4">
        <div className="flex justify-end">
          <Button to={newPortalHref}>{t("overviewPage.actions.create")}</Button>
        </div>

        <Card className="divide-y divide-border-low rounded-lg">
          {portals.map(portal => (
            <Link
              key={portal.id}
              to={`/organizations/${organizationId}/compliance-portals/${portal.id}`}
              className="flex items-center justify-between gap-4 p-4 hover:bg-muted/50 transition-colors"
            >
              <div className="min-w-0 flex-1">
                <div className="font-medium">{portal.entityName}</div>
                <div className="text-sm text-muted-foreground truncate">{portal.publicUrl}</div>
              </div>
              <div className="flex items-center gap-3">
                <Badge variant={portal.active ? "success" : "danger"}>
                  {portal.active
                    ? t("overviewPage.status.active")
                    : t("overviewPage.status.inactive")}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {dateFormat(i18n.language, portal.createdAt)}
                </span>
              </div>
            </Link>
          ))}
        </Card>
      </div>
    </div>
  );
}
