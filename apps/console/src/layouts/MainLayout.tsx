import { Link, Navigate, Outlet, useParams } from "react-router";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconBank,
  IconBook,
  IconCircleQuestionmark,
  IconClock,
  IconCrossLargeX,
  IconFire3,
  IconGroup1,
  IconInboxEmpty,
  IconPageTextLine,
  IconSettingsGear2,
  IconStore,
  IconTodo,
  IconListStack,
  IconBox,
  IconShield,
  IconRotateCw,
  IconCircleProgress,
  IconMedal,
  Layout,
  SidebarItem,
  UserDropdown as UserDropdownRoot,
  UserDropdownItem,
  Skeleton,
  Dropdown,
  Button,
  DropdownItem,
  IconChevronGrabberVertical,
  IconPlusLarge,
  Avatar,
  IconPeopleAdd,
  Badge,
  IconLock,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type { MainLayoutQuery as MainLayoutQueryType } from "./__generated__/MainLayoutQuery.graphql";
import { Suspense, useState, useEffect } from "react";
import { useToast } from "@probo/ui";
import { ErrorBoundary } from "react-error-boundary";
import { PageError } from "/components/PageError";
import { buildEndpoint } from "/providers/RelayProviders";

const MainLayoutQuery = graphql`
  query MainLayoutQuery($organizationId: ID!) {
    viewer {
      id
      user {
        fullName
        email
      }
      invitations(first: 1, filter: {statuses: [PENDING]}) {
        totalCount
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

/**
 * Site layout with a header and a sidebar
 */
export function MainLayout() {
  const { organizationId } = useParams();
  const { __ } = useTranslate();

  const prefix = `/organizations/${organizationId}`;

  if (!organizationId) {
    return <Navigate to="/" />;
  }

  return (
    <Layout
      header={
        <>
          <div className="mr-auto">
            <Suspense fallback={<Skeleton className="w-20 h-8" />}>
              <OrganizationSelectorWrapper organizationId={organizationId} />
            </Suspense>
          </div>
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <UserDropdown organizationId={organizationId} />
          </Suspense>
        </>
      }
      sidebar={
        <ul className="space-y-[2px]">
          <SidebarItem
            label={__("Tasks")}
            icon={IconInboxEmpty}
            to={`${prefix}/tasks`}
          />
          <SidebarItem
            label={__("Measures")}
            icon={IconTodo}
            to={`${prefix}/measures`}
          />
          <SidebarItem
            label={__("Risks")}
            icon={IconFire3}
            to={`${prefix}/risks`}
          />
          <SidebarItem
            label={__("Frameworks")}
            icon={IconBank}
            to={`${prefix}/frameworks`}
          />
          <SidebarItem
            label={__("People")}
            icon={IconGroup1}
            to={`${prefix}/people`}
          />
          <SidebarItem
            label={__("Vendors")}
            icon={IconStore}
            to={`${prefix}/vendors`}
          />
          <SidebarItem
            label={__("Documents")}
            icon={IconPageTextLine}
            to={`${prefix}/documents`}
          />
          <SidebarItem
            label={__("Assets")}
            icon={IconBox}
            to={`${prefix}/assets`}
          />
          <SidebarItem
            label={__("Data")}
            icon={IconListStack}
            to={`${prefix}/data`}
          />
          <SidebarItem
            label={__("Audits")}
            icon={IconMedal}
            to={`${prefix}/audits`}
          />
          <SidebarItem
            label={__("Nonconformities")}
            icon={IconCrossLargeX}
            to={`${prefix}/nonconformities`}
          />
          <SidebarItem
            label={__("Obligations")}
            icon={IconBook}
            to={`${prefix}/obligations`}
          />
           <SidebarItem
            label={__("Continual Improvements")}
            icon={IconRotateCw}
            to={`${prefix}/continual-improvements`}
          />
          <SidebarItem
            label={__("Processing Activities")}
            icon={IconCircleProgress}
            to={`${prefix}/processing-activities`}
          />
          <SidebarItem
            label={__("Snapshots")}
            icon={IconClock}
            to={`${prefix}/snapshots`}
          />
          <SidebarItem
            label={__("Trust Center")}
            icon={IconShield}
            to={`${prefix}/trust-center`}
          />
          <SidebarItem
            label={__("Settings")}
            icon={IconSettingsGear2}
            to={`${prefix}/settings`}
          />
        </ul>
      }
    >
      <ErrorBoundary FallbackComponent={PageError}>
        <Outlet />
      </ErrorBoundary>
    </Layout>
  );
}

function UserDropdown({ organizationId }: { organizationId: string }) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const user = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, { organizationId }).viewer
    .user;

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = async (
    e
  ) => {
    e.preventDefault();

    fetch(buildEndpoint("/auth/logout"), {
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

function OrganizationSelectorWrapper({ organizationId }: { organizationId: string }) {
  const data = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, { organizationId });
  return <OrganizationSelector viewer={data.viewer} currentOrganization={data.organization} />;
}

function OrganizationSelector({
  viewer,
  currentOrganization
}: {
  viewer: MainLayoutQueryType["response"]["viewer"];
  currentOrganization: MainLayoutQueryType["response"]["organization"];
}) {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { __ } = useTranslate();

  const pendingInvitationsCount = viewer.invitations.totalCount;

  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        setIsLoading(true);
        const response = await fetch('/auth/organizations', {
          credentials: 'include',
        });

        if (!response.ok) {
          throw new Error('Failed to fetch organizations');
        }

        const data: OrganizationsResponse = await response.json();
        setOrganizations(data.organizations);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
        console.error('Failed to fetch organizations:', err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchOrganizations();
  }, []);

  if (error) {
    return (
      <div className="flex items-center gap-1">
        <Button
          className="-ml-3"
          variant="tertiary"
          disabled
        >
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
            {isLoading ? __("Loading...") : (currentOrganization?.name || "")}
          </Button>
        }
      >
        <div className="max-h-150 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-transparent hover:scrollbar-thumb-gray-400">
          {isLoading ? (
            <div className="px-3 py-2 text-gray-500">
              {__("Loading organizations...")}
            </div>
          ) : organizations.length === 0 ? (
            <div className="px-3 py-2 text-gray-500">
              {__("No organizations found")}
            </div>
          ) : (
            organizations.map((organization) => {
              const isAuthenticated = organization.authStatus === "authenticated";
              const isExpired = organization.authStatus === "expired";
              const needsAuth = organization.authStatus === "unauthenticated";

              const targetUrl = isAuthenticated
                ? `/organizations/${organization.id}`
                : organization.loginUrl;

              const isSAMLUrl = targetUrl.includes('/auth/saml/');

              return (
                <DropdownItem
                  asChild
                  key={organization.id}
                >
                  {isSAMLUrl ? (
                    <a href={targetUrl} className="flex items-center gap-2">
                      <Avatar name={organization.name} src={organization.logoUrl} />
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
                      <Avatar name={organization.name} src={organization.logoUrl} />
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
