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
import {
  Button,
  Card,
  IconPageTextLine,
  IconPlusLarge,
  IconUpload,
  Option,
  PageHeader,
  Select,
  Table,
  Tbody,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { useEffect, useRef, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate, useSearchParams } from "react-router";

import type { BusinessFunctionsPageFragment$key } from "#/__generated__/core/BusinessFunctionsPageFragment.graphql";
import type { BusinessFunctionsPageListQuery } from "#/__generated__/core/BusinessFunctionsPageListQuery.graphql";
import type {
  BusinessFunctionClassification,
  BusinessFunctionsPageRefetchQuery,
} from "#/__generated__/core/BusinessFunctionsPageRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { BusinessFunctionListItem } from "./_components/BusinessFunctionListItem";
import { CreateBusinessFunctionDialog } from "./_components/CreateBusinessFunctionDialog";
import { PublishBusinessFunctionListDialog } from "./_components/PublishBusinessFunctionListDialog";
import {
  BusinessFunctionsConnectionKey,
  emptyBusinessFunctionFilter,
} from "./_lib/businessFunctionHelpers";

export const businessFunctionsPageQuery = graphql`
  query BusinessFunctionsPageListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateBusinessFunction: permission(action: "core:business-function:create")
        canPublishBusinessFunctions: permission(action: "core:business-function:publish")
        businessFunctionsDocument {
          id
        }
        ...BusinessFunctionsPageFragment
        ...PublishBusinessFunctionListDialog_organization
      }
    }
  }
`;

const businessFunctionsPageFragment = graphql`
  fragment BusinessFunctionsPageFragment on Organization
  @refetchable(queryName: "BusinessFunctionsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    classification: { type: "BusinessFunctionClassification", defaultValue: null }
  ) {
    id
    businessFunctions(
      first: $first
      after: $after
      filter: {
        classification: $classification
      }
    )
      @connection(
        key: "BusinessFunctionsPage_businessFunctions"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canDelete: permission(action: "core:business-function:delete")
          ...BusinessFunctionListItem_businessFunction
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface BusinessFunctionsPageProps {
  queryRef: PreloadedQuery<BusinessFunctionsPageListQuery>;
}

export function BusinessFunctionsPage({ queryRef }: BusinessFunctionsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  usePageTitle(t("businessFunctionsPage.title"));

  const navigate = useNavigate();
  const organization = usePreloadedQuery<BusinessFunctionsPageListQuery>(
    businessFunctionsPageQuery,
    queryRef,
  );

  const [isPending, startTransition] = useTransition();
  const [searchParams, setSearchParams] = useSearchParams();
  const urlClassificationParam = searchParams.get("classification");
  const urlClassification: BusinessFunctionClassification | null
    = urlClassificationParam === "CRITICAL"
      || urlClassificationParam === "IMPORTANT"
      || urlClassificationParam === "SECONDARY"
      || urlClassificationParam === "STANDARD"
      ? urlClassificationParam
      : null;

  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<
      BusinessFunctionsPageRefetchQuery,
      BusinessFunctionsPageFragment$key
    >(businessFunctionsPageFragment, organization.node);

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      refetch(
        {
          classification: urlClassification,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const initialUrlClassification = useRef(urlClassification);
  const prevUrlClassification = useRef(urlClassification);
  useEffect(() => {
    if (initialUrlClassification.current) {
      startTransition(() => {
        refetch(
          { classification: initialUrlClassification.current },
          { fetchPolicy: "network-only" },
        );
      });
    }
  }, [refetch, startTransition]);

  useEffect(() => {
    if (urlClassification !== prevUrlClassification.current) {
      prevUrlClassification.current = urlClassification;
      refetchFilters({ classification: urlClassification });
    }
  });

  const handleClassificationFilterChange = (value: string) => {
    const newClassification = value === "ALL"
      ? null
      : (value as BusinessFunctionClassification);
    setSearchParams((prev) => {
      const next = new URLSearchParams(prev);
      if (newClassification) {
        next.set("classification", newClassification);
      } else {
        next.delete("classification");
      }
      return next;
    }, { replace: true });
  };

  const currentFilter = {
    classification: urlClassification,
  };
  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    BusinessFunctionsConnectionKey,
    { filter: currentFilter },
  );
  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    BusinessFunctionsConnectionKey,
    { filter: emptyBusinessFunctionFilter },
  );
  const hasActiveFilter = urlClassification != null;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId, connectionId]
    : [connectionId];
  const businessFunctions = data?.businessFunctions?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction = businessFunctions.some(({ canDelete }) => canDelete);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("businessFunctionsPage.title")}
        description={t("businessFunctionsPage.description")}
      >
        <div className="flex gap-2">
          {organization.node.businessFunctionsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/governance/documents/${organization.node.businessFunctionsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("businessFunctionsPage.actions.document")}
              </Link>
            </Button>
          )}
          {organization.node.canPublishBusinessFunctions && (
            <PublishBusinessFunctionListDialog
              organizationKey={organization.node}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/governance/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("businessFunctionsPage.actions.publish")}
              </Button>
            </PublishBusinessFunctionListDialog>
          )}
          {organization.node.canCreateBusinessFunction && (
            <CreateBusinessFunctionDialog connectionIds={createConnectionIds}>
              <Button icon={IconPlusLarge}>{t("businessFunctionsPage.actions.add")}</Button>
            </CreateBusinessFunctionDialog>
          )}
        </div>
      </PageHeader>

      <div className="flex flex-wrap items-center gap-4">
        <Select
          value={urlClassification ?? "ALL"}
          onValueChange={handleClassificationFilterChange}
        >
          <Option value="ALL">{t("businessFunctionsPage.filters.allClassifications")}</Option>
          <Option value="CRITICAL">{t("businessFunctionsPage.classifications.critical")}</Option>
          <Option value="IMPORTANT">{t("businessFunctionsPage.classifications.important")}</Option>
          <Option value="SECONDARY">{t("businessFunctionsPage.classifications.secondary")}</Option>
          <Option value="STANDARD">{t("businessFunctionsPage.classifications.standard")}</Option>
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {businessFunctions.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{t("businessFunctionsPage.columns.referenceId")}</Th>
                      <Th>{t("businessFunctionsPage.columns.name")}</Th>
                      <Th>{t("businessFunctionsPage.columns.classification")}</Th>
                      <Th>{t("businessFunctionsPage.columns.mtd")}</Th>
                      <Th>{t("businessFunctionsPage.columns.rto")}</Th>
                      <Th>{t("businessFunctionsPage.columns.rpo")}</Th>
                      <Th>{t("businessFunctionsPage.columns.owner")}</Th>
                      <Th>{t("businessFunctionsPage.columns.assets")}</Th>
                      <Th>{t("businessFunctionsPage.columns.thirdParties")}</Th>
                      {hasAnyAction && <Th>{t("businessFunctionsPage.columns.actions")}</Th>}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {businessFunctions.map(businessFunction => (
                      <BusinessFunctionListItem
                        key={businessFunction.id}
                        businessFunctionKey={businessFunction}
                        hasAnyAction={hasAnyAction}
                      />
                    ))}
                  </Tbody>
                </Table>

                {hasNext && (
                  <div className="p-4 border-t">
                    <Button
                      variant="secondary"
                      onClick={() => loadNext(10)}
                      disabled={isLoadingNext}
                    >
                      {isLoadingNext
                        ? t("businessFunctionsPage.actions.loading")
                        : t("businessFunctionsPage.actions.loadMore")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {hasActiveFilter
                      ? t("businessFunctionsPage.empty.filteredTitle")
                      : t("businessFunctionsPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {hasActiveFilter
                      ? t("businessFunctionsPage.empty.filteredDescription")
                      : t("businessFunctionsPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}
