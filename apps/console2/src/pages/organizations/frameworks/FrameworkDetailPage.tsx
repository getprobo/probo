import {
  useFragment,
  useMutation,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { usePageTitle } from "@probo/hooks";
import { ConnectionHandler, graphql } from "relay-runtime";
import {
  ActionDropdown,
  Button,
  ControlItem,
  DropdownItem,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
  useConfirm,
} from "@probo/ui";
import { FrameworkLogo } from "/components/FrameworkLogo";
import {
  connectionListKey,
  frameworkNodeQuery,
  useDeleteFrameworkMutation,
} from "/hooks/graph/FrameworkGraph";
import { useTranslate } from "@probo/i18n";
import { LinkedMeasuresCard } from "/components/measures/LinkedMeasuresCard";
import { useNavigate, useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { FrameworkGraphNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphNodeQuery.graphql";
import type {
  FrameworkDetailPageFragment$data,
  FrameworkDetailPageFragment$key,
} from "./__generated__/FrameworkDetailPageFragment.graphql";
import { LinkedDocumentsCard } from "/components/documents/LinkedDocumentsCard";
import { FrameworkFormDialog } from "./dialogs/FrameworkFormDialog";
import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";
import type { FrameworkControlDialogFragment$key } from "./dialogs/__generated__/FrameworkControlDialogFragment.graphql";
import type { NodeOf } from "/types";
import { promisifyMutation } from "@probo/helpers";

const frameworkDetailFragment = graphql`
  fragment FrameworkDetailPageFragment on Framework {
    id
    name
    description
    controls(first: 100) {
      __id
      edges {
        node {
          id
          sectionTitle
          name
          description
          ...FrameworkControlDialogFragment
          measures(first: 100)
            @connection(key: "FrameworkDetailPage_measures") {
            __id
            edges {
              node {
                id
                ...LinkedMeasuresCardFragment
              }
            }
          }
          documents(first: 100)
            @connection(key: "FrameworkDetailPage_documents") {
            __id
            edges {
              node {
                id
                ...LinkedDocumentsCardFragment
              }
            }
          }
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation FrameworkDetailPageAttachMutation(
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
  mutation FrameworkDetailPageDetachMutation(
    $input: DeleteControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation FrameworkDetailPageAttachDocumentMutation(
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
  mutation FrameworkDetailPageDetachDocumentMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphNodeQuery>;
};

export default function FrameworkDetailPage(props: Props) {
  const { __ } = useTranslate();
  const { controlId } = useParams<{ controlId?: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(frameworkNodeQuery, props.queryRef);
  const framework = useFragment<FrameworkDetailPageFragment$key>(
    frameworkDetailFragment,
    data.node
  );
  const navigate = useNavigate();
  const controls = framework.controls.edges.map((edge) => edge.node);
  const selectedControl = controlId
    ? controls.find((control) => control.id === controlId)
    : controls[0];
  const connectionId = framework.controls.__id;
  const deleteFramework = useDeleteFrameworkMutation(
    framework,
    ConnectionHandler.getConnectionID(organizationId, connectionListKey)!
  );

  usePageTitle(`${framework.name} | ${selectedControl?.sectionTitle}`);
  const onDelete = () => {
    deleteFramework({
      onSuccess: () => {
        navigate(`/organizations/${organizationId}/frameworks`);
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <>
            <FrameworkLogo {...framework} />
            {framework.name}
          </>
        }
      >
        <FrameworkFormDialog
          organizationId={organizationId}
          framework={framework}
        >
          <Button icon={IconPencil} variant="secondary">
            {__("Edit")}
          </Button>
        </FrameworkFormDialog>
        <ActionDropdown variant="secondary">
          <DropdownItem icon={IconTrashCan} variant="danger" onClick={onDelete}>
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </PageHeader>
      <div className="text-lg font-semibold">
        {__("Requirement categories")}
      </div>
      <div className="divide-x divide-border-low grid grid-cols-[264px_1fr]">
        <div
          className="space-y-1 overflow-y-auto pr-6 mr-6 sticky top-0"
          style={{ maxHeight: "calc(100vh - 48px)" }}
        >
          {controls.map((control) => (
            <ControlItem
              key={control.id}
              id={control.sectionTitle}
              description={control.name ?? control.description}
              to={`/organizations/${organizationId}/frameworks/${framework.id}/controls/${control.id}`}
              active={selectedControl?.id === control.id}
            />
          ))}
          <FrameworkControlDialog
            frameworkId={framework.id}
            connectionId={connectionId}
          >
            <button className="flex gap-[6px] flex-col w-full p-4 space-y-[6px] rounded-xl cursor-pointer text-start text-sm text-txt-tertiary hover:bg-tertiary-hover">
              <IconPlusLarge size={20} className="text-txt-primary" />
              {__("Add new control")}
            </button>
          </FrameworkControlDialog>
        </div>
        {selectedControl ? (
          <ControlContent
            control={selectedControl}
            frameworkId={framework.id}
            connectionId={connectionId}
          />
        ) : (
          <div className="h-full flex items-center justify-center">
            <div className="text-sm text-txt-secondary">
              {__("No control selected")}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

const deleteControlMutation = graphql`
  mutation FrameworkDetailPageDeleteControlMutation(
    $input: DeleteControlInput!
    $connections: [ID!]!
  ) {
    deleteControl(input: $input) {
      deletedControlId @deleteEdge(connections: $connections)
    }
  }
`;

/**
 * Right side pannel (showing the content of the selected control)
 */
function ControlContent({
  control,
  frameworkId,
  connectionId,
}: {
  control: NodeOf<FrameworkDetailPageFragment$data["controls"]> &
    FrameworkControlDialogFragment$key;
  frameworkId: string;
  connectionId: string;
}) {
  const { __ } = useTranslate();
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
              `/organizations/${organizationId}/frameworks/${frameworkId}`
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
            frameworkId={frameworkId}
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
        measures={control?.measures.edges.map((edge) => edge.node) ?? []}
        params={{ controlId: control.id }}
        connectionId={control.measures.__id!}
        onAttach={attachMeasure}
        onDetach={detachMeasure}
        disabled={isAttachingMeasure || isDetachingMeasure}
      />
      <LinkedDocumentsCard
        variant="card"
        documents={control?.documents.edges.map((edge) => edge.node) ?? []}
        params={{ controlId: control.id }}
        connectionId={control.documents.__id!}
        onAttach={attachDocument}
        onDetach={detachDocument}
        disabled={isAttachingDocument || isDetachingDocument}
      />
    </div>
  );
}
