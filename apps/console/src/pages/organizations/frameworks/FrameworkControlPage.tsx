import { formatError, type GraphQLError } from "@probo/helpers";
import { promisifyMutation } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  useConfirm,
  useToast,
} from "@probo/ui";
import {
  type PreloadedQuery,
  useMutation,
  type UseMutationConfig,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate, useOutletContext } from "react-router";
import { graphql, type MutationParameters } from "relay-runtime";

import type { FrameworkDetailPageFragment$data } from "#/__generated__/core/FrameworkDetailPageFragment.graphql";
import type { FrameworkGraphControlNodeQuery } from "#/__generated__/core/FrameworkGraphControlNodeQuery.graphql";
import { LinkedAuditsCard } from "#/components/audits/LinkedAuditsCard";
import { LinkedDocumentsCard } from "#/components/documents/LinkedDocumentsCard";
import { LinkedMeasuresCard } from "#/components/measures/LinkedMeasuresCard";
import { LinkedObligationsCard } from "#/components/obligations/LinkedObligationsCard";
import { LinkedSnapshotsCard } from "#/components/snapshots/LinkedSnapshotsCard";
import { frameworkControlNodeQuery } from "#/hooks/graph/FrameworkGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { FrameworkControlDialog } from "./dialogs/FrameworkControlDialog";

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

const attachObligationMutation = graphql`
  mutation FrameworkControlPageAttachObligationMutation(
      $input: CreateControlObligationMappingInput!
      $connections: [ID!]!
  ) {
      createControlObligationMapping(input: $input) {
          obligationEdge @prependEdge(connections: $connections) {
              node {
                  id
                  ...LinkedObligationsCardFragment
              }
          }
      }
  }
`;

