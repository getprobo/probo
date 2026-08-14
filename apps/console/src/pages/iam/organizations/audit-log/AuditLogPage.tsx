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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
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
  PageHeader,
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
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditLogPageExportMutation } from "#/__generated__/iam/AuditLogPageExportMutation.graphql";
import type { AuditLogPageFragment$key } from "#/__generated__/iam/AuditLogPageFragment.graphql";
import type { AuditLogPageQuery } from "#/__generated__/iam/AuditLogPageQuery.graphql";
import type { AuditLogPageRefetchQuery } from "#/__generated__/iam/AuditLogPageRefetchQuery.graphql";
import type { AuditLogPageRowFragment$key } from "#/__generated__/iam/AuditLogPageRowFragment.graphql";

export const auditLogPageQuery = graphql`
  query AuditLogPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) @required(action: THROW) {
      __typename
      ... on Organization {
        id
        canExportAuditLog: permission(action: "iam:audit-log:export")
        ...AuditLogPageFragment
      }
    }
  }
`;

const auditLogPageFragment = graphql`
  fragment AuditLogPageFragment on Organization
  @refetchable(queryName: "AuditLogPageRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey" }
  ) {
    auditLogEntries(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: DESC }
    ) @connection(key: "AuditLogPage_auditLogEntries") {
      edges {
        node {
          id
          ...AuditLogPageRowFragment
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
  fragment AuditLogPageRowFragment on AuditLogEntry {
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
  mutation AuditLogPageExportMutation(
    $input: RequestAuditLogExportInput!
  ) {
    requestAuditLogExport(input: $input) {
      exportJobId
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
  language,
}: {
  entryKey: AuditLogPageRowFragment$key;
  language: string;
}) {
  const entry = useFragment(auditLogEntryRowFragment, entryKey);

  return (
    <Tr>
      <Td>
        <span className="text-sm text-txt-secondary whitespace-nowrap">
          {dateFormat(language, entry.createdAt)}
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
  const { t } = useTranslation();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const [commitExport, isExporting] = useMutation<AuditLogPageExportMutation>(exportMutation);

  const handleExport = () => {
    if (!fromDate || !toDate) return;

    commitExport({
      variables: {
        input: {
          organizationId,
          fromTime: new Date(`${fromDate}T00:00:00Z`).toISOString(),
          toTime: new Date(Date.parse(`${toDate}T00:00:00Z`) + 24 * 60 * 60 * 1000).toISOString(),
        },
      },
      onCompleted: (_response, errors) => {
        if (errors) {
          toast({
            title: t("auditLogPage.export.errors.title"),
            description: formatError(t("auditLogPage.export.errors.request"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("auditLogPage.export.messages.successTitle"),
          description: t("auditLogPage.export.messages.success"),
          variant: "success",
        });
        dialogRef.current?.close();
        setFromDate("");
        setToDate("");
      },
      onError: (error) => {
        toast({
          title: t("auditLogPage.export.errors.title"),
          description: formatError(t("auditLogPage.export.errors.request"), error),
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
        {t("auditLogPage.export.actions.export")}
      </Button>
      <Dialog
        className="max-w-md"
        ref={dialogRef}
        title={t("auditLogPage.export.title")}
      >
        <DialogContent className="space-y-4" padded>
          <p className="text-sm text-txt-secondary">
            {t("auditLogPage.export.description")}
          </p>
          <Field label={t("auditLogPage.export.fields.from")}>
            <Input
              type="date"
              value={fromDate}
              onChange={e => setFromDate(e.target.value)}
              required
            />
          </Field>
          <Field label={t("auditLogPage.export.fields.to")}>
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
            disabled={isExporting || !fromDate || !toDate || fromDate > toDate}
          >
            {isExporting
              ? (
                  <>
                    <Spinner size={16} />
                    {t("auditLogPage.export.actions.exporting")}
                  </>
                )
              : t("auditLogPage.export.actions.export")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
  );
}

export function AuditLogPage(props: {
  queryRef: PreloadedQuery<AuditLogPageQuery>;
}) {
  const { t, i18n } = useTranslation();
  usePageTitle(t("auditLogPage.title"));

  const { organization } = usePreloadedQuery<AuditLogPageQuery>(
    auditLogPageQuery,
    props.queryRef,
  );
  if (organization.__typename === "%other") {
    throw new Error("Relay node is not an organization");
  }

  const { data, loadNext, hasNext, isLoadingNext }
    = usePaginationFragment<
      AuditLogPageRefetchQuery,
      AuditLogPageFragment$key
    >(auditLogPageFragment, organization);

  const entries = data?.auditLogEntries?.edges?.map(e => e.node) ?? [];
  const totalCount = data?.auditLogEntries?.totalCount ?? 0;

  return (
    <div className="space-y-6">
      <PageHeader title={t("auditLogPage.title")}>
        {organization.canExportAuditLog && (
          <ExportAuditLogDialog organizationId={organization.id} />
        )}
      </PageHeader>
      <p className="text-sm text-txt-tertiary">
        {t("auditLogPage.description")}
      </p>

      {entries.length === 0
        ? (
            <div className="text-center py-8">
              <p className="text-sm text-txt-tertiary">
                {t("auditLogPage.empty")}
              </p>
            </div>
          )
        : (
            <div className="space-y-4">
              <p className="text-sm text-txt-tertiary">
                {t("auditLogPage.showing", { shown: entries.length, total: totalCount })}
              </p>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{t("auditLogPage.columns.date")}</Th>
                    <Th>{t("auditLogPage.columns.actor")}</Th>
                    <Th>{t("auditLogPage.columns.action")}</Th>
                    <Th>{t("auditLogPage.columns.resource")}</Th>
                  </Tr>
                </Thead>
                <Tbody>
                  {entries.map(entry => (
                    <AuditLogEntryRow key={entry.id} entryKey={entry} language={i18n.language} />
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
                  {t("auditLogPage.actions.showMore")}
                </Button>
              )}
            </div>
          )}
    </div>
  );
}
