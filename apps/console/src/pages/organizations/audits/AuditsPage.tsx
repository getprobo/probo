import {
  formatDate,
  getAuditStateLabel,
  getAuditStateVariant,
} from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Badge,
  Button,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  IconUpload,
  PageHeader,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useCallback, useEffect, useRef } from "react";
import { useDropzone } from "react-dropzone";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditGraphListQuery } from "#/__generated__/core/AuditGraphListQuery.graphql";
import type {
  AuditsPageFragment$data,
  AuditsPageFragment$key,
} from "#/__generated__/core/AuditsPageFragment.graphql";
import { SortableTable } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import type { NodeOf } from "#/types";

import { auditsQuery, useDeleteAudit } from "../../../hooks/graph/AuditGraph";

import { CreateAuditDialog } from "./dialogs/CreateAuditDialog";

const paginatedAuditsFragment = graphql`
  fragment AuditsPageFragment on Organization
  @refetchable(queryName: "AuditsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    orderBy: { type: "AuditOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    audits(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $orderBy
    ) @connection(key: "AuditsPage_audits") {
      __id
      edges {
        node {
          id
          name
          validFrom
          validUntil
          report {
            id
          }
          state
          framework {
            id
            name
          }
          canUpdate: permission(action: "core:audit:update")
          canDelete: permission(action: "core:audit:delete")
        }
      }
    }
  }
`;

type AuditEntry = NodeOf<AuditsPageFragment$data["audits"]>;

type Props = {
  queryRef: PreloadedQuery<AuditGraphListQuery>;
};

export default function AuditsPage(props: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery(auditsQuery, props.queryRef);
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedAuditsFragment,
    data.node as AuditsPageFragment$key,
  );
  const audits = pagination.data.audits?.edges?.map(edge => edge.node) ?? [];
  const connectionId = pagination.data.audits.__id;

  usePageTitle(__("Audits"));

  const hasAnyAction = audits.some(
    audit => audit.canDelete || audit.canUpdate,
  );

  const canCreateAudit = data.node.canCreateAudit;
  const [droppedFile, setDroppedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const dropDialogRef = useDialogRef();
  const dragCounterRef = useRef(0);

  const onDrop = useCallback(
    (acceptedFiles: File[], fileRejections: { file: File }[]) => {
      setIsDragging(false);
      dragCounterRef.current = 0;
      if (fileRejections.length > 0) {
        toast({
          title: __("Unsupported file type"),
          description: __("Only PDF files are supported."),
          variant: "error",
        });
        return;
      }
      if (!canCreateAudit || acceptedFiles.length === 0) return;
      setDroppedFile(acceptedFiles[0]);
      dropDialogRef.current?.open();
    },
    [canCreateAudit, dropDialogRef, toast, __],
  );

  useEffect(() => {
    if (!canCreateAudit) return;

    const handleDragEnter = (e: DragEvent) => {
      e.preventDefault();
      dragCounterRef.current++;
      if (e.dataTransfer?.types.includes("Files")) {
        setIsDragging(true);
      }
    };

    const handleDragLeave = (e: DragEvent) => {
      e.preventDefault();
      dragCounterRef.current = Math.max(0, dragCounterRef.current - 1);
      if (dragCounterRef.current <= 0) {
        setIsDragging(false);
      }
    };

    const handleDragOver = (e: DragEvent) => {
      e.preventDefault();
    };

    const handleDrop = () => {
      setIsDragging(false);
      dragCounterRef.current = 0;
    };

    window.addEventListener("dragenter", handleDragEnter);
    window.addEventListener("dragleave", handleDragLeave);
    window.addEventListener("dragover", handleDragOver);
    window.addEventListener("drop", handleDrop);

    return () => {
      window.removeEventListener("dragenter", handleDragEnter);
      window.removeEventListener("dragleave", handleDragLeave);
      window.removeEventListener("dragover", handleDragOver);
      window.removeEventListener("drop", handleDrop);
    };
  }, [canCreateAudit]);

  const { getRootProps, getInputProps } = useDropzone({
    noClick: true,
    noKeyboard: true,
    accept: { "application/pdf": [".pdf"] },
    multiple: false,
    onDrop,
    disabled: !canCreateAudit,
  });

  const handleDropDialogClose = () => {
    setDroppedFile(null);
  };

  return (
    <div className="space-y-6">
      {isDragging && canCreateAudit && (
        <div
          {...getRootProps()}
          className="border-primary bg-primary/5 pointer-events-auto fixed inset-0 top-12 z-40 flex flex-col items-center justify-center border-2 border-dashed"
        >
          <input {...getInputProps()} />
          <IconUpload className="text-primary mb-2 size-8" />
          <p className="text-primary text-sm font-medium">
            {__("Drop a PDF to create an audit with a report")}
          </p>
        </div>
      )}
      <PageHeader
        title={__("Audits")}
        description={__(
          "Manage your organization's compliance audits and their progress.",
        )}
      >
        {canCreateAudit && (
          <CreateAuditDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>{__("Add audit")}</Button>
          </CreateAuditDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination} pageSize={10}>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Framework")}</Th>
            <Th>{__("State")}</Th>
            <Th>{__("Valid From")}</Th>
            <Th>{__("Valid Until")}</Th>
            <Th>{__("Report")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {audits.map(entry => (
            <AuditRow
              key={entry.id}
              entry={entry}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
            />
          ))}
        </Tbody>
      </SortableTable>
      {canCreateAudit && (
        <CreateAuditDialog
          ref={dropDialogRef}
          connection={connectionId}
          organizationId={organizationId}
          file={droppedFile}
          onClose={handleDropDialogClose}
        />
      )}
    </div>
  );
}

function AuditRow({
  entry,
  connectionId,
  hasAnyAction,
}: {
  entry: AuditEntry;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deleteAudit = useDeleteAudit(entry, connectionId);

  return (
    <Tr to={`/organizations/${organizationId}/audits/${entry.id}`}>
      <Td>{entry.name || __("Untitled")}</Td>
      <Td>{entry.framework?.name ?? __("Unknown Framework")}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(entry.state)}>
          {getAuditStateLabel(__, entry.state)}
        </Badge>
      </Td>
      <Td>{formatDate(entry.validFrom) || __("Not set")}</Td>
      <Td>{formatDate(entry.validUntil) || __("Not set")}</Td>
      <Td>
        {entry.report
          ? (
              <div className="flex flex-col">
                <Badge variant="success">{__("Uploaded")}</Badge>
              </div>
            )
          : (
              <Badge variant="neutral">{__("Not uploaded")}</Badge>
            )}
      </Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            {entry.canDelete && (
              <DropdownItem
                onClick={deleteAudit}
                variant="danger"
                icon={IconTrashCan}
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
