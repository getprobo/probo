// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import {
  formatDate,
  formatError,
  getStatusLabel,
  getStatusOptions,
  getStatusVariant,
  type GraphQLError,
  sprintf,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  usePageTitle(__("Findings"));

  const navigate = useNavigate();
  const organization = usePreloadedQuery(findingsPageQuery, queryRef);
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
        title={__("Findings")}
        description={__("Manage your organization's findings.")}
      >
        <div className="flex gap-2">
          {organization.node.findingsDocument?.id && (
            <Button variant="secondary" asChild>
              <Link
                to={`/organizations/${organizationId}/documents/${organization.node.findingsDocument.id}`}
              >
                <IconPageTextLine size={16} />
                {__("Document")}
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
                {__("Publish")}
              </Button>
            </PublishFindingListDialog>
          )}
          {organization.node.canCreateFinding && (
            <CreateFindingDialog
              organizationId={organizationId}
              connectionIds={createConnectionIds}
            >
              <Button icon={IconPlusLarge}>{__("Add finding")}</Button>
            </CreateFindingDialog>
          )}
        </div>
      </PageHeader>

      <div className="flex items-center gap-4">
        <Select
          value={kindFilter ?? "ALL"}
          onValueChange={handleKindFilterChange}
        >
          <Option value="ALL">{__("All kinds")}</Option>
          <Option value="MINOR_NONCONFORMITY">{__("Minor nonconformity")}</Option>
          <Option value="MAJOR_NONCONFORMITY">{__("Major nonconformity")}</Option>
          <Option value="OBSERVATION">{__("Observation")}</Option>
          <Option value="EXCEPTION">{__("Exception")}</Option>
        </Select>
        <Select
          value={statusFilter ?? "ALL"}
          onValueChange={handleStatusFilterChange}
        >
          <Option value="ALL">{__("All statuses")}</Option>
          {getStatusOptions(__).map(opt => (
            <Option key={opt.value} value={opt.value}>
              {opt.label}
            </Option>
          ))}
        </Select>
        <Select
          value={priorityFilter ?? "ALL"}
          onValueChange={handlePriorityFilterChange}
        >
          <Option value="ALL">{__("All priorities")}</Option>
          <Option value="LOW">{__("Low")}</Option>
          <Option value="MEDIUM">{__("Medium")}</Option>
          <Option value="HIGH">{__("High")}</Option>
        </Select>
        <Suspense fallback={<Select loading placeholder={__("Loading...")} />}>
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
                      <Th>{__("Kind")}</Th>
                      <Th>{__("Reference ID")}</Th>
                      <Th>{__("Description")}</Th>
                      <Th>{__("Status")}</Th>
                      <Th>{__("Priority")}</Th>
                      <Th>{__("Owner")}</Th>
                      <Th>{__("Due Date")}</Th>
                      {hasAnyAction && <Th>{__("Actions")}</Th>}
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
                      {isLoadingNext ? __("Loading...") : __("Load more")}
                    </Button>
                  </div>
                )}
              </Card>
            )
          : (
              <Card padded>
                <div className="text-center py-12">
                  <h3 className="text-lg font-semibold mb-2">
                    {__("No findings yet")}
                  </h3>
                  <p className="text-txt-tertiary mb-4">
                    {__("Create your first finding to get started.")}
                  </p>
                </div>
              </Card>
            )}
      </div>
    </div>
  );
}

function getKindLabel(kind: string, __: (s: string) => string): string {
  switch (kind) {
    case "MINOR_NONCONFORMITY":
      return __("Minor nonconformity");
    case "MAJOR_NONCONFORMITY":
      return __("Major nonconformity");
    case "OBSERVATION":
      return __("Observation");
    case "EXCEPTION":
      return __("Exception");
    default:
      return kind;
  }
}

type FindingRowProps = {
  findingKey: FindingsPageRowFragment$key;
  connectionId: string;
  hasAnyAction: boolean;
};

function FindingRow(props: FindingRowProps) {
  const finding = useFragment(findingRowFragment, props.findingKey);
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
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
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete finding"),
                    error as GraphQLError[],
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Finding deleted successfully"),
                  variant: "success",
                });
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete finding"),
                  error as GraphQLError,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the finding %s. This action cannot be undone.",
          ),
          finding.referenceId,
        ),
      },
    );
  };

  const detailsUrl = `/organizations/${organizationId}/findings/${finding.id}`;

  return (
    <Tr to={detailsUrl}>
      <Td>
        <Badge variant="neutral">
          {getKindLabel(finding.kind, __)}
        </Badge>
      </Td>
      <Td>
        <span className="font-mono text-sm">{finding.referenceId}</span>
      </Td>
      <Td>
        <div className="min-w-0">
          <p className="whitespace-pre-wrap break-words">
            {finding.description || __("No description")}
          </p>
        </div>
      </Td>
      <Td>
        <Badge variant={getStatusVariant(finding.status)}>
          {getStatusLabel(finding.status)}
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
            ? __("High")
            : finding.priority === "MEDIUM"
              ? __("Medium")
              : __("Low")}
        </Badge>
      </Td>
      <Td>{finding.owner?.fullName || "-"}</Td>
      <Td>
        {finding.dueDate
          ? (
              <time dateTime={finding.dueDate}>
                {formatDate(finding.dueDate)}
              </time>
            )
          : (
              <span className="text-txt-tertiary">{__("No due date")}</span>
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
                {__("Delete")}
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
  const { __ } = useTranslate();
  const people = usePeople(organizationId, { excludeContractEnded: true });

  return (
    <Select value={value ?? "ALL"} onValueChange={onChange}>
      <Option value="ALL">{__("All owners")}</Option>
      {people.map(p => (
        <Option key={p.id} value={p.id}>
          <Avatar name={p.fullName} />
          {p.fullName}
        </Option>
      ))}
    </Select>
  );
}