const detachObligationMutation = graphql`
  mutation FrameworkControlPageDetachObligationMutation(
      $input: DeleteControlObligationMappingInput!
      $connections: [ID!]!
  ) {
      deleteControlObligationMapping(input: $input) {
          deletedObligationId @deleteEdge(connections: $connections)
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
  const [detachMeasure, isDetachingMeasure] = useMutation(
    detachMeasureMutation,
  );
  const [attachMeasure, isAttachingMeasure] = useMutation(
    attachMeasureMutation,
  );
  const [detachDocument, isDetachingDocument] = useMutation(
    detachDocumentMutation,
  );
  const [attachDocument, isAttachingDocument] = useMutation(
    attachDocumentMutation,
  );
  const [detachAudit, isDetachingAudit] = useMutation(detachAuditMutation);
  const [attachAudit, isAttachingAudit] = useMutation(attachAuditMutation);
  const [detachSnapshot, isDetachingSnapshot] = useMutation(
    detachSnapshotMutation,
  );
  const [attachSnapshot, isAttachingSnapshot] = useMutation(
    attachSnapshotMutation,
  );
  const [deleteControl] = useMutation(deleteControlMutation);

  const [attachObligation, isAttachingObligation] = useMutation(
    attachObligationMutation,
  );
  const [detachObligation, isDetachingObligation] = useMutation(
    detachObligationMutation,
  );

  const canLinkMeasure = control.canCreateMeasureMapping;
  const canUnlinkMeasure = control.canDeleteMeasureMapping;
  const measuresReadOnly = !canLinkMeasure && !canUnlinkMeasure;

  const canLinkDocument = control.canCreateDocumentMapping;
  const canUnlinkDocument = control.canDeleteDocumentMapping;
  const documentsReadOnly = !canLinkDocument && !canUnlinkDocument;

  const canLinkAudit = control.canCreateAuditMapping;
  const canUnlinkAudit = control.canDeleteAuditMapping;
  const auditsReadOnly = !canLinkAudit && !canUnlinkAudit;

  const canLinkSnapshot = control.canCreateSnapshotMapping;
  const canUnlinkSnapshot = control.canDeleteSnapshotMapping;
  const snapshotsReadOnly = !canLinkSnapshot && !canUnlinkSnapshot;

  const canLinkObligation = control.canCreateObligationMapping;
  const canUnlinkObligation = control.canDeleteObligationMapping;
  const obligationsReadOnly = !canLinkObligation && !canUnlinkObligation;

  const withErrorHandling
    = <T extends MutationParameters>(
      mutationFn: (config: UseMutationConfig<T>) => void,
      errorMessage: string,
    ) =>
      (options: UseMutationConfig<T>) => {
        mutationFn({
          ...options,
          onCompleted: (response, error) => {
            if (error) {
              toast({
                title: __("Error"),
                description: formatError(
                  errorMessage,
                  error as GraphQLError,
                ),
                variant: "error",
              });
            }
            options.onCompleted?.(response, error);
          },
          onError: (error) => {
            toast({
              title: __("Error"),
              description: formatError(
                errorMessage,
                error as GraphQLError,
              ),
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
            void navigate(
              `/organizations/${organizationId}/frameworks/${framework.id}`,
            );
          },
        });
      },
      {
        message: __("Are you sure you want to delete this control?"),
      },
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
          {control.canUpdate && (
            <FrameworkControlDialog
              frameworkId={framework.id}
              connectionId={connectionId}
              control={control}
            >
              <Button icon={IconPencil} variant="secondary">
                {__("Edit control")}
              </Button>
            </FrameworkControlDialog>
          )}
          {control.canDelete && (
            <ActionDropdown variant="secondary">
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onClick={onDelete}
              >
                {__("Delete")}
              </DropdownItem>
            </ActionDropdown>
          )}
        </div>
      </div>

      <div>
        <div className="text-base mb-1">{control.name}</div>
        {control.description && (
          <div className="text-sm text-txt-secondary mb-4 whitespace-pre-wrap">
            {control.description}
          </div>
        )}
        <div className="mb-4">
          <LinkedMeasuresCard
            variant="card"
            measures={
              control.measures?.edges.map(edge => edge.node)
              ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.measures?.__id ?? ""}
            onAttach={withErrorHandling(
              attachMeasure,
              __("Failed to link measure"),
            )}
            onDetach={withErrorHandling(
              detachMeasure,
              __("Failed to unlink measure"),
            )}
            disabled={isAttachingMeasure || isDetachingMeasure}
            readOnly={measuresReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedDocumentsCard
            variant="card"
            documents={
              control.documents?.edges.map(edge => edge.node)
              ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.documents?.__id ?? ""}
            onAttach={withErrorHandling(
              attachDocument,
              __("Failed to link document"),
            )}
            onDetach={withErrorHandling(
              detachDocument,
              __("Failed to unlink document"),
            )}
            disabled={isAttachingDocument || isDetachingDocument}
            readOnly={documentsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedAuditsCard
            variant="card"
            audits={
              control.audits?.edges.map(edge => edge.node) ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.audits?.__id ?? ""}
            onAttach={withErrorHandling(
              attachAudit,
              __("Failed to link audit"),
            )}
            onDetach={withErrorHandling(
              detachAudit,
              __("Failed to unlink audit"),
            )}
            disabled={isAttachingAudit || isDetachingAudit}
            readOnly={auditsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedObligationsCard
            variant="card"
            obligations={
              control.obligations?.edges.map(
                edge => edge.node,
              ) ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.obligations?.__id ?? ""}
            onAttach={withErrorHandling(
              attachObligation,
              __("Failed to link obligation"),
            )}
            onDetach={withErrorHandling(
              detachObligation,
              __("Failed to unlink obligation"),
            )}
            disabled={
              isAttachingObligation || isDetachingObligation
            }
            readOnly={obligationsReadOnly}
          />
        </div>
        <div className="mb-4">
          <LinkedSnapshotsCard
            variant="card"
            snapshots={
              control.snapshots?.edges.map(edge => edge.node)
              ?? []
            }
            params={{ controlId: control.id }}
            connectionId={control.snapshots?.__id ?? ""}
            onAttach={withErrorHandling(
              attachSnapshot,
              __("Failed to link snapshot"),
            )}
            onDetach={withErrorHandling(
              detachSnapshot,
              __("Failed to unlink snapshot"),
            )}
            disabled={isAttachingSnapshot || isDetachingSnapshot}
            readOnly={snapshotsReadOnly}
          />
        </div>
      </div>
    </div>
  );
}
