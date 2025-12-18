import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, Card, IconLock } from "@probo/ui";
import { Link } from "react-router";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { OrganizationCardFragment$key } from "./__generated__/OrganizationCardFragment.graphql";

const fragment = graphql`
  fragment OrganizationCardFragment on Organization {
    id
    name
    logoUrl
  }
`;

interface OrganizationCardProps {
  fKey: OrganizationCardFragment$key;
}

export function OrganizationCard(props: OrganizationCardProps) {
  const { fKey } = props;
  const { __ } = useTranslate();

  const organization = useFragment<OrganizationCardFragment$key>(
    fragment,
    fKey,
  );
  // const isAuthenticated = organization.authStatus === "authenticated";
  // const isExpired = organization.authStatus === "expired";
  // const needsAuth = organization.authStatus === "unauthenticated";

  // Determine target URL and button text based on auth status
  // const targetUrl = isAuthenticated
  //   ? `/organizations/${organization.id}`
  //   : organization.loginUrl;
  const targetUrl = `/organizations/${organization.id}`;

  const getAuthBadge = () => {
    // if (isAuthenticated) {
    //   return (
    //     <Badge variant="success" className="flex items-center gap-1">
    //       <IconCheckmark1 size={14} />
    //       {__("Authenticated")}
    //     </Badge>
    //   );
    // }

    // if (isExpired) {
    //   return (
    //     <Badge variant="warning" className="flex items-center gap-1">
    //       <IconClock size={14} />
    //       {__("Session expired")}
    //     </Badge>
    //   );
    // }

    // if (needsAuth) {
    return (
      <Badge variant="neutral" className="flex items-center gap-1">
        <IconLock size={14} />
        {__("Authentication required")}
      </Badge>
    );
    // }

    return null;
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
