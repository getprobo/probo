import { useTranslate } from "@probo/i18n";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import type { OrganizationsPageQuery as OrganizationsPageQueryType } from "./__generated__/OrganizationsPageQuery.graphql";
import { useEffect } from "react";
import { Link, useNavigate } from "react-router";
import {
  Avatar,
  Button,
  Card,
  IconPlusLarge,
} from "@probo/ui";
import { usePageTitle } from "@probo/hooks";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { formatDate } from "@probo/helpers";

const OrganizationsPageQuery = graphql`
  query OrganizationsPageQuery {
    viewer {
      organizations(first: 1000, orderBy: {field: NAME, direction: ASC}) @connection(key: "OrganizationsPage_organizations") {
        __id
        edges {
          node {
            id
            name
            logoUrl
          }
        }
      }
      invitations(first: 1000, orderBy: {field: CREATED_AT, direction: DESC}, filter: {status: PENDING}) @connection(key: "OrganizationsPage_invitations") {
        __id
        edges {
          node {
            id
            email
            fullName
            role
            expiresAt
            acceptedAt
            createdAt
            organization {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const acceptInvitationMutation = graphql`
  mutation OrganizationsPage_AcceptInvitationMutation($input: AcceptInvitationInput!) {
    acceptInvitation(input: $input) {
      invitation {
        id
      }
    }
  }
`;

export default function OrganizationsPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const data = useLazyLoadQuery<OrganizationsPageQueryType>(
    OrganizationsPageQuery,
    {}
  );

  const organizations = data.viewer.organizations.edges.map(
    (edge) => edge.node
  );

  const pendingInvitations = data.viewer.invitations.edges.map(
    (edge) => edge.node
  );

  const [acceptInvitation, isAccepting] = useMutationWithToasts(
    acceptInvitationMutation,
    {
      successMessage: __("Invitation accepted successfully"),
      errorMessage: __("Failed to accept invitation"),
    }
  );

  const handleAcceptInvitation = (invitationId: string, organizationId: string) => {
    acceptInvitation({
      variables: {
        input: {
          invitationId,
        },
      },
      onSuccess: () => {
        navigate(`/organizations/${organizationId}`);
      },
    });
  };

  usePageTitle(__("Select an organization"));

  useEffect(() => {
    if (organizations.length === 1 && pendingInvitations.length === 0) {
      navigate(`/organizations/${organizations[0].id}`);
    } else if (organizations.length === 0 && pendingInvitations.length === 0) {
      navigate("/organizations/new");
    }
  }, [organizations, pendingInvitations]);

  return (
    <>
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">
          {__("Select an organization")}
        </h1>
        <div className="space-y-4 w-full">
          {pendingInvitations.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {__("Pending invitations")}
              </h2>
              {pendingInvitations.map((invitation) => (
                <InvitationCard
                  key={invitation.id}
                  invitation={invitation}
                  onAccept={handleAcceptInvitation}
                  isAccepting={isAccepting}
                />
              ))}
            </div>
          )}
          {organizations.length > 0 && (
            <div className="space-y-3">
              {pendingInvitations.length > 0 && (
                <h2 className="text-xl font-semibold">
                  {__("Your organizations")}
                </h2>
              )}
              {organizations.map((organization) => (
                <OrganizationCard
                  key={organization.id}
                  organization={organization}
                />
              ))}
            </div>
          )}
          <Card padded>
            <h2 className="text-xl font-semibold mb-1">
              {__("Create an organization")}
            </h2>
            <p className="text-txt-tertiary mb-4">
              {__("Add a new organization to your account")}
            </p>
            <Button
              to="/organizations/new"
              variant="quaternary"
              icon={IconPlusLarge}
              className="w-full"
            >
              {__("Create organization")}
            </Button>
          </Card>
        </div>
      </div>
    </>
  );
}

type InvitationCardProps = {
  invitation: {
    id: string;
    email: string;
    fullName: string;
    role: string;
    expiresAt: string;
    createdAt: string;
    organization: {
      id: string;
      name: string;
    };
  };
  onAccept: (invitationId: string, organizationId: string) => void;
  isAccepting: boolean;
};

function InvitationCard({ invitation, onAccept, isAccepting }: InvitationCardProps) {
  const { __ } = useTranslate();

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
        <Button
          onClick={() => onAccept(invitation.id, invitation.organization.id)}
          disabled={isAccepting}
        >
          {isAccepting ? __("Accepting...") : __("Accept invitation")}
        </Button>
      </div>
    </Card>
  );
}

type OrganizationCardProps = {
  organization: {
    id: string;
    name: string;
    logoUrl: string | null | undefined;
  };
};

function OrganizationCard({ organization }: OrganizationCardProps) {
  const { __ } = useTranslate();

  return (
    <Card padded className="w-full">
      <div className="flex items-center justify-between">
        <Link
          to={`/organizations/${organization.id}`}
          className="flex items-center gap-4 hover:text-primary flex-1"
        >
          <Avatar
            src={organization.logoUrl}
            name={organization.name}
            size="l"
          />
          <h2 className="font-semibold text-xl">{organization.name}</h2>
        </Link>
        <div className="flex items-center gap-3">
          <Button asChild>
            <Link to={`/organizations/${organization.id}`}>
              {__("Select")}
            </Link>
          </Button>
        </div>
      </div>
    </Card>
  );
}
