import { parseDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Badge,
  Button,
  Card,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MembershipCardFragment$key } from "#/__generated__/iam/MembershipCardFragment.graphql";

const fragment = graphql`
  fragment MembershipCardFragment on Membership {
    lastSession {
      id
      expiresAt
    }
    organization @required(action: THROW) {
      id
      name
      logoUrl
    }
  }
`;

interface MembershipCardProps {
  fKey: MembershipCardFragment$key;
}

export function MembershipCard(props: MembershipCardProps) {
  const { fKey } = props;
  const { __ } = useTranslate();

  const { lastSession, organization } = useFragment<MembershipCardFragment$key>(
    fragment,
    fKey,
  );
  const isExpired
    = lastSession && parseDate(lastSession.expiresAt) < new Date();
  const isAssuming = !!lastSession && !isExpired;

  const getAuthBadge = () => {
    if (isAssuming) {
      return (
        <Badge variant="success" className="flex items-center gap-1">
          <IconCheckmark1 size={14} />
          {__("Authenticated")}
        </Badge>
      );
    } else if (isExpired) {
      return (
        <Badge variant="warning" className="flex items-center gap-1">
          <IconClock size={14} />
          {__("Session expired")}
        </Badge>
      );
    } else {
      return (
        <Badge variant="neutral" className="flex items-center gap-1">
          <IconLock size={14} />
          {__("Authentication required")}
        </Badge>
      );
    }
  };

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 hover:text-primary flex-1">
          <Avatar
            src={organization.logoUrl}
            name={organization.name}
            size="l"
          />
          <div className="flex flex-col gap-1">
            <h2 className="font-semibold text-xl">{organization.name}</h2>
            {getAuthBadge()}
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Link to={`/organizations/${organization.id}`}>
            {isAssuming
              ? <Button variant="secondary">{__("Start")}</Button>
              : <Button>{__("Login")}</Button>}
          </Link>
        </div>
      </div>
    </Card>
  );
}
