import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card } from "@probo/ui";
import { graphql } from "relay-runtime";
import { useFragment, useMutation } from "react-relay";
import type { InvitationCardFragment$key } from "/__generated__/iam/InvitationCardFragment.graphql";
import { useNavigate } from "react-router";

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
  mutation InvitationCardMutation($input: AcceptInvitationInput!) {
    acceptInvitation(input: $input) {
      membershipEdge {
        node {
          id
        }
      }
    }
  }
`;

interface InvitationCardProps {
  fKey: InvitationCardFragment$key;
}

export function InvitationCard(props: InvitationCardProps) {
  const { fKey } = props;

  const navigate = useNavigate();
  const { __ } = useTranslate();

  const invitation = useFragment<InvitationCardFragment$key>(fragment, fKey);

  const [acceptInvitation, isAccepting] = useMutation(acceptMutation);

  const handleAccept = () => {
    acceptInvitation({
      variables: {
        input: {
          invitationId: invitation.id,
        },
      },
      onCompleted: () => {
        navigate(`/organizations/${invitation.organization.id}`);
      },
      onError: (err) => {
        console.error("Failed to accept invitation:", err);
        alert(__("Failed to accept invitation"));
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
            {__("Role")}: <span className="font-medium">{invitation.role}</span>
          </p>
          <p className="text-xs text-txt-tertiary">
            {__("Invited on")} {formatDate(invitation.createdAt)}
          </p>
        </div>
        <Button onClick={handleAccept} disabled={isAccepting}>
          {isAccepting ? __("Accepting...") : __("Accept invitation")}
        </Button>
      </div>
    </Card>
  );
}
