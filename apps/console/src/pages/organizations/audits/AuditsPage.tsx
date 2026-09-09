// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  IconPlusLarge,
  IconUpload,
  PageHeader,
  Tbody,
  Th,
  Thead,
  Tr,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useCallback, useEffect, useRef, useState } from "react";
import { useDropzone } from "react-dropzone";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";

import type { AuditGraphListQuery } from "#/__generated__/core/AuditGraphListQuery.graphql";
import type {
  AuditsPageFragment$key,
} from "#/__generated__/core/AuditsPageFragment.graphql";
import { SortableTable } from "#/components/SortableTable";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { auditsQuery } from "../../../hooks/graph/AuditGraph";

import { AuditListItem } from "./_components/AuditListItem";
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
          canUpdate: permission(action: "core:audit:update")
          canDelete: permission(action: "core:audit:delete")
          ...AuditListItem_audit
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<AuditGraphListQuery>;
};

export default function AuditsPage(props: Props) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  const data = usePreloadedQuery<AuditGraphListQuery>(auditsQuery, props.queryRef);
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    paginatedAuditsFragment,
    data.node as AuditsPageFragment$key,
  );
  const audits = pagination.data.audits?.edges?.map(edge => edge.node) ?? [];
  const connectionId = pagination.data.audits.__id;

  usePageTitle(t("auditsPage.title"));

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
          title: t("auditsPage.errors.unsupportedFileType.title"),
          description: t("auditsPage.errors.unsupportedFileType.description"),
          variant: "error",
        });
        return;
      }
      if (!canCreateAudit || acceptedFiles.length === 0) return;
      setDroppedFile(acceptedFiles[0]);
      dropDialogRef.current?.open();
    },
    [canCreateAudit, dropDialogRef, toast, t],
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
            {t("auditsPage.dropzoneOverlay")}
          </p>
        </div>
      )}
      <PageHeader
        title={t("auditsPage.title")}
        description={t("auditsPage.description")}
      >
        {canCreateAudit && (
          <CreateAuditDialog
            connection={connectionId}
            organizationId={organizationId}
          >
            <Button icon={IconPlusLarge}>
              {t("auditsPage.actions.addAudit")}
            </Button>
          </CreateAuditDialog>
        )}
      </PageHeader>
      <SortableTable {...pagination} pageSize={10}>
        <Thead>
          <Tr>
            <Th>{t("auditsPage.columns.name")}</Th>
            <Th>{t("auditsPage.columns.framework")}</Th>
            <Th>{t("auditsPage.columns.state")}</Th>
            <Th>{t("auditsPage.columns.auditPeriod")}</Th>
            <Th>{t("auditsPage.columns.validityPeriod")}</Th>
            <Th>{t("auditsPage.columns.report")}</Th>
            {hasAnyAction && <Th></Th>}
          </Tr>
        </Thead>
        <Tbody>
          {audits.map(entry => (
            <AuditListItem
              key={entry.id}
              auditKey={entry}
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
