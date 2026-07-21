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

import {
  formatError,
  getStatusVariant,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Badge,
  Button,
  Card,
  DropdownItem,
  IconPageTextLine,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  Option,
  PageHeader,
  Select,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Suspense, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";

import type { FindingsPageDeleteMutation } from "#/__generated__/core/FindingsPageDeleteMutation.graphql";
import type { FindingsPageFragment$key } from "#/__generated__/core/FindingsPageFragment.graphql";
import type { FindingsPageListQuery } from "#/__generated__/core/FindingsPageListQuery.graphql";
import type {
  FindingKind,
  FindingPriority,
  FindingsPageRefetchQuery,
  FindingStatus,
} from "#/__generated__/core/FindingsPageRefetchQuery.graphql";
import type { FindingsPageRowFragment$key } from "#/__generated__/core/FindingsPageRowFragment.graphql";
import { usePeople } from "#/hooks/graph/PeopleGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateFindingDialog } from "./dialogs/CreateFindingDialog";
import { PublishFindingListDialog } from "./dialogs/PublishFindingListDialog";

export const FindingsConnectionKey = "FindingsPage_findings";

export const findingsPageQuery = graphql`
  query FindingsPageListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        canCreateFinding: permission(action: "core:finding:create")
        canPublishFindings: permission(action: "core:finding:publish")
        findingsDocument {
          id
          defaultApprovers {
            id
          }
        }
        ...FindingsPageFragment
      }
    }
  }
`;

const deleteFindingMutation = graphql`
  mutation FindingsPageDeleteMutation(
    $input: DeleteFindingInput!
    $connections: [ID!]!
  ) {
    deleteFinding(input: $input) {
      deletedFindingId @deleteEdge(connections: $connections)
    }
  }
`;

const findingRowFragment = graphql`
  fragment FindingsPageRowFragment on Finding {
    id
    kind
    referenceId
    description
    status
    priority
    dueDate
    owner {
      id
      fullName
    }
    canUpdate: permission(action: "core:finding:update")
    canDelete: permission(action: "core:finding:delete")
  }
`;

const findingsPageFragment = graphql`
  fragment FindingsPageFragment on Organization
  @refetchable(queryName: "FindingsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey" }
    kind: { type: "FindingKind", defaultValue: null }
    status: { type: "FindingStatus", defaultValue: null }
    priority: { type: "FindingPriority", defaultValue: null }
    ownerId: { type: "ID", defaultValue: null }
  ) {
    id
    findings(
      first: $first
      after: $after
      filter: {
        kind: $kind
        status: $status
        priority: $priority
        ownerId: $ownerId
      }
    )
      @connection(
        key: "FindingsPage_findings"
        filters: ["filter"]
      ) {
      edges {
        node {
          id
          canUpdate: permission(action: "core:finding:update")
          canDelete: permission(action: "core:finding:delete")
          ...FindingsPageRowFragment
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

interface FindingsPageProps {
  queryRef: PreloadedQuery<FindingsPageListQuery>;
}

export default function FindingsPage({ queryRef }: FindingsPageProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();

  usePageTitle(t("findingsPage.pageTitle"));

  const navigate = useNavigate();
  const organization = usePreloadedQuery<FindingsPageListQuery>(findingsPageQuery, queryRef);
  const defaultApproverIds = (organization.node.findingsDocument?.defaultApprovers ?? []).map(a => a.id);

  const [isPending, startTransition] = useTransition();
  const [kindFilter, setKindFilter] = useState<FindingKind | null>(null);
  const [statusFilter, setStatusFilter] = useState<FindingStatus | null>(null);
  const [priorityFilter, setPriorityFilter] = useState<FindingPriority | null>(null);
  const [ownerFilter, setOwnerFilter] = useState<string | null>(null);

  const { data, loadNext, hasNext, isLoadingNext, refetch }
    = usePaginationFragment<FindingsPageRefetchQuery, FindingsPageFragment$key>(
      findingsPageFragment,
      organization.node,
    );

  const refetchFilters = (overrides: Record<string, unknown> = {}) => {
    startTransition(() => {
      refetch(
        {
          kind: kindFilter,
          status: statusFilter,
          priority: priorityFilter,
          ownerId: ownerFilter,
          ...overrides,
        },
        { fetchPolicy: "network-only" },
      );
    });
  };

  const handleKindFilterChange = (value: string) => {
    const newKind = value === "ALL" ? null : (value as FindingKind);
    setKindFilter(newKind);
    refetchFilters({ kind: newKind });
  };

  const handleStatusFilterChange = (value: string) => {
    const newStatus = value === "ALL" ? null : (value as FindingStatus);
    setStatusFilter(newStatus);
    refetchFilters({ status: newStatus });
  };

  const handlePriorityFilterChange = (value: string) => {
    const newPriority = value === "ALL" ? null : (value as FindingPriority);
    setPriorityFilter(newPriority);
    refetchFilters({ priority: newPriority });
  };

  const handleOwnerFilterChange = (value: string) => {
    const newOwner = value === "ALL" ? null : value;
    setOwnerFilter(newOwner);
    refetchFilters({ ownerId: newOwner });
  };

  const currentFilter = {
    kind: kindFilter,
    status: statusFilter,
    priority: priorityFilter,
    ownerId: ownerFilter,
  };

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    FindingsConnectionKey,
    { filter: currentFilter },
  );
  const allFiltersNullConnectionId = ConnectionHandler.getConnectionID(
    organizationId,
    FindingsConnectionKey,
    {
      filter: {
        kind: null,
        status: null,
        priority: null,
        ownerId: null,
      },
    },
  );
  const hasActiveFilter = kindFilter || statusFilter || priorityFilter || ownerFilter;
  const createConnectionIds = hasActiveFilter
    ? [allFiltersNullConnectionId, connectionId]
    : [connectionId];
  const findings = data?.findings?.edges?.map(edge => edge.node) ?? [];

  const hasAnyAction
    = findings.some(({ canDelete, canUpdate }) => canDelete || canUpdate);

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("findingsPage.title")}
        description={t("findingsPage.description")}
      >
        <div className="flex gap-2">
          {organization.node.findingsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.findingsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {t("findingsPage.actions.document")}
              </Link>
            </Button>
          )}
          {organization.node.canPublishFindings && (
            <PublishFindingListDialog
              organizationId={organizationId}
              defaultApproverIds={defaultApproverIds}
              onPublished={(documentId) => {
                void navigate(
                  `/organizations/${organizationId}/documents/${documentId}`,
                );
              }}
            >
              <Button variant="secondary" icon={IconUpload}>
                {t("findingsPage.actions.publish")}
              </Button>
            </PublishFindingListDialog>
          )}
          {organization.node.canCreateFinding && (
            <CreateFindingDialog
              organizationId={organizationId}
              connectionIds={createConnectionIds}
            >
              <Button icon={IconPlusLarge}>{t("findingsPage.actions.add")}</Button>
            </CreateFindingDialog>
          )}
        </div>
      </PageHeader>

      <div className="flex items-center gap-4">
        <Select
          value={kindFilter ?? "ALL"}
          onValueChange={handleKindFilterChange}
        >
          <Option value="ALL">{t("findingsPage.filters.allKinds")}</Option>
          <Option value="MINOR_NONCONFORMITY">{t("findingsPage.kinds.minorNonconformity")}</Option>
          <Option value="MAJOR_NONCONFORMITY">{t("findingsPage.kinds.majorNonconformity")}</Option>
          <Option value="OBSERVATION">{t("findingsPage.kinds.observation")}</Option>
          <Option value="EXCEPTION">{t("findingsPage.kinds.exception")}</Option>
        </Select>
        <Select
          value={statusFilter ?? "ALL"}
          onValueChange={handleStatusFilterChange}
        >
          <Option value="ALL">{t("findingsPage.filters.allStatuses")}</Option>
          {(["OPEN", "IN_PROGRESS", "CLOSED", "RISK_ACCEPTED", "MITIGATED", "FALSE_POSITIVE"] as const).map(status => (
            <Option key={status} value={status}>
              {t(`findingsPage.status.${status.toLowerCase()}`)}
            </Option>
          ))}
        </Select>
        <Select
          value={priorityFilter ?? "ALL"}
          onValueChange={handlePriorityFilterChange}
        >
          <Option value="ALL">{t("findingsPage.filters.allPriorities")}</Option>
          <Option value="LOW">{t("findingsPage.priority.low")}</Option>
          <Option value="MEDIUM">{t("findingsPage.priority.medium")}</Option>
          <Option value="HIGH">{t("findingsPage.priority.high")}</Option>
        </Select>
        <Suspense fallback={<Select loading placeholder={t("findingsPage.loading")} />}>
          <OwnerFilterSelect
            organizationId={organizationId}
            value={ownerFilter}
            onChange={handleOwnerFilterChange}
          />
        </Suspense>
      </div>

      <div className={isPending ? "opacity-50 pointer-events-none transition-opacity" : ""}>
        {findings.length > 0
          ? (
              <Card>
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{t("findingsPage.columns.kind")}</Th>
                      <Th>{t("findingsPage.columns.referenceId")}</Th>
                      <Th>{t("findingsPage.columns.description")}</Th>
                      <Th>{t("findingsPage.columns.status")}</Th>
                      <Th>{t("findingsPage.columns.priority")}</Th>
                      <Th>{t("findingsPage.columns.owner")}</Th>
                      <Th>{t("findingsPage.columns.dueDate")}</Th>
                      {hasAnyAction && <Th>{t("findingsPage.columns.actions")}</Th>}
                    </Tr>
                  </Thead>
                  <Tbody>
                    {findings.map(finding => (
                      <FindingRow
                        key={finding.id}
                        findingKey={finding}
                        connectionId={connectionId}
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
                      {isLoadingNext ? t("findingsPage.loading") : t("findingsPage.actions.loadMore")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {t("findingsPage.empty.title")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {t("findingsPage.empty.description")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

type FindingRowProps = {
  findingKey: FindingsPageRowFragment$key;
  connectionId: string;
  hasAnyAction: boolean;
};

function FindingRow(props: FindingRowProps) {
  const finding = useFragment(findingRowFragment, props.findingKey);
  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation();
  const [deleteFinding] = useMutation<FindingsPageDeleteMutation>(deleteFindingMutation);
  const { toast } = useToast();
  const confirm = useConfirm();

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteFinding({
            variables: {
              input: {
                findingId: finding.id,
              },
              connections: [props.connectionId],
            },
            onCompleted(_, error) {
              if (error) {
                toast({
                  title: t("findingsPage.errors.title"),
                  description: formatError(
                    t("findingsPage.errors.delete"),
                    error,
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: t("findingsPage.messages.successTitle"),
                  description: t("findingsPage.messages.deleted"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: t("findingsPage.errors.title"),
                description: formatError(
                  t("findingsPage.errors.delete"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("findingsPage.deleteConfirmation", {
          referenceId: finding.referenceId,
        }),
      },
    );
  };

  const detailsUrl = `/organizations/${organizationId}/findings/${finding.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <Badge variant="neutral">
          {t(`findingsPage.kinds.${finding.kind.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>
        <span className="font-mono text-sm">{finding.referenceId}</span>
      </Td>
      <Td>
        <div className="min-w-0">
          <p className="whitespace-pre-wrap break-words">
            {finding.description || t("findingsPage.noDescription")}
          </p>
        </div>
      </Td>
      <Td>
        <Badge variant={getStatusVariant(finding.status)}>
          {t(`findingsPage.status.${finding.status.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>
        <Badge
          variant={
            finding.priority === "HIGH"
              ? "danger"
              : finding.priority === "MEDIUM"
                ? "warning"
                : "success"
          }
        >
          {finding.priority === "HIGH"
            ? t("findingsPage.priority.high")
            : finding.priority === "MEDIUM"
              ? t("findingsPage.priority.medium")
              : t("findingsPage.priority.low")}
        </Badge>
      </Td>
      <Td>{finding.owner?.fullName || "-"}</Td>
      <Td>
        {finding.dueDate
          ? (
              <time dateTime={finding.dueDate}>
                {dateFormat(i18n.language, finding.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{t("findingsPage.noDueDate")}</span>
            )}
      </Td>
      {props.hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {finding.canDelete && (
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onSelect={handleDelete}
              >
                {t("findingsPage.actions.delete")}
              </DropdownItem>
            )}
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}

type OwnerFilterSelectProps = {
  organizationId: string;
  value: string | null;
  onChange: (value: string) => void;
};

function OwnerFilterSelect({
  organizationId,
  value,
  onChange,
}: OwnerFilterSelectProps) {
  const { t } = useTranslation();
  const people = usePeople(organizationId, { contractEnded: false });

  return (
    <Select value={value ?? "ALL"} onValueChange={onChange}>
      <Option value="ALL">{t("findingsPage.filters.allOwners")}</Option>
      {people.map(p => (
        <Option key={p.id} value={p.id}>
          <Avatar name={p.fullName} />
          {p.fullName}
        </Option>
      ))}
    </Select>
  );
}
