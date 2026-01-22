import { formatDate, formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, useToast } from "@probo/ui";
import { useFragment, useMutation } from "react-relay";
import { type DataID, graphql } from "relay-runtime";

import type { InvitationCardFragment$key } from "#/__generated__/iam/InvitationCardFragment.graphql";
import type { InvitationCardMutation } from "#/__generated__/iam/InvitationCardMutation.graphql";

const fragment = graphql`
  fragment InvitationCardFragment on Invitation {
    id
    role
    createdAt
    organization @required(action: THROW) {
      id
      name
    }
  }
`;

const acceptMutation = graphql`
  mutation InvitationCardMutation(
    $input: AcceptInvitationInput!
    $membershipConnections: [ID!]!
    $pendingInvitationConnections: [ID!]!
  ) {
    acceptInvitation(input: $input) {
      invitation {
        id @deleteEdge(connections: $pendingInvitationConnections)
      }
      membershipEdge @prependEdge(connections: $membershipConnections) {
        node {
          id
          ...MembershipCardFragment
          organization {
            name
          }
        }
      }
    }
  }
`;

interface InvitationCardProps {
  pendingInvitationsConnectionId: DataID;
  membershipConnectionId: DataID;
  fKey: InvitationCardFragment$key;
}

export function InvitationCard(props: InvitationCardProps) {
  const { pendingInvitationsConnectionId, membershipConnectionId, fKey }
    = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const invitation = useFragment<InvitationCardFragment$key>(fragment, fKey);

  const [acceptInvitation, isAccepting]
    = useMutation<InvitationCardMutation>(acceptMutation);

  const handleAccept = () => {
    acceptInvitation({
      variables: {
        input: {
          invitationId: invitation.id,
        },
        pendingInvitationConnections: [pendingInvitationsConnectionId],
        membershipConnections: [membershipConnectionId],
      },
      onCompleted: (_, e) => {
        if (e) {
          toast({
            title: __("Request failed"),
            description: formatError(__("Cannot accept invitation"), e),
            variant: "error",
          });
          return;
        }
      },
      onError: (e) => {
        toast({
          title: __("Request failed"),
          description: e.message,
          variant: "error",
        });
      },
    });
  };

  return (
    <Card padded className="w-full">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 space-y-1">
          <h3 className="text-lg font-semibold">
            {invitation.organization.name}
          </h3>
          <p className="text-sm text-txt-secondary">
            {__("Role")}
            :
            <span className="font-medium">{invitation.role}</span>
          </p>
          <p className="text-xs text-txt-tertiary">
            {__("Invited on")}
            {" "}
            {formatDate(invitation.createdAt)}
          </p>
        </div>
        <Button onClick={handleAccept} disabled={isAccepting}>
          {isAccepting ? __("Accepting...") : __("Accept invitation")}
        </Button>
      </div>
    </Card>
  );
}
