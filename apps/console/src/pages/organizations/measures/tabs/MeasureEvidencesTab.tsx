import { useNavigate, useOutletContext, useParams } from "react-router";
import type { MeasureEvidencesTabFragment$key } from "/__generated__/core/MeasureEvidencesTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
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
import { graphql } from "relay-runtime";
import {
  useFragment,
  usePaginationFragment,
  useRelayEnvironment,
} from "react-relay";
import { SortableTable } from "/components/SortableTable";
import type { MeasureEvidencesTabFragment_evidence$key } from "/__generated__/core/MeasureEvidencesTabFragment_evidence.graphql";
import { fileSize, fileType, sprintf, formatDate } from "@probo/helpers";
import { EvidencePreviewDialog } from "../dialog/EvidencePreviewDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateEvidenceDialog } from "../dialog/CreateEvidenceDialog";
import { use, useState } from "react";
import { EvidenceDownloadDialog } from "../dialog/EvidenceDownloadDialog";
import { updateStoreCounter } from "/hooks/useMutationWithIncrement";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { PermissionsContext } from "/providers/PermissionsContext";

export const evidencesFragment = graphql`
  fragment MeasureEvidencesTabFragment on Measure
  @refetchable(queryName: "MeasureEvidencesTabQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "EvidenceOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    evidences(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MeasureEvidencesTabFragment_evidences") {
      __id
      edges {
        node {
          id
          file {
            fileName
            mimeType
            size
          }
          ...MeasureEvidencesTabFragment_evidence
        }
      }
    }
  }
`;

export const evidenceFragment = graphql`
  fragment MeasureEvidencesTabFragment_evidence on Evidence {
    id
    file {
      fileName
      mimeType
      size
    }
    type
    createdAt
  }
`;

const deleteEvidenceMutation = graphql`
  mutation MeasureEvidencesTabDeleteMutation(
    $input: DeleteEvidenceInput!
    $connections: [ID!]!
  ) {
    deleteEvidence(input: $input) {
      deletedEvidenceId @deleteEdge(connections: $connections)
    }
  }
`;

export default function MeasureEvidencesTab() {
  const { measure } = useOutletContext<{
    measure: MeasureEvidencesTabFragment$key & { id: string; name: string };
  }>();
  const { evidenceId, snapshotId } = useParams<{
    evidenceId: string;
    snapshotId?: string;
  }>();
  const pagination = usePaginationFragment(evidencesFragment, measure);
  const connectionId = pagination.data.evidences.__id;
  const evidences =
    pagination.data.evidences?.edges?.map((edge) => edge.node) ?? [];
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const evidence = evidences.find((e) => e.id === evidenceId);
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const isSnapshotMode = Boolean(snapshotId);
  const { isAuthorized } = use(PermissionsContext);

  const canAddEvidence = isAuthorized("Measure", "uploadMeasureEvidence");
  const canDeleteEvidence = isAuthorized("Evidence", "deleteEvidence");

  usePageTitle(measure.name + " - " + __("Evidences"));

  return (
    <div className="space-y-6">
      <SortableTable {...pagination}>
        <Thead>
          <Tr>
            <Th>{__("Evidence name")}</Th>
            <Th>{__("Type")}</Th>
            <Th>{__("File size")}</Th>
            <Th>{__("Created at")}</Th>
            <Th width={50}></Th>
          </Tr>
        </Thead>
        <Tbody>
          {evidences.map((evidence) => (
            <EvidenceRow
              key={evidence.id}
              evidenceKey={evidence}
              measureId={measure.id}
              organizationId={organizationId}
              connectionId={connectionId}
              hideActions={isSnapshotMode}
              canDelete={canDeleteEvidence}
              snapshotId={snapshotId}
            />
          ))}
          {!isSnapshotMode && canAddEvidence && (
            <TrButton
              colspan={5}
              onClick={() => dialogRef.current?.open()}
              icon={IconPlusLarge}
            >
              {__("Add evidence")}
            </TrButton>
          )}
        </Tbody>
      </SortableTable>
      {evidence && (
        <EvidencePreviewDialog
          key={evidence?.id}
          onClose={() => {
            const baseUrl = isSnapshotMode
              ? `/organizations/${organizationId}/snapshots/${snapshotId}/risks/measures/${measure.id}/evidences`
              : `/organizations/${organizationId}/measures/${measure.id}/evidences`;
            navigate(baseUrl);
          }}
          evidenceId={evidence.id}
          filename={evidence.file?.fileName || ""}
        />
      )}
      {!isSnapshotMode && canAddEvidence && (
        <CreateEvidenceDialog
          ref={dialogRef}
          measureId={measure.id}
          connectionId={connectionId}
        />
      )}
    </div>
  );
}

function EvidenceRow(props: {
  evidenceKey: MeasureEvidencesTabFragment_evidence$key;
  measureId: string;
  organizationId: string;
  connectionId: string;
  hideActions?: boolean;
  canDelete?: boolean;
  snapshotId?: string;
}) {
  const evidence = useFragment(evidenceFragment, props.evidenceKey);
  const { __ } = useTranslate();

  const [mutateWithToasts, isDeleting] = useMutationWithToasts(
    deleteEvidenceMutation,
    {
      successMessage: sprintf(
        __('Evidence "%s" has been deleted successfully'),
        evidence.file?.fileName || __("Link evidence"),
      ),
      errorMessage: __("Failed to delete evidence"),
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
                props.measureId,
                "evidences(first:0)",
                -1,
              );
            }
          },
        });
      },
      {
        message: sprintf(
          __(
            'This will permanently delete the evidence "%s". This action cannot be undone.',
          ),
          evidence.file?.fileName || __("Link evidence"),
        ),
      },
    );
  };

  const evidenceUrl = props.snapshotId
    ? `/organizations/${props.organizationId}/snapshots/${props.snapshotId}/risks/measures/${props.measureId}/evidences/${evidence.id}`
    : `/organizations/${props.organizationId}/measures/${props.measureId}/evidences/${evidence.id}`;

  return (
    <>
      {isDownloading && (
        <EvidenceDownloadDialog
          evidenceId={evidence.id}
          onClose={() => setIsDownloading(false)}
        />
      )}
      <Tr to={evidenceUrl}>
        <Td>{evidence.file?.fileName}</Td>
        <Td>
          {fileType(__, {
            type: evidence.type,
            mimeType: evidence.file?.mimeType || "",
          })}
        </Td>
        <Td>{fileSize(__, evidence.file?.size || 0)}</Td>
        <Td>{formatDate(evidence.createdAt)}</Td>
        <Td noLink>
          {!props.hideActions && (
            <div className="flex gap-2">
              <ActionDropdown>
                <DropdownItem onClick={() => setIsDownloading(true)}>
                  <IconArrowInbox size={16} />
                  {__("Download")}
                </DropdownItem>
                {props.canDelete && (
                  <DropdownItem
                    variant="danger"
                    icon={IconTrashCan}
                    onClick={handleDelete}
                    disabled={isDeleting}
                  >
                    {__("Delete")}
                  </DropdownItem>
                )}
              </ActionDropdown>
            </div>
          )}
        </Td>
      </Tr>
    </>
  );
}
