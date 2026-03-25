import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconArrowDown,
  IconChevronDown,
  Input,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditLogSettingsPageExportMutation } from "#/__generated__/iam/AuditLogSettingsPageExportMutation.graphql";
import type { AuditLogSettingsPageFragment$key } from "#/__generated__/iam/AuditLogSettingsPageFragment.graphql";
import type { AuditLogSettingsPageQuery } from "#/__generated__/iam/AuditLogSettingsPageQuery.graphql";
import type { AuditLogSettingsPageRefetchQuery } from "#/__generated__/iam/AuditLogSettingsPageRefetchQuery.graphql";
import type { AuditLogSettingsPageRowFragment$key } from "#/__generated__/iam/AuditLogSettingsPageRowFragment.graphql";

export const auditLogSettingsPageQuery = graphql`
  query AuditLogSettingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canExportAuditLog: permission(action: "iam:audit-log:export")
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

const exportMutation = graphql`
  mutation AuditLogSettingsPageExportMutation(
    $input: RequestAuditLogExportInput!
  ) {
    requestAuditLogExport(input: $input) {
      logExportId
    }
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
    verb === "create"
    || verb === "upload"
    || verb === "import"
    || verb === "publish"
  ) {
    return <Badge variant="success" size="sm">{action}</Badge>;
  }
  if (verb === "delete" || verb === "archive") {
    return <Badge variant="danger" size="sm">{action}</Badge>;
  }
  if (
    verb === "update"
    || verb === "assign"
    || verb === "unassign"
    || verb === "unarchive"
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

function ExportAuditLogDialog({
  organizationId,
}: {
  organizationId: string;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [commitExport, isExporting] = useMutation<AuditLogSettingsPageExportMutation>(exportMutation);

  const handleExport = () => {
    if (!fromDate || !toDate) return;

    commitExport({
      variables: {
        input: {
          organizationId,
          fromTime: new Date(fromDate).toISOString(),
          toTime: new Date(toDate).toISOString(),
        },
      },
      onCompleted: (_response, errors) => {
        if (errors) {
          toast({
            title: __("Error"),
            description: __("Failed to request audit log export."),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Export started. You will receive an email with a download link when it is ready."),
          variant: "success",
        });
        dialogRef.current?.close();
        setFromDate("");
        setToDate("");
      },
      onError: () => {
        toast({
          title: __("Error"),
          description: __("Failed to request audit log export."),
          variant: "error",
        });
      },
    });
  };

  return (
    <>
      <Button
        variant="secondary"
        icon={IconArrowDown}
        onClick={() => dialogRef.current?.open()}
      >
        {__("Export")}
      </Button>
      <Dialog
        className="max-w-md"
        ref={dialogRef}
        title={__("Export Audit Log")}
      >
        <DialogContent className="space-y-4" padded>
          <p className="text-sm text-txt-secondary">
            {__("Select a date range to export audit log entries as JSONL. You will receive an email with a download link.")}
          </p>
          <Field label={__("From")}>
            <Input
              type="date"
              value={fromDate}
              onChange={e => setFromDate(e.target.value)}
              required
            />
          </Field>
          <Field label={__("To")}>
            <Input
              type="date"
              value={toDate}
              onChange={e => setToDate(e.target.value)}
              required
            />
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button
            onClick={handleExport}
            disabled={isExporting || !fromDate || !toDate}
          >
            {isExporting
              ? (
                  <>
                    <Spinner size={16} />
                    {__("Exporting...")}
                  </>
                )
              : __("Export")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
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

  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginationFragment<
      AuditLogSettingsPageRefetchQuery,
      AuditLogSettingsPageFragment$key
    >(auditLogSettingsPageFragment, organization);

  const entries = data?.auditLogEntries?.edges?.map(e => e.node) ?? [];
  const totalCount = data?.auditLogEntries?.totalCount ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-base font-medium">{__("Audit Log")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__(
              "A record of all actions performed in your organization. Entries are immutable and cannot be modified or deleted.",
            )}
          </p>
        </div>
        {organization.canExportAuditLog && (
          <ExportAuditLogDialog organizationId={organization.id} />
        )}
      </div>

      {entries.length === 0
        ? (
            <div className="text-center py-8">
              <p className="text-sm text-txt-tertiary">
                {__("No audit log entries yet.")}
              </p>
            </div>
          )
        : (
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
                  {entries.map(entry => (
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
