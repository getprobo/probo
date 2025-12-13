import { Link, Navigate, Outlet, useParams } from "react-router";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconCircleQuestionmark,
  UserDropdown as UserDropdownRoot,
  UserDropdownItem,
  Skeleton,
  Dropdown,
  Button,
  DropdownItem,
  IconChevronGrabberVertical,
  IconLock,
  IconKey,
  IconPeopleAdd,
  IconPlusLarge,
  IconCheckmark1,
  IconClock,
  IconMagnifyingGlass,
  useToast,
  Logo,
  Toasts,
  ConfirmDialog,
  Avatar,
  Badge,
  Input,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type { EmployeeLayoutQuery as EmployeeLayoutQueryType } from "./__generated__/EmployeeLayoutQuery.graphql";
import { Suspense, useState, useEffect, useMemo, use } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { PageError } from "/components/PageError";
import { buildEndpoint } from "/providers/RelayProviders";
import { PermissionsProvider } from "/providers/PermissionsProvider";
import { PermissionsContext } from "/providers/PermissionsContext";

const EmployeeLayoutQuery = graphql`
  query EmployeeLayoutQuery($organizationId: ID!) {
    viewer {
      id
      user {
        fullName
        email
      }
    }
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        name
        logoUrl
      }
    }
  }
`;

export function EmployeeLayout() {
  const { organizationId } = useParams();

  if (!organizationId) {
    return <Navigate to="/" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <PermissionsProvider>
        <EmployeeLayoutContent organizationId={organizationId} />
      </PermissionsProvider>
    </Suspense>
  );
}

function EmployeeLayoutContent({
  organizationId,
}: {
  organizationId: string;
}) {
  const data = useLazyLoadQuery<EmployeeLayoutQueryType>(EmployeeLayoutQuery, {
    organizationId,
  });

  return (
    <div className="text-txt-primary bg-level-0">
      <header className="absolute z-2 left-0 right-0 px-4 flex items-center border-b border-border-solid h-12 bg-level-0">
        <Logo className="w-12 h-5" />
        <svg
          className="mx-3 text-txt-tertiary"
          width="8"
          height="18"
          viewBox="0 0 8 18"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path d="M1 17L7 1" stroke="currentColor" />
        </svg>
        <div className="mr-auto">
          <OrganizationSelector currentOrganization={data.organization} />
        </div>
        <Suspense fallback={<Skeleton className="w-32 h-8" />}>
          <UserDropdown organizationId={organizationId} />
        </Suspense>
      </header>
      <main className="overflow-y-auto w-full pt-12 h-[calc(100vh-3rem)]">
        <div className="px-8 pb-8 pt-8">
          <ErrorBoundary FallbackComponent={PageError}>
            <Outlet />
          </ErrorBoundary>
        </div>
      </main>
      <Toasts />
      <ConfirmDialog />
    </div>
  );
}

interface Organization {
  id: string;
  name: string;
  logoUrl?: string | null;
  authenticationMethod: string;
  authStatus: "authenticated" | "unauthenticated" | "expired";
  loginUrl: string;
}

