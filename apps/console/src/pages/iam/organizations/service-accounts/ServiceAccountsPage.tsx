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

import { PlusIcon, RobotIcon } from "@phosphor-icons/react";
import { usePageTitle } from "@probo/hooks";
import { PageHeader } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Table } from "@probo/ui/src/v2/Table/Table";
import { TableBody } from "@probo/ui/src/v2/Table/TableBody";
import { TableColumnHeaderCell } from "@probo/ui/src/v2/Table/TableColumnHeaderCell";
import { TableHeader } from "@probo/ui/src/v2/Table/TableHeader";
import { TableRow } from "@probo/ui/src/v2/Table/TableRow";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { ServiceAccountsPage_serviceAccounts$key } from "#/__generated__/iam/ServiceAccountsPage_serviceAccounts.graphql";
import type { ServiceAccountsPagePaginationQuery } from "#/__generated__/iam/ServiceAccountsPagePaginationQuery.graphql";
import type { ServiceAccountsPageQuery } from "#/__generated__/iam/ServiceAccountsPageQuery.graphql";

import { CreateServiceAccountDialog } from "./_components/CreateServiceAccountDialog";
import { ServiceAccountListItem } from "./_components/ServiceAccountListItem";

export const serviceAccountsPageQuery = graphql`
  query ServiceAccountsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        canCreateServiceAccount: permission(
          action: "iam:service-account:create"
        )
        ...ServiceAccountsPage_serviceAccounts
          @arguments(first: 50)
      }
    }
  }
`;

const serviceAccountsFragment = graphql`
  fragment ServiceAccountsPage_serviceAccounts on Organization
  @refetchable(queryName: "ServiceAccountsPagePaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    serviceAccounts(
      first: $first
      after: $after
      orderBy: { field: NAME, direction: ASC }
    ) @connection(key: "ServiceAccountsPage_serviceAccounts", filters: []) {
      __id
      edges {
        node {
          id
          ...ServiceAccountListItem_serviceAccount
        }
      }
    }
  }
`;

interface ServiceAccountsPageProps {
  queryRef: PreloadedQuery<ServiceAccountsPageQuery>;
}

export function ServiceAccountsPage({ queryRef }: ServiceAccountsPageProps) {
  const { t } = useTranslation("iam/organizations/service-accounts");
  const [createOpen, setCreateOpen] = useState(false);
  usePageTitle(t("page.title"));

  const { organization } = usePreloadedQuery<ServiceAccountsPageQuery>(
    serviceAccountsPageQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("Relay node is not an organization");
  }

  const pagination = usePaginationFragment<
    ServiceAccountsPagePaginationQuery,
    ServiceAccountsPage_serviceAccounts$key
  >(serviceAccountsFragment, organization);
  const serviceAccounts = pagination.data.serviceAccounts.edges;

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("page.title")}
        description={t("page.description")}
      >
        {organization.canCreateServiceAccount && (
          <Button
            iconStart={<PlusIcon aria-hidden />}
            onClick={() => setCreateOpen(true)}
          >
            {t("actions.create")}
          </Button>
        )}
      </PageHeader>

      {serviceAccounts.length === 0
        ? (
            <div className="flex flex-col items-center gap-3 rounded-4 border border-sand-6 px-6 py-12 text-center">
              <RobotIcon className="size-8 text-sand-10" aria-hidden />
              <Text size={3} weight="medium">
                {t("empty.title")}
              </Text>
              <Text size={2} color="faint">
                {t("empty.description")}
              </Text>
              {organization.canCreateServiceAccount && (
                <Button
                  variant="soft"
                  iconStart={<PlusIcon aria-hidden />}
                  onClick={() => setCreateOpen(true)}
                >
                  {t("actions.create")}
                </Button>
              )}
            </div>
          )
        : (
            <Table variant="surface">
              <TableHeader>
                <TableRow>
                  <TableColumnHeaderCell>{t("columns.name")}</TableColumnHeaderCell>
                  <TableColumnHeaderCell>{t("columns.status")}</TableColumnHeaderCell>
                  <TableColumnHeaderCell>{t("columns.scopes")}</TableColumnHeaderCell>
                  <TableColumnHeaderCell>{t("columns.createdAt")}</TableColumnHeaderCell>
                  <TableColumnHeaderCell>{t("columns.actions")}</TableColumnHeaderCell>
                </TableRow>
              </TableHeader>
              <TableBody>
                {serviceAccounts.map(({ node }) => (
                  <ServiceAccountListItem
                    key={node.id}
                    serviceAccountKey={node}
                    connectionId={pagination.data.serviceAccounts.__id}
                  />
                ))}
              </TableBody>
            </Table>
          )}

      {pagination.hasNext && (
        <Button
          variant="soft"
          loading={pagination.isLoadingNext}
          onClick={() => pagination.loadNext(50)}
        >
          {t("actions.loadMore")}
        </Button>
      )}

      <CreateServiceAccountDialog
        connectionId={pagination.data.serviceAccounts.__id}
        open={createOpen}
        onOpenChange={setCreateOpen}
      />
    </div>
  );
}
