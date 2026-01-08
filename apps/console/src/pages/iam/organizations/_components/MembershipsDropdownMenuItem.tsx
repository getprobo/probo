import {
  Avatar,
  DropdownItem,
  IconCheckmark1,
  IconClock,
  IconLock,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import type { MembershipsDropdownMenuItemFragment$key } from "/__generated__/iam/MembershipsDropdownMenuItemFragment.graphql";
import { useFragment, useMutation } from "react-relay";
import { parseDate } from "@probo/helpers";
import { useCallback } from "react";
import { useNavigate } from "react-router";
import type { MembershipsDropdownMenuItem_assumeMutation } from "/__generated__/iam/MembershipsDropdownMenuItem_assumeMutation.graphql";

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

const assumeOrganizationSessionMutation = graphql`
  mutation MembershipsDropdownMenuItem_assumeMutation(
    $input: AssumeOrganizationSessionInput!
  ) {
    assumeOrganizationSession(input: $input) {
      result {
        __typename
        ... on OrganizationSessionCreated {
          membership {
            id
            lastSession {
              id
              expiresAt
            }
          }
        }
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

export function MembershipsDropdownMenuItem(props: {
  fKey: MembershipsDropdownMenuItemFragment$key;
}) {
  const { fKey } = props;

  const navigate = useNavigate();

  const { id, lastSession, organization } =
    useFragment<MembershipsDropdownMenuItemFragment$key>(fragment, fKey);

  const isAuthenticated = !!lastSession;
  const isExpired =
    lastSession && parseDate(lastSession.expiresAt) < new Date();

  const [assumeOrganizationSession] =
    useMutation<MembershipsDropdownMenuItem_assumeMutation>(
      assumeOrganizationSessionMutation,
    );

  const handleAssumeOrganizationSession = useCallback(() => {
    if (isAuthenticated) {
      navigate(`/organizations/${organization.id}`);
    } else {
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
              navigate("/auth/login");
              break;
            case "SAMLAuthenticationRequired":
              window.location.href = result.redirectUrl;
              break;
            default:
              navigate(`/organizations/${organization.id}`);
          }
        },
      });
    }
  }, [assumeOrganizationSession, navigate, organization.id, isAuthenticated]);

  return (
    <DropdownItem key={id} onClick={handleAssumeOrganizationSession}>
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
