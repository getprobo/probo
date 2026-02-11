import { parseDate } from "@probo/helpers";
import {
  Avatar,
  DropdownItem,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MembershipsDropdownMenuItemFragment$key } from "#/__generated__/iam/MembershipsDropdownMenuItemFragment.graphql";

const fragment = graphql`
  fragment MembershipsDropdownMenuItemFragment on Membership {
    id
    lastSession {
      id
      expiresAt
    }
    organization @required(action: THROW) {
      id
      logoUrl
      name
    }
  }
`;

export function MembershipsDropdownMenuItem(props: {
  fKey: MembershipsDropdownMenuItemFragment$key;
}) {
  const { fKey } = props;

  const { id, lastSession, organization }
    = useFragment<MembershipsDropdownMenuItemFragment$key>(fragment, fKey);

  const isAssuming = !!lastSession;
  const isExpired
    = lastSession && parseDate(lastSession.expiresAt) < new Date();

  return (
    <DropdownItem key={id} asChild>
      <Link to={`/organizations/${organization.id}`}>
        <Avatar name={organization.name} src={organization.logoUrl} />
        <span className="flex-1">{organization.name}</span>
        {isAssuming && (
          <IconCheckmark1 size={16} className="text-green-600" />
        )}
        {isExpired && <IconClock size={16} className="text-orange-600" />}
        {!lastSession && <IconLock size={16} className="text-gray-400" />}
      </Link>
    </DropdownItem>
  );
}
