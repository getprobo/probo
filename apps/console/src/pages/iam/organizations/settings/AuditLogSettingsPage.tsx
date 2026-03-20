import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconChevronDown,
  Spinner,
  Tbody,
  Td,
  Th,
  Thead,
  Table,
  Tr,
} from "@probo/ui";
import {
  type PreloadedQuery,
  graphql,
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditLogSettingsPageQuery } from "#/__generated__/iam/AuditLogSettingsPageQuery.graphql";
import type { AuditLogSettingsPageFragment$key } from "#/__generated__/iam/AuditLogSettingsPageFragment.graphql";
import type { AuditLogSettingsPageRefetchQuery } from "#/__generated__/iam/AuditLogSettingsPageRefetchQuery.graphql";
import type { AuditLogSettingsPageRowFragment$key } from "#/__generated__/iam/AuditLogSettingsPageRowFragment.graphql";

export const auditLogSettingsPageQuery = graphql`
  query AuditLogSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        ...AuditLogSettingsPageFragment
      }
    }
  }
`;

const auditLogSettingsPageFragment = graphql`
  fragment AuditLogSettingsPageFragment on Organization
  @refetchable(queryName: "AuditLogSettingsPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    auditLogEntries(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "AuditLogSettingsPage_auditLogEntries") {
      edges {
        node {
          id
          ...AuditLogSettingsPageRowFragment
        }
      }
      totalCount
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

const auditLogEntryRowFragment = graphql`
  fragment AuditLogSettingsPageRowFragment on AuditLogEntry {
    id
    actorId
    actorType
    action
    resourceType
    resourceId
    createdAt
  }
`;

function ActorTypeBadge({ type }: { type: string }) {
  switch (type) {
    case "USER":
      return <Badge variant="info" size="sm">{type}</Badge>;
    case "API_KEY":
      return <Badge variant="warning" size="sm">{type}</Badge>;
    case "SYSTEM":
      return <Badge variant="neutral" size="sm">{type}</Badge>;
    default:
      return <Badge size="sm">{type}</Badge>;
  }
}

function ActionBadge({ action }: { action: string }) {
  const parts = action.split(":");
  const verb = parts[parts.length - 1];

  if (
    verb === "create" ||
    verb === "upload" ||
    verb === "import" ||
    verb === "publish"
  ) {
    return <Badge variant="success" size="sm">{action}</Badge>;
  }
  if (verb === "delete" || verb === "archive") {
    return <Badge variant="danger" size="sm">{action}</Badge>;
  }
  if (
    verb === "update" ||
    verb === "assign" ||
    verb === "unassign" ||
    verb === "unarchive"
  ) {
    return <Badge variant="warning" size="sm">{action}</Badge>;
  }
  if (verb === "get" || verb === "list") {
    return <Badge variant="neutral" size="sm">{action}</Badge>;
  }
  return <Badge size="sm">{action}</Badge>;
}

function AuditLogEntryRow({
  entryKey,
}: {
  entryKey: AuditLogSettingsPageRowFragment$key;
}) {
  const entry = useFragment(auditLogEntryRowFragment, entryKey);

  return (
    <Tr>
      <Td>
        <span className="text-sm text-txt-secondary whitespace-nowrap">
          {formatDate(entry.createdAt)}
        </span>
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <ActorTypeBadge type={entry.actorType} />
          <span className="text-sm font-mono text-txt-secondary truncate max-w-48">
            {entry.actorId}
          </span>
        </div>
      </Td>
      <Td>
        <ActionBadge action={entry.action} />
      </Td>
      <Td>
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">
            {entry.resourceType}
          </span>
          <span className="text-sm font-mono text-txt-tertiary truncate max-w-48">
            {entry.resourceId}
          </span>
        </div>
      </Td>
    </Tr>
  );
}

export function AuditLogSettingsPage(props: {
  queryRef: PreloadedQuery<AuditLogSettingsPageQuery>;
}) {
  const { __ } = useTranslate();

  const { organization } = usePreloadedQuery(
    auditLogSettingsPageQuery,
    props.queryRef,
  );
  if (organization.__typename === "%other") {
    throw new Error("Relay node is not an organization");
  }

  const { data, loadNext, hasNext, isLoadingNext } =
    usePaginationFragment<
      AuditLogSettingsPageRefetchQuery,
      AuditLogSettingsPageFragment$key
    >(auditLogSettingsPageFragment, organization);

  const entries = data?.auditLogEntries?.edges?.map((e) => e.node) ?? [];
  const totalCount = data?.auditLogEntries?.totalCount ?? 0;

  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-base font-medium">{__("Audit Log")}</h2>
        <p className="text-sm text-txt-tertiary">
          {__(
            "A record of all actions performed in your organization. Entries are immutable and cannot be modified or deleted.",
          )}
        </p>
      </div>

      {entries.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-sm text-txt-tertiary">
            {__("No audit log entries yet.")}
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          <p className="text-sm text-txt-tertiary">
            {`${__("Showing")} ${entries.length} ${__("of")} ${totalCount} ${__("entries")}`}
          </p>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Date")}</Th>
                <Th>{__("Actor")}</Th>
                <Th>{__("Action")}</Th>
                <Th>{__("Resource")}</Th>
              </Tr>
            </Thead>
            <Tbody>
              {entries.map((entry) => (
                <AuditLogEntryRow key={entry.id} entryKey={entry} />
              ))}
            </Tbody>
          </Table>
          {hasNext && (
            <Button
              variant="tertiary"
              onClick={() => loadNext(50)}
              className="mx-auto"
              disabled={isLoadingNext}
              icon={isLoadingNext ? Spinner : IconChevronDown}
            >
              {__("Show more")}
            </Button>
          )}
        </div>
      )}
    </div>
  );
}
