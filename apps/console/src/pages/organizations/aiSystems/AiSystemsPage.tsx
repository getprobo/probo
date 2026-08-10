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

import type { AiSystemsPageFragment$key } from "#/__generated__/core/AiSystemsPageFragment.graphql";
import type { AiSystemsPageListQuery } from "#/__generated__/core/AiSystemsPageListQuery.graphql";
import type {
  AiSystemRiskClassification,
  AiSystemsPageRefetchQuery,
  AiSystemStatus,
} from "#/__generated__/core/AiSystemsPageRefetchQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AiSystemListItem } from "./_components/AiSystemListItem";
import { CreateAiSystemDialog } from "./_components/CreateAiSystemDialog";
import { PublishAiSystemListDialog } from "./_components/PublishAiSystemListDialog";
import {
  AI_SYSTEM_RISK_CLASSIFICATIONS,
  AI_SYSTEM_STATUSES,
  AiSystemsConnectionKey,
  emptyAiSystemFilter,
  getRiskClassificationLabel,
  getStatusLabel,
} from "./_lib/aiSystemHelpers";

export const aiSystemsPageQuery = graphql`
  query AiSystemsPageListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateAiSystem: permission(action: "core:ai-system:create")
        canPublishAiSystems: permission(action: "core:ai-system:publish")
        aiSystemsDocument {
          id
        }
        ...AiSystemsPageFragment
        ...PublishAiSystemListDialog_organization
      }
    }
  }
`;

