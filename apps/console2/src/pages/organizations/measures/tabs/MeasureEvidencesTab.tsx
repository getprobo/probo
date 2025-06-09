import { useOutletContext } from "react-router";
import type { MeasureEvidencesTabFragment$key } from "./__generated__/MeasureEvidencesTabFragment.graphql";
import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Dropzone,
  IconArrowInbox,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import { useFragment, useMutation, useRefetchableFragment } from "react-relay";
import { SortableTable } from "/components/SortableTable";
import type { MeasureEvidencesTabFragment_evidence$key } from "./__generated__/MeasureEvidencesTabFragment_evidence.graphql";
import { fileSize, fileType } from "@probo/helpers";
import { promisifyMutation, sprintf } from "@probo/helpers";

const evidencesFragment = graphql`
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
          ...MeasureEvidencesTabFragment_evidence
        }
      }
    }
  }
`;

const evidenceFragment = graphql`
  fragment MeasureEvidencesTabFragment_evidence on Evidence {
    id
    filename
    size
    type
    createdAt
    fileUrl
    mimeType
  }
`;

const uploadEvidenceMutation = graphql`
  mutation MeasureEvidencesTabUploadMutation(
    $input: UploadMeasureEvidenceInput!
    $connections: [ID!]!
  ) {
    uploadMeasureEvidence(input: $input) {
      evidenceEdge @appendEdge(connections: $connections) {
        node {
          id
          ...MeasureEvidencesTabFragment_evidence
        }
      }
    }
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
  const [data, refetch] = useRefetchableFragment(evidencesFragment, measure);
  const connectionId = data.evidences.__id;
  const evidences = data.evidences?.edges?.map((edge) => edge.node) ?? [];

  const { __ } = useTranslate();
  const [mutate, isUpdating] = useMutation(uploadEvidenceMutation);

  usePageTitle(measure.name + " - " + __("Evidences"));

  const handleDrop = (files: File[]) => {
    for (const file of files) {
      mutate({
        variables: {
          connections: [connectionId],
          input: {
            measureId: measure.id,
            file: null,
          },
        },
        uploadables: {
          "input.file": file,
        },
      });
    }
  };

  return (
    <div className="space-y-6">
      <Dropzone
        description={__("Only PDF files up to 10MB are allowed")}
        isUploading={isUpdating}
        onDrop={handleDrop}
        accept={{
          "application/pdf": [".pdf"],
        }}
        maxSize={10}
      />
      <SortableTable refetch={refetch}>
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
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function EvidenceRow(props: {
  evidenceKey: MeasureEvidencesTabFragment_evidence$key;
  connectionId: string;
}) {
  const evidence = useFragment(evidenceFragment, props.evidenceKey);
  const { __, dateFormat } = useTranslate();

  const [mutate, isDeleting] = useMutation(deleteEvidenceMutation);
  const confirm = useConfirm();

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
            'This will permanently delete the evidence "%s". This action cannot be undone.'
          ),
          evidence.filename
        ),
      }
    );
  };

  return (
    <Tr>
      <Td>{evidence.filename}</Td>
      <Td>{fileType(__, evidence)}</Td>
      <Td>{fileSize(__, evidence.size)}</Td>
      <Td>{dateFormat(evidence.createdAt)}</Td>
      <Td>
        <div className="flex gap-2">
          <Button asChild variant="secondary">
            <a href={evidence.fileUrl ?? ""} target="_blank">
              <IconArrowInbox size={16} />
              {__("Download")}
            </a>
          </Button>
          <Button
            variant="danger"
            icon={IconTrashCan}
            onClick={handleDelete}
            disabled={isDeleting}
          />
        </div>
      </Td>
    </Tr>
  );
}
