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
import { dateFormat, fileSize } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconArrowInbox,
  IconPlusLarge,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  TrButton,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  useFragment,
  usePaginationFragment,
  useRelayEnvironment,
} from "react-relay";
import { useNavigate, useOutletContext, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { InternalControlEvidencesTabFragment$key } from "#/__generated__/core/InternalControlEvidencesTabFragment.graphql";
import type { InternalControlEvidencesTabFragment_evidence$key } from "#/__generated__/core/InternalControlEvidencesTabFragment_evidence.graphql";
import { SortableTable } from "#/components/SortableTable";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { CreateEvidenceDialog } from "../dialog/CreateEvidenceDialog";
import { EvidenceDownloadDialog } from "../dialog/EvidenceDownloadDialog";
import { EvidencePreviewDialog } from "../dialog/EvidencePreviewDialog";

export const evidencesFragment = graphql`
  fragment InternalControlEvidencesTabFragment on InternalControl
  @refetchable(queryName: "InternalControlEvidencesTabQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "EvidenceOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    name
    canUploadEvidence: permission(action: "core:internal-control:upload-evidence")
    evidences(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "InternalControlEvidencesTabFragment_evidences") {
      __id
      edges {
        node {
          id
          file {
            fileName
            mimeType
            size
          }
          ...InternalControlEvidencesTabFragment_evidence
        }
      }
    }
  }
`;

export const evidenceFragment = graphql`
  fragment InternalControlEvidencesTabFragment_evidence on Evidence {
    id
    file {
      fileName
      mimeType
      size
    }
    description
    createdAt
    canDelete: permission(action: "core:evidence:delete")
  }
`;

const deleteEvidenceMutation = graphql`
  mutation InternalControlEvidencesTabDeleteMutation(
    $input: DeleteEvidenceInput!
    $connections: [ID!]!
  ) {
    deleteEvidence(input: $input) {
      deletedEvidenceId @deleteEdge(connections: $connections)
    }
  }
`;

export default function InternalControlEvidencesTab() {
  const { internalControl } = useOutletContext<{
    internalControl: InternalControlEvidencesTabFragment$key;
  }>();
  const { internalControlId, evidenceId } = useParams<{
    internalControlId: string;
    evidenceId: string;
  }>();
  if (!internalControlId) {
    throw new Error("Missing :internalControlId param in route");
  }
  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(evidencesFragment, internalControl);
  const connectionId = pagination.data.evidences.__id;
  const evidences
    = pagination.data.evidences?.edges?.map(edge => edge.node) ?? [];
  const navigate = useNavigate();
  const { t } = useTranslation();
  const evidence = evidences.find(e => e.id === evidenceId);
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  usePageTitle(t("internalControlEvidencesTab.pageTitle", { name: pagination.data.name }));

  return (
    <div className="space-y-6">
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <Th>{t("internalControlEvidencesTab.columns.name")}</Th>
            <Th>{t("internalControlEvidencesTab.columns.description")}</Th>
            <Th>{t("internalControlEvidencesTab.columns.fileType")}</Th>
            <Th>{t("internalControlEvidencesTab.columns.fileSize")}</Th>
            <Th>{t("internalControlEvidencesTab.columns.createdAt")}</Th>
            <Th width={50}></Th>
          </Tr>
        </Thead>
        <Tbody>
          {evidences.map(evidence => (
            <EvidenceRow
              key={evidence.id}
              evidenceKey={evidence}
              internalControlId={internalControlId}
              organizationId={organizationId}
              connectionId={connectionId}
            />
          ))}
          {pagination.data.canUploadEvidence && (
            <TrButton
              colspan={6}
              onClick={() => dialogRef.current?.open()}
              icon={IconPlusLarge}
            >
              {t("internalControlEvidencesTab.actions.add")}
            </TrButton>
          )}
        </Tbody>
      </SortableTable>
      {evidence && (
        <EvidencePreviewDialog
          key={evidence.id}
          onClose={() => {
            void navigate(`/organizations/${organizationId}/governance/internal-controls/${internalControlId}/evidences`);
          }}
          evidenceId={evidence.id}
          filename={evidence.file?.fileName || ""}
        />
      )}
      {pagination.data.canUploadEvidence && (
        <CreateEvidenceDialog
          ref={dialogRef}
          internalControlId={internalControlId}
          connectionId={connectionId}
        />
      )}
    </div>
  );
}

function EvidenceRow(props: {
  evidenceKey: InternalControlEvidencesTabFragment_evidence$key;
  internalControlId: string;
  organizationId: string;
  connectionId: string;
}) {
  const evidence = useFragment(evidenceFragment, props.evidenceKey);
  const { i18n, t } = useTranslation();

  const [mutateWithToasts, isDeleting] = useMutationWithToasts(
    deleteEvidenceMutation,
    {
      successMessage: t("internalControlEvidencesTab.messages.deleted", {
        name: evidence.file?.fileName || t("internalControlEvidencesTab.linkEvidence"),
      }),
      errorMessage: t("internalControlEvidencesTab.errors.delete"),
    },
  );
  const confirm = useConfirm();
  const [isDownloading, setIsDownloading] = useState(false);
  const relayEnv = useRelayEnvironment();

  const handleDelete = () => {
    confirm(
      () => {
        return mutateWithToasts({
          variables: {
            connections: [props.connectionId],
            input: {
              evidenceId: evidence.id,
            },
          },
          onCompleted: (_response, errors) => {
            if (!errors) {
              updateStoreCounter(
                relayEnv,
                props.internalControlId,
                "evidences(first:0)",
                -1,
              );
            }
          },
        });
      },
      {
        message: t("internalControlEvidencesTab.deleteConfirmation", {
          name: evidence.file?.fileName || t("internalControlEvidencesTab.linkEvidence"),
        }),
      },
    );
  };

  const evidenceUrl = `/organizations/${props.organizationId}/governance/internal-controls/${props.internalControlId}/evidences/${evidence.id}`;

  return (
    <>
      {isDownloading && (
        <EvidenceDownloadDialog
          evidenceId={evidence.id}
          onClose={() => setIsDownloading(false)}
        />
      )}
      <Tr to={evidenceUrl}>
        <Td>{evidence.file?.fileName || "—"}</Td>
        <Td>
          <span className="text-txt-secondary text-sm line-clamp-2">
            {evidence.description || "—"}
          </span>
        </Td>
        <Td>{evidence.file?.mimeType || "—"}</Td>
        <Td>{fileSize(evidence.file?.size || 0, t)}</Td>
        <Td>{dateFormat(i18n.language, evidence.createdAt)}</Td>
        <Td noLink>
          <div className="flex gap-2">
            <ActionDropdown>
              <DropdownItem onClick={() => setIsDownloading(true)}>
                <IconArrowInbox size={16} />
                {t("internalControlEvidencesTab.actions.download")}
              </DropdownItem>
              {evidence.canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={handleDelete}
                  disabled={isDeleting}
                >
                  {t("internalControlEvidencesTab.actions.delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </div>
        </Td>
      </Tr>
    </>
  );
}
