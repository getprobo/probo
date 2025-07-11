import { useNavigate, useOutletContext, useParams } from "react-router";
import type { MeasureEvidencesTabFragment$key } from "./__generated__/MeasureEvidencesTabFragment.graphql";
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
import { useFragment, useMutation, usePaginationFragment } from "react-relay";
import { SortableTable } from "/components/SortableTable";
import type { MeasureEvidencesTabFragment_evidence$key } from "./__generated__/MeasureEvidencesTabFragment_evidence.graphql";
import { fileSize, fileType, promisifyMutation, sprintf } from "@probo/helpers";
import { EvidencePreviewDialog } from "../dialog/EvidencePreviewDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { CreateEvidenceDialog } from "../dialog/CreateEvidenceDialog";
import { useState } from "react";
import { EvidenceDownloadDialog } from "../dialog/EvidenceDownloadDialog";

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
          filename
          mimeType
          ...MeasureEvidencesTabFragment_evidence
        }
      }
    }
  }
`;

export const evidenceFragment = graphql`
  fragment MeasureEvidencesTabFragment_evidence on Evidence {
    id
    filename
    size
    type
    createdAt
    mimeType
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
  const { evidenceId } = useParams<{ evidenceId: string }>();
  const pagination = usePaginationFragment(evidencesFragment, measure);
  const connectionId = pagination.data.evidences.__id;
  const evidences =
    pagination.data.evidences?.edges?.map((edge) => edge.node) ?? [];
  const navigate = useNavigate();
  const { __ } = useTranslate();
  const evidence = evidences.find((e) => e.id === evidenceId);
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

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
            />
          ))}
          <TrButton
            colspan={5}
            onClick={() => dialogRef.current?.open()}
            icon={IconPlusLarge}
          >
            {__("Add evidence")}
          </TrButton>
        </Tbody>
      </SortableTable>
      {evidence && (
        <EvidencePreviewDialog
          key={evidence?.id}
          onClose={() =>
            navigate(
              `/organizations/${organizationId}/measures/${measure.id}/evidences`,
            )
          }
          evidenceId={evidence.id}
          filename={evidence.filename}
        />
      )}
      <CreateEvidenceDialog
        ref={dialogRef}
        measureId={measure.id}
        connectionId={connectionId}
      />
    </div>
  );
}

function EvidenceRow(props: {
  evidenceKey: MeasureEvidencesTabFragment_evidence$key;
  measureId: string;
  organizationId: string;
  connectionId: string;
}) {
  const evidence = useFragment(evidenceFragment, props.evidenceKey);
  const { __, dateFormat } = useTranslate();

  const [mutate, isDeleting] = useMutation(deleteEvidenceMutation);
  const confirm = useConfirm();
  const [isDownloading, setIsDownloading] = useState(false);

  const handleDelete = () => {
    confirm(
      () => {
        return promisifyMutation(mutate)({
          variables: {
            connections: [props.connectionId],
            input: {
              evidenceId: evidence.id,
            },
          },
        });
      },
      {
        message: sprintf(
          __(
            'This will permanently delete the evidence "%s". This action cannot be undone.',
          ),
          evidence.filename,
        ),
      },
    );
  };

  return (
    <>
      {isDownloading && (
        <EvidenceDownloadDialog
          evidenceId={evidence.id}
          onClose={() => setIsDownloading(false)}
        />
      )}
      <Tr
        to={`/organizations/${props.organizationId}/measures/${props.measureId}/evidences/${evidence.id}`}
      >
        <Td>{evidence.filename}</Td>
        <Td>{fileType(__, evidence)}</Td>
        <Td>{fileSize(__, evidence.size)}</Td>
        <Td>{dateFormat(evidence.createdAt)}</Td>
        <Td noLink>
          <div className="flex gap-2">
            <ActionDropdown>
              <DropdownItem onClick={() => setIsDownloading(true)}>
                <IconArrowInbox size={16} />
                {__("Download")}
              </DropdownItem>
              <DropdownItem
                variant="danger"
                icon={IconTrashCan}
                onClick={handleDelete}
                disabled={isDeleting}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          </div>
        </Td>
      </Tr>
    </>
  );
}
