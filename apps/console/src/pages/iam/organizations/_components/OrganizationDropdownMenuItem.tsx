import {
  Avatar,
  DropdownItem,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import type { OrganizationDropdownMenuItemFragment$key } from "./__generated__/OrganizationDropdownMenuItemFragment.graphql";
import { useFragment } from "react-relay";
import { parseDate } from "@probo/helpers";

const fragment = graphql`
  fragment OrganizationDropdownMenuItemFragment on Membership {
    id
    lastSession {
      id
      expiresAt
    }
    organization @required(action: THROW) {
      logoUrl
      name
    }
  }
`;

export function OrganizationDropdownMenuItem(props: {
  fKey: OrganizationDropdownMenuItemFragment$key;
}) {
  const { fKey } = props;
  const { id, lastSession, organization } =
    useFragment<OrganizationDropdownMenuItemFragment$key>(fragment, fKey);

  const isAuthenticated = !!lastSession;
  const isExpired =
    lastSession && parseDate(lastSession.expiresAt) >= new Date();

  return (
    <DropdownItem key={id}>
      {/* TODO add link or anchor */}
      <Avatar name={organization.name} src={organization.logoUrl} />
      <span className="flex-1">{organization.name}</span>
      {isAuthenticated && (
        <IconCheckmark1 size={16} className="text-green-600" />
      )}
      {isExpired && <IconClock size={16} className="text-orange-600" />}
      {!lastSession && <IconLock size={16} className="text-gray-400" />}
    </DropdownItem>
  );
}
