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
import { Link } from "react-router";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { MembershipCardFragment$key } from "./__generated__/MembershipCardFragment.graphql";
import { parseDate } from "@probo/helpers";

const fragment = graphql`
  fragment MembershipCardFragment on Membership {
    activeSession {
      id
      expiresAt
    }
    organization {
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

  const { activeSession, organization } =
    useFragment<MembershipCardFragment$key>(fragment, fKey);
  const isAuthenticated = !!activeSession;
  const isExpired =
    activeSession && parseDate(activeSession.expiresAt) >= new Date();

  // Determine target URL and button text based on auth status
  // const targetUrl = isAuthenticated
  //   ? `/organizations/${organization.id}`
  //   : organization.loginUrl;
  const targetUrl = `/organizations/${organization.id}`;

  const getAuthBadge = () => {
    if (isAuthenticated) {
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

  // const getButtonText = () => {
  //   if (isAuthenticated) return __("Select");
  //   if (organization.authenticationMethod === "saml")
  //     return __("Login with SAML");
  //   return __("Login");
  // };

  // Check if the URL is a backend SAML endpoint
  const isSAMLUrl = targetUrl.includes("/connect/saml/");

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        {isSAMLUrl ? (
          <a
            href={targetUrl}
            className="flex items-center gap-4 hover:text-primary flex-1"
          >
            <Avatar
              src={organization.logoUrl}
              name={organization.name}
              size="l"
            />
            <div className="flex flex-col gap-1">
              <h2 className="font-semibold text-xl">{organization.name}</h2>
              {getAuthBadge()}
            </div>
          </a>
        ) : (
          <Link
            to={targetUrl}
            className="flex items-center gap-4 hover:text-primary flex-1"
          >
            <Avatar
              src={organization.logoUrl}
              name={organization.name}
              size="l"
            />
            <div className="flex flex-col gap-1">
              <h2 className="font-semibold text-xl">{organization.name}</h2>
              {getAuthBadge()}
            </div>
          </Link>
        )}
        <div className="flex items-center gap-3">
          <Button asChild>
            {/* {isSAMLUrl ? (
              <a href={targetUrl}>{getButtonText()}</a>
            ) : (
              <Link to={targetUrl}>{getButtonText()}</Link>
            )} */}
            <Link to={targetUrl}>LOGIN</Link>
          </Button>
        </div>
      </div>
    </Card>
  );
}
