import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Badge,
  Button,
  Dropdown,
  DropdownItem,
  DropdownSeparator,
  IconArrowBoxLeft,
  IconBank,
  IconBook,
  IconBox,
  IconCalendar2,
  IconCheckmark1,
  IconChevronGrabberVertical,
  IconCircleProgress,
  IconCircleQuestionmark,
  IconClock,
  IconCrossLargeX,
  IconFire3,
  IconGroup1,
  IconInboxEmpty,
  IconKey,
  IconListStack,
  IconLock,
  IconMagnifyingGlass,
  IconMedal,
  IconPageTextLine,
  IconPeopleAdd,
  IconPlusLarge,
  IconRotateCw,
  IconSettingsGear2,
  IconShield,
  IconStore,
  IconTodo,
  Input,
  Layout,
  SidebarItem,
  Skeleton,
  UserDropdownItem,
  UserDropdown as UserDropdownRoot,
  useToast,
} from "@probo/ui";
import { Suspense, use, useEffect, useMemo, useState } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { useLazyLoadQuery } from "react-relay";
import { Link, Navigate, Outlet, useParams } from "react-router";
import { graphql } from "relay-runtime";
import type { MainLayoutQuery as MainLayoutQueryType } from "/__generated__/core/MainLayoutQuery.graphql";
import { PageError } from "/components/PageError";
import { PermissionsProvider } from "/providers/PermissionsProvider";
import { PermissionsContext } from "/providers/PermissionsContext";

const MainLayoutQuery = graphql`
  query MainLayoutQuery($organizationId: ID!) {
    viewer {
      id
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

  const prefix = `/organizations/${organizationId}`;

  if (!organizationId) {
    return <Navigate to="/" />;
  }

  return (
    <Suspense fallback={<Skeleton className="w-full h-screen" />}>
      <PermissionsProvider>
        <MainLayoutContent organizationId={organizationId} prefix={prefix} />
      </PermissionsProvider>
    </Suspense>
  );
}

function MainLayoutContent({
  organizationId,
  prefix,
}: {
  organizationId: string;
  prefix: string;
}) {
  const { __ } = useTranslate();
  const { isAuthorized } = use(PermissionsContext);
  const data = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, {
    organizationId,
  });

  return (
    <Layout
      header={
        <>
          <div className="mr-auto">
            <OrganizationSelector currentOrganization={data.organization} />
          </div>
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <UserDropdown organizationId={organizationId} />
          </Suspense>
        </>
      }
      sidebar={
        <ul className="space-y-[2px]">
          {isAuthorized("Organization", "listMeetings") && (
            <SidebarItem
              label={__("Meetings")}
              icon={IconCalendar2}
              to={`${prefix}/meetings`}
            />
          )}
          {isAuthorized("Organization", "listTasks") && (
            <SidebarItem
              label={__("Tasks")}
              icon={IconInboxEmpty}
              to={`${prefix}/tasks`}
            />
          )}
          {isAuthorized("Organization", "listMeasures") && (
            <SidebarItem
              label={__("Measures")}
              icon={IconTodo}
              to={`${prefix}/measures`}
            />
          )}
          {isAuthorized("Organization", "listRisks") && (
            <SidebarItem
              label={__("Risks")}
              icon={IconFire3}
              to={`${prefix}/risks`}
            />
          )}
          {isAuthorized("Organization", "listFrameworks") && (
            <SidebarItem
              label={__("Frameworks")}
              icon={IconBank}
              to={`${prefix}/frameworks`}
            />
          )}
          {isAuthorized("Organization", "listPeople") && (
            <SidebarItem
              label={__("People")}
              icon={IconGroup1}
              to={`${prefix}/people`}
            />
          )}
          {isAuthorized("Organization", "listVendors") && (
            <SidebarItem
              label={__("Vendors")}
              icon={IconStore}
              to={`${prefix}/vendors`}
            />
          )}
          {isAuthorized("Organization", "listDocuments") && (
            <SidebarItem
              label={__("Documents")}
              icon={IconPageTextLine}
              to={`${prefix}/documents`}
            />
          )}
          {isAuthorized("Organization", "listAssets") && (
            <SidebarItem
              label={__("Assets")}
              icon={IconBox}
              to={`${prefix}/assets`}
            />
          )}
          {isAuthorized("Organization", "listData") && (
            <SidebarItem
              label={__("Data")}
              icon={IconListStack}
              to={`${prefix}/data`}
            />
          )}
          {isAuthorized("Organization", "listAudits") && (
            <SidebarItem
              label={__("Audits")}
              icon={IconMedal}
              to={`${prefix}/audits`}
            />
          )}
          {isAuthorized("Organization", "listNonconformities") && (
            <SidebarItem
              label={__("Nonconformities")}
              icon={IconCrossLargeX}
              to={`${prefix}/nonconformities`}
            />
          )}
          {isAuthorized("Organization", "listObligations") && (
            <SidebarItem
              label={__("Obligations")}
              icon={IconBook}
              to={`${prefix}/obligations`}
            />
          )}
          {isAuthorized("Organization", "listContinualImprovements") && (
            <SidebarItem
              label={__("Continual Improvements")}
              icon={IconRotateCw}
              to={`${prefix}/continual-improvements`}
            />
          )}
          {isAuthorized("Organization", "listProcessingActivities") && (
            <SidebarItem
              label={__("Processing Activities")}
              icon={IconCircleProgress}
              to={`${prefix}/processing-activities`}
            />
          )}
          {isAuthorized("Organization", "listRightsRequests") && (
            <SidebarItem
              label={__("Rights Requests")}
              icon={IconLock}
              to={`${prefix}/rights-requests`}
            />
          )}
          {isAuthorized("Organization", "listSnapshots") && (
            <SidebarItem
              label={__("Snapshots")}
              icon={IconClock}
              to={`${prefix}/snapshots`}
            />
          )}
          {isAuthorized("Organization", "getTrustCenter") && (
            <SidebarItem
              label={__("Trust Center")}
              icon={IconShield}
              to={`${prefix}/trust-center`}
            />
          )}
          {isAuthorized("Organization", "listMembers") && (
            <SidebarItem
              label={__("Settings")}
              icon={IconSettingsGear2}
              to={`${prefix}/settings`}
            />
          )}
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
  const { isAuthorized } = use(PermissionsContext);
  // const user = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, {
  //   organizationId,
  // }).viewer.user;
  const user = {
    fullName: "",
    email: "",
  };

  const handleLogout: React.MouseEventHandler<HTMLAnchorElement> = async (
    e,
  ) => {
    e.preventDefault();

    fetch("/connect/logout", {
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
      {isAuthorized("Organization", "listSignableDocuments") && (
        <UserDropdownItem
          to={`/organizations/${organizationId}/employee`}
          icon={IconPageTextLine}
          label={__("My Signatures")}
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
  currentOrganization: MainLayoutQueryType["response"]["organization"];
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
      org.name.toLowerCase().includes(search.toLowerCase()),
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
          (inv) => !inv.acceptedAt,
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
        <div className="px-3 py-2">
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
        <div className="max-h-150 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-transparent hover:scrollbar-thumb-gray-400">
          {isLoading ? (
            <div className="px-3 py-2 text-gray-500">
              {__("Loading organizations...")}
            </div>
          ) : filteredOrganizations.length === 0 ? (
            <div className="px-3 py-2 text-txt-secondary">
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

              // Use organization endpoint for all logos for consistency
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
