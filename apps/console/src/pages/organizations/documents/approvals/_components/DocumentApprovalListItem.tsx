import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconCircleCheck,
  IconCircleX,
  IconClock,
  IconTrashCan,
  Spinner,
  useToast,
} from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentApprovalListItem_removeApproverMutation } from "#/__generated__/core/DocumentApprovalListItem_removeApproverMutation.graphql";
import type { DocumentApprovalListItemFragment$key } from "#/__generated__/core/DocumentApprovalListItemFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment DocumentApprovalListItemFragment on DocumentVersionApprovalDecision {
    id
    approver {
      fullName
    }
    state
    comment
    decidedAt
    createdAt
    canApprove: permission(action: "core:document-version:approve")
    canReject: permission(action: "core:document-version:reject")
    documentVersion {
      id
      document {
        id
      }
    }
  }
`;

const removeApproverMutation = graphql`
  mutation DocumentApprovalListItem_removeApproverMutation(
    $input: RemoveDocumentVersionApproverInput!
    $connections: [ID!]!
  ) {
    removeDocumentVersionApprover(input: $input) {
      deletedApprovalDecisionId @deleteEdge(connections: $connections)
      documentVersion {
        id
        approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
          edges {
            node {
              id
              status
              decisions(first: 0) {
                totalCount
              }
              approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
                totalCount
              }
            }
          }
        }
      }
    }
  }
`;

export function DocumentApprovalListItem(props: {
  fragmentRef: DocumentApprovalListItemFragment$key;
  canManage: boolean;
  connectionId: string;
  onRefetch: () => void;
}) {
  const { fragmentRef, canManage, connectionId, onRefetch } = props;
  const { __, dateTimeFormat } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  const decision = useFragment(fragment, fragmentRef);

  const isPending = decision.state === "PENDING";
  const isApproved = decision.state === "APPROVED";
  const isRejected = decision.state === "REJECTED";

  const [removeApprover, isRemoving] = useMutation<DocumentApprovalListItem_removeApproverMutation>(
    removeApproverMutation,
  );

  const reviewUrl = `/organizations/${organizationId}/employee/approvals/${decision.documentVersion.document.id}`;

  return (
    <div className="flex gap-3 items-center py-3">
      <div className="space-y-1">
        <div className="text-sm text-txt-primary font-medium">
          {decision.approver.fullName}
        </div>
        <div className="text-xs text-txt-secondary flex items-center gap-1">
          {isApproved && <IconCircleCheck size={16} className="text-txt-accent" />}
          {isRejected && <IconCircleX size={16} className="text-txt-danger" />}
          {isPending && <IconClock size={16} />}
          <span>
            {isPending && sprintf(__("Requested on %s"), dateTimeFormat(decision.createdAt))}
            {isApproved && sprintf(__("Approved on %s"), dateTimeFormat(decision.decidedAt))}
            {isRejected && sprintf(__("Rejected on %s"), dateTimeFormat(decision.decidedAt))}
          </span>
        </div>
        {decision.comment && (
          <div className="text-xs text-txt-secondary italic">
            {decision.comment}
          </div>
        )}
      </div>
      <div className="ml-auto flex items-center gap-2">
        {isApproved && (
          <Badge variant="success">{__("Approved")}</Badge>
        )}
        {isRejected && (
          <Badge variant="danger">{__("Rejected")}</Badge>
        )}
        {isPending && (decision.canApprove || decision.canReject) && (
          <Button variant="secondary" to={reviewUrl} target="_blank">
            {__("Review")}
          </Button>
        )}
        {canManage && (
          <Button
            variant="quaternary"
            icon={isRemoving ? Spinner : IconTrashCan}
            disabled={isRemoving}
            onClick={() => {
              void removeApprover({
                variables: {
                  input: {
                    approvalDecisionId: decision.id,
                  },
                  connections: [connectionId],
                },
                onCompleted: (_data, errors) => {
                  if (errors?.length) {
                    toast({
                      title: __("Error"),
                      description: errors[0].message,
                      variant: "error",
                    });
                    return;
                  }
                  toast({
                    title: __("Approver removed"),
                    description: __("The approver has been removed successfully."),
                    variant: "success",
                  });
                  onRefetch();
                },
                onError: (error) => {
                  toast({
                    title: __("Error"),
                    description: error.message,
                    variant: "error",
                  });
                },
              });
            }}
          />
        )}
        {isPending && !decision.canApprove && !decision.canReject && !canManage && (
          <Badge variant="warning">{__("Pending")}</Badge>
        )}
      </div>
    </div>
  );
}
