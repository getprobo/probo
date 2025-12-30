import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Card } from "@probo/ui";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { InvitationCardFragment$key } from "/__generated__/iam/InvitationCardFragment.graphql";

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

interface InvitationCardProps {
  fKey: InvitationCardFragment$key;
  // onAccept: (invitationId: string, organizationId: string) => void;
  // isAccepting: boolean;
}

export function InvitationCard(props: InvitationCardProps) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const invitation = useFragment<InvitationCardFragment$key>(fragment, fKey);

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
        {/* <Button
          onClick={() => onAccept(invitation.id, invitation.organization.id)}
          disabled={isAccepting}
        >
          {isAccepting ? __("Accepting...") : __("Accept invitation")}
        </Button> */}
      </div>
    </Card>
  );
}
