import {
  useMutation,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { graphql } from "relay-runtime";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  useConfirm,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { LinkedMeasuresCard } from "/components/measures/LinkedMeasuresCard";
import { useNavigate, useOutletContext } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedDocumentsCard } from "/components/documents/LinkedDocumentsCard";
import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import { promisifyMutation } from "@probo/helpers";
import type { FrameworkGraphControlNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphControlNodeQuery.graphql";
import { frameworkControlNodeQuery } from "/hooks/graph/FrameworkGraph";
import type { FrameworkDetailPageFragment$data } from "./__generated__/FrameworkDetailPageFragment.graphql";

const attachMeasureMutation = graphql`
  mutation FrameworkControlPageAttachMutation(
    $input: CreateControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createControlMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedMeasuresCardFragment
        }
      }
    }
  }
`;

const detachMeasureMutation = graphql`
  mutation FrameworkControlPageDetachMutation(
    $input: DeleteControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation FrameworkControlPageAttachDocumentMutation(
    $input: CreateControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createControlDocumentMapping(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

const detachDocumentMutation = graphql`
  mutation FrameworkControlPageDetachDocumentMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

const deleteControlMutation = graphql`
  mutation FrameworkControlPageDeleteControlMutation(
    $input: DeleteControlInput!
    $connections: [ID!]!
  ) {
    deleteControl(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphControlNodeQuery>;
};

/**
 * Display the control detail on the right panel
 */
export default function FrameworkControlPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const { framework } = useOutletContext<{
    framework: FrameworkDetailPageFragment$data;
  }>();
  const connectionId = framework.controls.__id;
  const control = usePreloadedQuery(frameworkControlNodeQuery, queryRef).node;
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const navigate = useNavigate();
  // Mutations
  const [detachMeasure, isDetachingMeasure] = useMutation(
    detachMeasureMutation
  );
  const [attachMeasure, isAttachingMeasure] = useMutation(
    attachMeasureMutation
  );
  const [detachDocument, isDetachingDocument] = useMutation(
    detachDocumentMutation
  );
  const [attachDocument, isAttachingDocument] = useMutation(
    attachDocumentMutation
  );
  const [deleteControl] = useMutation(deleteControlMutation);

  const onDelete = () => {
    confirm(
      () => {
        return promisifyMutation(deleteControl)({
          variables: {
            input: {
              controlId: control.id,
            },
            connections: [connectionId],
          },
          onCompleted: () => {
            navigate(
              `/organizations/${organizationId}/frameworks/${framework.id}`
            );
          },
        });
      },
      {
        message: __("Are you sure you want to delete this control?"),
      }
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between">
        <div className="text-xl font-medium px-[6px] py-[2px] border border-border-low rounded-lg w-max bg-active mb-3">
          {control.sectionTitle}
        </div>
        <div className="flex gap-2">
          <FrameworkControlDialog
            frameworkId={framework.id}
            connectionId={connectionId}
            control={control}
          >
            <Button icon={IconPencil} variant="secondary">
              {__("Edit control")}
            </Button>
          </FrameworkControlDialog>
          <ActionDropdown variant="secondary">
            <DropdownItem
              icon={IconTrashCan}
              variant="danger"
              onClick={onDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </div>
      </div>

      <div className="text-base">{control.name}</div>
      <LinkedMeasuresCard
        variant="card"
        measures={control.measures?.edges.map((edge) => edge.node) ?? []}
        params={{ controlId: control.id }}
        connectionId={control.measures?.__id!}
        onAttach={attachMeasure}
        onDetach={detachMeasure}
        disabled={isAttachingMeasure || isDetachingMeasure}
      />
      <LinkedDocumentsCard
        variant="card"
        documents={control.documents?.edges.map((edge) => edge.node) ?? []}
        params={{ controlId: control.id }}
        connectionId={control.documents?.__id!}
        onAttach={attachDocument}
        onDetach={detachDocument}
        disabled={isAttachingDocument || isDetachingDocument}
      />
    </div>
  );
}
