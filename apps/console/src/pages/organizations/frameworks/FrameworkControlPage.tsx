import {
  useMutation,
  usePreloadedQuery,
  type PreloadedQuery,
  type UseMutationConfig,
} from "react-relay";
import { graphql, type MutationParameters } from "relay-runtime";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  useConfirm,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { formatError, type GraphQLError } from "@probo/helpers";
import { LinkedMeasuresCard } from "/components/measures/LinkedMeasuresCard";
import { useNavigate, useOutletContext } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { LinkedDocumentsCard } from "/components/documents/LinkedDocumentsCard";
import { LinkedAuditsCard } from "/components/audits/LinkedAuditsCard";
import { LinkedSnapshotsCard } from "/components/snapshots/LinkedSnapshotsCard";
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

const attachAuditMutation = graphql`
  mutation FrameworkControlPageAttachAuditMutation(
    $input: CreateControlAuditMappingInput!
    $connections: [ID!]!
  ) {
    createControlAuditMapping(input: $input) {
      auditEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedAuditsCardFragment
        }
      }
    }
  }
`;

const detachAuditMutation = graphql`
  mutation FrameworkControlPageDetachAuditMutation(
    $input: DeleteControlAuditMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlAuditMapping(input: $input) {
      deletedAuditId @deleteEdge(connections: $connections)
    }
  }
`;

const attachSnapshotMutation = graphql`
  mutation FrameworkControlPageAttachSnapshotMutation(
    $input: CreateControlSnapshotMappingInput!
    $connections: [ID!]!
  ) {
    createControlSnapshotMapping(input: $input) {
      snapshotEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedSnapshotsCardFragment
        }
      }
    }
  }
`;

const detachSnapshotMutation = graphql`
  mutation FrameworkControlPageDetachSnapshotMutation(
    $input: DeleteControlSnapshotMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlSnapshotMapping(input: $input) {
      deletedSnapshotId @deleteEdge(connections: $connections)
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
  const { toast } = useToast();
  const { framework } = useOutletContext<{
    framework: FrameworkDetailPageFragment$data;
  }>();
  const connectionId = framework.controls.__id;
  const control = usePreloadedQuery(frameworkControlNodeQuery, queryRef).node;
  const organizationId = useOrganizationId();
  const confirm = useConfirm();
  const navigate = useNavigate();

  const [detachMeasure, isDetachingMeasure] = useMutation(detachMeasureMutation);
  const [attachMeasure, isAttachingMeasure] = useMutation(attachMeasureMutation);
  const [detachDocument, isDetachingDocument] = useMutation(detachDocumentMutation);
  const [attachDocument, isAttachingDocument] = useMutation(attachDocumentMutation);
  const [detachAudit, isDetachingAudit] = useMutation(detachAuditMutation);
  const [attachAudit, isAttachingAudit] = useMutation(attachAuditMutation);
  const [detachSnapshot, isDetachingSnapshot] = useMutation(detachSnapshotMutation);
  const [attachSnapshot, isAttachingSnapshot] = useMutation(attachSnapshotMutation);
  const [deleteControl] = useMutation(deleteControlMutation);

  const withErrorHandling = <T extends MutationParameters>(
    mutationFn: (config: UseMutationConfig<T>) => void,
    errorMessage: string
  ) => (options: UseMutationConfig<T>) => {
    mutationFn({
      ...options,
      onCompleted: (response, error) => {
        if (error) {
          toast({
            title: __("Error"),
            description: formatError(errorMessage, error as GraphQLError),
            variant: "error",
          });
        }
        options.onCompleted?.(response, error);
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(errorMessage, error as GraphQLError),
          variant: "error",
        });
        options.onError?.(error);
      },
    });
  };

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
        <div className="flex items-center gap-3">
          <div className="text-xl font-medium px-[6px] py-[2px] border border-border-low rounded-lg w-max bg-active mb-3">
            {control.sectionTitle}
          </div>
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

      {control.status === "EXCLUDED" && (
        <div className="bg-danger border border-border-danger rounded-lg p-4">
          <div className="flex items-center gap-2 mb-2">
            <div className="font-medium text-txt-danger">
              {__("This control is excluded")}
            </div>
          </div>
          <div className="text-sm">
            <strong>{__("Justification:")}</strong> {control.exclusionJustification || __("No justification provided")}
          </div>
        </div>
      )}
      <div className={control.status === "EXCLUDED" ? "opacity-60" : ""}>
        <div className="text-base mb-4">{control.name}</div>
        <div className="mb-4">
          <LinkedMeasuresCard
            variant="card"
            measures={control.measures?.edges.map((edge) => edge.node) ?? []}
            params={{ controlId: control.id }}
            connectionId={control.measures?.__id!}
            onAttach={withErrorHandling(attachMeasure, __("Failed to link measure"))}
            onDetach={withErrorHandling(detachMeasure, __("Failed to unlink measure"))}
            disabled={isAttachingMeasure || isDetachingMeasure}
          />
        </div>
        <div className="mb-4">
          <LinkedDocumentsCard
            variant="card"
            documents={control.documents?.edges.map((edge) => edge.node) ?? []}
            params={{ controlId: control.id }}
            connectionId={control.documents?.__id!}
            onAttach={withErrorHandling(attachDocument, __("Failed to link document"))}
            onDetach={withErrorHandling(detachDocument, __("Failed to unlink document"))}
            disabled={isAttachingDocument || isDetachingDocument}
          />
        </div>
        <div className="mb-4">
          <LinkedAuditsCard
            variant="card"
            audits={control.audits?.edges.map((edge) => edge.node) ?? []}
            params={{ controlId: control.id }}
            connectionId={control.audits?.__id!}
            onAttach={withErrorHandling(attachAudit, __("Failed to link audit"))}
            onDetach={withErrorHandling(detachAudit, __("Failed to unlink audit"))}
            disabled={isAttachingAudit || isDetachingAudit}
          />
        </div>
        <div className="mb-4">
          <LinkedSnapshotsCard
            variant="card"
            snapshots={control.snapshots?.edges.map((edge) => edge.node) ?? []}
            params={{ controlId: control.id }}
            connectionId={control.snapshots?.__id!}
            onAttach={withErrorHandling(attachSnapshot, __("Failed to link snapshot"))}
            onDetach={withErrorHandling(detachSnapshot, __("Failed to unlink snapshot"))}
            disabled={isAttachingSnapshot || isDetachingSnapshot}
          />
        </div>
      </div>
    </div>
  );
}
