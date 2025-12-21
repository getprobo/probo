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
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { useFragment, useMutation } from "react-relay";
import type { MembershipCardFragment$key } from "./__generated__/MembershipCardFragment.graphql";
import { parseDate } from "@probo/helpers";
import { useCallback } from "react";
import type { MembershipCard_assumeMutation } from "./__generated__/MembershipCard_assumeMutation.graphql";

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

const assumeOrganizationSessionMutation = graphql`
  mutation MembershipCard_assumeMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
      result {
        __typename
        ... on PasswordRequired {
          reason
        }
        ... on SAMLAuthenticationRequired {
          reason
          redirectUrl
        }
      }
    }
  }
`;

interface MembershipCardProps {
  fKey: MembershipCardFragment$key;
}

export function MembershipCard(props: MembershipCardProps) {
  const { fKey } = props;
  const { __ } = useTranslate();
  const navigate = useNavigate();

  const { lastSession, organization } = useFragment<MembershipCardFragment$key>(
    fragment,
    fKey,
  );
  const isAuthenticated = !!lastSession;
  const isExpired =
    lastSession && parseDate(lastSession.expiresAt) < new Date();

  const [assumeOrganizationSession] =
    useMutation<MembershipCard_assumeMutation>(
      assumeOrganizationSessionMutation,
    );

  const handleAssumeOrganizationSession = useCallback(() => {
    assumeOrganizationSession({
      variables: {
        input: {
          organizationId: organization.id,
        },
      },
      onCompleted: ({ assumeOrganizationSession }) => {
        if (!assumeOrganizationSession) {
          throw new Error("complete mutation result is empty");
        }

        const { result } = assumeOrganizationSession;

        switch (result.__typename) {
          case "PasswordRequired":
            navigate("auth/login");
            break;
          case "SAMLAuthenticationRequired":
            window.location.href = result.redirectUrl;
            break;
          default:
            navigate(`/organizations/${organization.id}`);
        }
      },
    });
  }, [assumeOrganizationSession, navigate, organization.id]);

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
          <Button onClick={handleAssumeOrganizationSession}>
            {__("Login")}
          </Button>
        </div>
      </div>
    </Card>
  );
}