interface OrganizationsResponse {
  organizations: Organization[];
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

interface InvitationsResponse {
  invitations: Invitation[];
}

function OrganizationSelector({
  currentOrganization,
}: {
  currentOrganization: EmployeeLayoutQueryType["response"]["organization"];
}) {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [pendingInvitationsCount, setPendingInvitationsCount] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const { __ } = useTranslate();

  const filteredOrganizations = useMemo(() => {
    if (!search.trim()) {
      return organizations;
    }
    return organizations.filter((org) =>
      org.name.toLowerCase().includes(search.toLowerCase())
    );
  }, [organizations, search]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setIsLoading(true);

        const [orgsResponse, invitationsResponse] = await Promise.all([
          fetch("/connect/organizations", { credentials: "include" }),
          fetch("/connect/invitations", { credentials: "include" }),
        ]);

        if (!orgsResponse.ok) {
          throw new Error("Failed to fetch organizations");
        }

        if (!invitationsResponse.ok) {
          throw new Error("Failed to fetch invitations");
        }

        const orgsData: OrganizationsResponse = await orgsResponse.json();
        const invitationsData: InvitationsResponse =
          await invitationsResponse.json();

        const pendingCount = invitationsData.invitations.filter(
          (inv) => !inv.acceptedAt
        ).length;

        setOrganizations(orgsData.organizations);
        setPendingInvitationsCount(pendingCount);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Unknown error");
        console.error("Failed to fetch data:", err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, []);

  if (error) {
    return (
      <div className="flex items-center gap-1">
        <Button className="-ml-3" variant="tertiary" disabled>
          {__("Error loading organizations")}
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-1">
      <Dropdown
        toggle={
          <Button
            className="-ml-3"
            variant="tertiary"
            iconAfter={IconChevronGrabberVertical}
            disabled={isLoading}
          >
            {isLoading ? __("Loading...") : currentOrganization?.name || ""}
          </Button>
        }
      >
        {!isLoading && organizations.length > 0 && (
          <div className="px-3 py-2 border-b border-border-low">
            <Input
              icon={IconMagnifyingGlass}
              placeholder={__("Search organizations...")}
              value={search}
              onValueChange={setSearch}
              onKeyDown={(e) => { 
                  e.stopPropagation();
              }}
              autoFocus
            />
          </div>
        )}
        <div className="max-h-150 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-transparent hover:scrollbar-thumb-gray-400">
          {isLoading ? (
            <div className="px-3 py-2 text-gray-500">
              {__("Loading organizations...")}
            </div>
          ) : filteredOrganizations.length === 0 ? (
            <div className="px-3 py-2 text-gray-500">
              {__("No organizations found")}
            </div>
          ) : (
            filteredOrganizations.map((organization) => {
              const isAuthenticated =
                organization.authStatus === "authenticated";
              const isExpired = organization.authStatus === "expired";
              const needsAuth = organization.authStatus === "unauthenticated";

              const targetUrl = isAuthenticated
                ? `/organizations/${organization.id}`
                : organization.loginUrl;

              const isSAMLUrl = targetUrl.includes("/connect/saml/");

              const logoUrl = organization.logoUrl;

              return (
                <DropdownItem asChild key={organization.id}>
                  {isSAMLUrl ? (
                    <a href={targetUrl} className="flex items-center gap-2">
                      <Avatar name={organization.name} src={logoUrl} />
                      <span className="flex-1">{organization.name}</span>
                      {isAuthenticated && (
                        <IconCheckmark1 size={16} className="text-green-600" />
                      )}
                      {isExpired && (
                        <IconClock size={16} className="text-orange-600" />
                      )}
                      {needsAuth && (
                        <IconLock size={16} className="text-gray-400" />
                      )}
                    </a>
                  ) : (
                    <Link to={targetUrl} className="flex items-center gap-2">
                      <Avatar name={organization.name} src={logoUrl} />
                      <span className="flex-1">{organization.name}</span>
                      {isAuthenticated && (
                        <IconCheckmark1 size={16} className="text-green-600" />
                      )}
                      {isExpired && (
                        <IconClock size={16} className="text-orange-600" />
                      )}
                      {needsAuth && (
                        <IconLock size={16} className="text-gray-400" />
                      )}
                    </Link>
                  )}
                </DropdownItem>
              );
            })
          )}
        </div>
        <DropdownSeparator />
        {pendingInvitationsCount > 0 && (
          <DropdownItem asChild>
            <Link to="/">
              <IconPeopleAdd size={16} />
              <span className="flex-1">{__("Invitations")}</span>
              <Badge variant="info" size="sm">
                {pendingInvitationsCount}
              </Badge>
            </Link>
          </DropdownItem>
        )}
        <DropdownItem asChild>
          <Link to="/organizations/new">
            <IconPlusLarge size={16} />
            {__("Add organization")}
          </Link>
        </DropdownItem>
      </Dropdown>
      {pendingInvitationsCount > 0 && (
        <Link to="/" className="relative" title={__("Invitations")}>
          <Button variant="tertiary" icon={IconPeopleAdd} />
          <Badge
            variant="info"
            size="sm"
            className="absolute -top-1 -right-1 min-w-[20px] h-5 flex items-center justify-center"
          >
            {pendingInvitationsCount}
          </Badge>
        </Link>
      )}
    </div>
  );
}

function UserDropdown({ organizationId }: { organizationId: string }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const { isAuthorized } = use(PermissionsContext);
  const user = useLazyLoadQuery<EmployeeLayoutQueryType>(EmployeeLayoutQuery, {
    organizationId,
  }).viewer.user;

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = async (
    e
  ) => {
    e.preventDefault();

    fetch(buildEndpoint("/connect/logout"), {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({}),
    })
      .then(async (res) => {
        if (!res.ok) {
          const error = await res.json();
          throw new Error(error.message || __("Failed to login"));
        }

        window.location.reload();
      })
      .catch((e) => {
        toast({
          title: __("Error"),
          description: e.message as string,
          variant: "error",
        });
      });
  };

  return (
    <UserDropdownRoot fullName={user.fullName} email={user.email}>
      {isAuthorized("Organization", "deleteOrganization") && (
        <UserDropdownItem
          to="/api-keys"
          icon={IconKey}
          label={__("API Keys")}
        />
      )}
      <UserDropdownItem
        to="mailto:support@getprobo.com"
        icon={IconCircleQuestionmark}
        label={__("Help")}
      />
      <DropdownSeparator />
      <UserDropdownItem
        variant="danger"
        to="/logout"
        icon={IconArrowBoxLeft}
        label="Logout"
        onClick={handleLogout}
      />
    </UserDropdownRoot>
  );
}
