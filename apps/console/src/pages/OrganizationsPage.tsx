import { useTranslate } from "@probo/i18n";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router";
import {
  Avatar,
  Button,
  Card,
  IconPlusLarge,
  IconCheckmark1,
  IconLock,
  IconClock,
  Badge,
} from "@probo/ui";
import { usePageTitle } from "@probo/hooks";
import { formatDate } from "@probo/helpers";

interface Organization {
  id: string;
  name: string;
  logoUrl?: string | null;
  authenticationMethod: string;
  authStatus: "authenticated" | "unauthenticated" | "expired";
  loginUrl: string;
}

interface Invitation {
  id: string;
  email: string;
  fullName: string;
  role: string;
  expiresAt: string;
  acceptedAt?: string | null;
  createdAt: string;
  organization: {
    id: string;
    name: string;
  };
}

export default function OrganizationsPage() {
  const { __ } = useTranslate();
  const navigate = useNavigate();

  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoadingOrganizations, setIsLoadingOrganizations] = useState(true);
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [isLoadingInvitations, setIsLoadingInvitations] = useState(true);
  const [isAccepting, setIsAccepting] = useState(false);

  // Fetch organizations from REST endpoint
  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        const response = await fetch('/connect/organizations', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to fetch organizations');
        }

        const data: { organizations: Organization[] } = await response.json();
        setOrganizations(data.organizations);
      } catch (err) {
        console.error('Failed to fetch organizations:', err);
      } finally {
        setIsLoadingOrganizations(false);
      }
    };

    fetchOrganizations();
  }, []);

  // Fetch pending invitations from REST endpoint
  useEffect(() => {
    const fetchInvitations = async () => {
      try {
        const response = await fetch('/connect/invitations', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to fetch invitations');
        }

        const data: { invitations: Invitation[] } = await response.json();
        setInvitations(data.invitations);
      } catch (err) {
        console.error('Failed to fetch invitations:', err);
      } finally {
        setIsLoadingInvitations(false);
      }
    };

    fetchInvitations();
  }, []);

  const handleAcceptInvitation = async (invitationId: string, organizationId: string) => {
    setIsAccepting(true);
    try {
      const response = await fetch('/connect/invitations/accept', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ invitationId }),
      });

      if (!response.ok) {
        throw new Error('Failed to accept invitation');
      }

      // Navigate to the organization after successful acceptance
      navigate(`/organizations/${organizationId}`);
    } catch (err) {
      console.error('Failed to accept invitation:', err);
      alert(__('Failed to accept invitation'));
    } finally {
      setIsAccepting(false);
    }
  };

  usePageTitle(__("Select an organization"));

  useEffect(() => {
    // Only auto-navigate once both organizations and invitations are loaded
    if (!isLoadingOrganizations && !isLoadingInvitations) {
      if (organizations.length === 1 && invitations.length === 0) {
        navigate(`/organizations/${organizations[0].id}`);
      } else if (organizations.length === 0 && invitations.length === 0) {
        navigate("/organizations/new");
      }
    }
  }, [organizations, invitations, isLoadingOrganizations, isLoadingInvitations, navigate]);

  return (
    <>
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">
          {__("Select an organization")}
        </h1>
        <div className="space-y-4 w-full">
          {invitations.length > 0 && (
            <div className="space-y-3">
              <h2 className="text-xl font-semibold">
                {__("Pending invitations")}
              </h2>
              {invitations.map((invitation) => (
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
              {invitations.length > 0 && (
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
  invitation: Invitation;
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
  organization: Organization;
};

function OrganizationCard({ organization }: OrganizationCardProps) {
  const { __ } = useTranslate();

  const isAuthenticated = organization.authStatus === "authenticated";
  const isExpired = organization.authStatus === "expired";
  const needsAuth = organization.authStatus === "unauthenticated";

  // Determine target URL and button text based on auth status
  const targetUrl = isAuthenticated
    ? `/organizations/${organization.id}`
    : organization.loginUrl;

  const getAuthBadge = () => {
    if (isAuthenticated) {
      return (
        <Badge variant="success" className="flex items-center gap-1">
          <IconCheckmark1 size={14} />
          {__("Authenticated")}
        </Badge>
      );
    }

    if (isExpired) {
      return (
        <Badge variant="warning" className="flex items-center gap-1">
          <IconClock size={14} />
          {__("Session expired")}
        </Badge>
      );
    }

    if (needsAuth) {
      return (
        <Badge variant="neutral" className="flex items-center gap-1">
          <IconLock size={14} />
          {__("Authentication required")}
        </Badge>
      );
    }

    return null;
  };

  const getButtonText = () => {
    if (isAuthenticated) return __("Select");
    if (organization.authenticationMethod === "saml") return __("Login with SAML");
    return __("Login");
  };

  // Check if the URL is a backend SAML endpoint
  const isSAMLUrl = targetUrl.includes('/connect/saml/');

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
            {isSAMLUrl ? (
              <a href={targetUrl}>
                {getButtonText()}
              </a>
            ) : (
              <Link to={targetUrl}>
                {getButtonText()}
              </Link>
            )}
          </Button>
        </div>
      </div>
    </Card>
  );
}