const aiSystemsPageFragment = graphql`
  fragment AiSystemsPageFragment on Organization
  @refetchable(queryName: "AiSystemsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    status: { type: "AiSystemStatus", defaultValue: null }
    riskClassification: { type: "AiSystemRiskClassification", defaultValue: null }
  ) {
    id
    aiSystems(
      first: $first
      after: $after
      filter: {
        status: $status
        riskClassification: $riskClassification
      }
    )
      @connection(
        key: "AiSystemsPage_aiSystems"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canDelete: permission(action: "core:ai-system:delete")
          ...AiSystemListItem_aiSystem
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface AiSystemsPageProps {
  queryRef: PreloadedQuery<AiSystemsPageListQuery>;
}

export function AiSystemsPage({ queryRef }: AiSystemsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const [isPending, startTransition] = useTransition();

  usePageTitle(t("aiSystemsPage.pageTitle"));

  const organization = usePreloadedQuery<AiSystemsPageListQuery>(
    aiSystemsPageQuery,
    queryRef,
  );

  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<AiSystemsPageRefetchQuery, AiSystemsPageFragment$key>(
      aiSystemsPageFragment,
      organization.node,
    );

  const urlStatusParam = searchParams.get("status");
  const urlStatus: AiSystemStatus | null
    = urlStatusParam && AI_SYSTEM_STATUSES.includes(urlStatusParam as AiSystemStatus)
      ? urlStatusParam as AiSystemStatus
      : null;
  const urlRiskParam = searchParams.get("risk");
  const urlRisk: AiSystemRiskClassification | null
    = urlRiskParam
      && AI_SYSTEM_RISK_CLASSIFICATIONS.includes(urlRiskParam as AiSystemRiskClassification)
      ? urlRiskParam as AiSystemRiskClassification
      : null;

  const initialUrlFilters = useRef({ status: urlStatus, risk: urlRisk });
  const prevUrlFilters = useRef({ status: urlStatus, risk: urlRisk });

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      refetch(
        {
          status: urlStatus,
          riskClassification: urlRisk,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  useEffect(() => {
    const { status, risk } = initialUrlFilters.current;
    if (status || risk) {
      startTransition(() => {
        refetch(
          { status, riskClassification: risk },
          { fetchPolicy: "network-only" },
        );
      });
    }
  }, [refetch, startTransition]);

  useEffect(() => {
    if (
      urlStatus !== prevUrlFilters.current.status
      || urlRisk !== prevUrlFilters.current.risk
    ) {
      prevUrlFilters.current = { status: urlStatus, risk: urlRisk };
      refetchFilters({ status: urlStatus, riskClassification: urlRisk });
    }
  });

  const handleStatusFilterChange = (value: string) => {
    const newStatus = value === "ALL" ? null : (value as AiSystemStatus);
    updateSearchParam(setSearchParams, "status", newStatus);
  };

  const handleRiskFilterChange = (value: string) => {
    const newRisk = value === "ALL" ? null : (value as AiSystemRiskClassification);
    updateSearchParam(setSearchParams, "risk", newRisk);
  };

  const currentFilter = {
    status: urlStatus,
    riskClassification: urlRisk,
  };

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    AiSystemsConnectionKey,
    { filter: currentFilter },
  );
  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    AiSystemsConnectionKey,
    { filter: emptyAiSystemFilter },
  );
  const hasActiveFilter = urlStatus || urlRisk;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId]
    : [connectionId];
  const aiSystems = data?.aiSystems?.edges?.map(edge => edge.node) ?? [];
  const hasAnyAction = aiSystems.some(({ canDelete }) => canDelete);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("aiSystemsPage.title")}
        description={t("aiSystemsPage.description")}
      >
        <div className="flex gap-2">
          {organization.node.aiSystemsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.aiSystemsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("aiSystemsPage.actions.document")}
              </Link>
            </Button>
          )}
          {organization.node.canPublishAiSystems && (
            <PublishAiSystemListDialog
              organizationKey={organization.node}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("aiSystemsPage.actions.publish")}
              </Button>
            </PublishAiSystemListDialog>
          )}
          {organization.node.canCreateAiSystem && (
            <CreateAiSystemDialog connectionIds={createConnectionIds}>
              <Button icon={IconPlusLarge}>{t("aiSystemsPage.actions.add")}</Button>
            </CreateAiSystemDialog>
          )}
        </div>
      </PageHeader>

      <div className="flex items-center gap-4">
        <Select
          value={urlStatus ?? "ALL"}
          onValueChange={handleStatusFilterChange}
        >
          <Option value="ALL">{t("aiSystemsPage.filters.allStatuses")}</Option>
          {AI_SYSTEM_STATUSES.map(status => (
            <Option key={status} value={status}>
              {getStatusLabel(status, t, "aiSystemsPage")}
            </Option>
          ))}
        </Select>
        <Select
          value={urlRisk ?? "ALL"}
          onValueChange={handleRiskFilterChange}
        >
          <Option value="ALL">{t("aiSystemsPage.filters.allRiskClassifications")}</Option>
          {AI_SYSTEM_RISK_CLASSIFICATIONS.map(classification => (
            <Option key={classification} value={classification}>
              {getRiskClassificationLabel(classification, t, "aiSystemsPage")}
            </Option>
          ))}
        </Select>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {aiSystems.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{t("aiSystemsPage.columns.name")}</Th>
                      <Th>{t("aiSystemsPage.columns.version")}</Th>
                      <Th>{t("aiSystemsPage.columns.status")}</Th>
                      <Th>{t("aiSystemsPage.columns.riskClassification")}</Th>
                      <Th>{t("aiSystemsPage.columns.owner")}</Th>
                      <Th>{t("aiSystemsPage.columns.nextReviewDate")}</Th>
                      {hasAnyAction && <Th>{t("aiSystemsPage.columns.actions")}</Th>}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {aiSystems.map(aiSystem => (
                      <AiSystemListItem
                        key={aiSystem.id}
                        aiSystemKey={aiSystem}
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
                        ? t("aiSystemsPage.actions.loading")
                        : t("aiSystemsPage.actions.loadMore")}
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
                      ? t("aiSystemsPage.empty.filteredTitle")
                      : t("aiSystemsPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {hasActiveFilter
                      ? t("aiSystemsPage.empty.filteredDescription")
                      : t("aiSystemsPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

function updateSearchParam(
  setSearchParams: ReturnType<typeof useSearchParams>[1],
  key: string,
  value: string | null,
) {
  setSearchParams((params) => {
    const next = new URLSearchParams(params);
    if (value) {
      next.set(key, value);
    } else {
      next.delete(key);
    }
    return next;
  }, { replace: true });
}
