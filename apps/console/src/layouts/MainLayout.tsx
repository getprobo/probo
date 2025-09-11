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
  IconCheckmark1,
  IconStore,
  IconTodo,
  IconListStack,
  IconBox,
  IconShield,
  IconRotateCw,
  IconCircleProgress,
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
  IconChevronDown,
  Avatar,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery, usePaginationFragment } from "react-relay";
import type { MainLayoutQuery as MainLayoutQueryType } from "./__generated__/MainLayoutQuery.graphql";
import type { MainLayout_OrganizationSelector_viewer$key } from "./__generated__/MainLayout_OrganizationSelector_viewer.graphql";
import { Suspense, useState } from "react";
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
      ...MainLayout_OrganizationSelector_viewer
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

const OrganizationSelectorFragment = graphql`
  fragment MainLayout_OrganizationSelector_viewer on Viewer
  @refetchable(queryName: "MainLayoutOrganizationSelectorPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 25 }
    after: { type: "CursorKey" }
  ) {
    organizations(first: $first, after: $after, orderBy: {field: NAME, direction: ASC})
    @connection(key: "MainLayout_OrganizationSelector_organizations") {
      edges {
        node {
          id
          name
          logoUrl
        }
      }
      pageInfo {
        hasNextPage
        endCursor
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
            icon={IconCheckmark1}
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

    fetch(buildEndpoint("/api/console/v1/auth/logout"), {
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

function OrganizationSelectorWrapper({ organizationId }: { organizationId: string }) {
  const data = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, { organizationId });
  return <OrganizationSelector viewer={data.viewer} currentOrganization={data.organization} />;
}

function OrganizationSelector({
  viewer,
  currentOrganization
}: {
  viewer: MainLayout_OrganizationSelector_viewer$key;
  currentOrganization: MainLayoutQueryType["response"]["organization"];
}) {
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const { __ } = useTranslate();

  const { data, loadNext, hasNext } = usePaginationFragment(
    OrganizationSelectorFragment,
    viewer
  );

  const organizations = data.organizations.edges.map((edge) => edge.node);

  const handleLoadMore = (e?: React.MouseEvent) => {
    e?.preventDefault();
    e?.stopPropagation();

    if (hasNext && !isLoadingMore) {
      setIsLoadingMore(true);
      loadNext(25, {
        onComplete: () => setIsLoadingMore(false),
      });
    }
  };

  return (
    <Dropdown
      toggle={
        <Button
          className="-ml-3"
          variant="tertiary"
          iconAfter={IconChevronGrabberVertical}
        >
          {currentOrganization?.name || ""}
        </Button>
      }
    >
      <div className="max-h-150 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-transparent hover:scrollbar-thumb-gray-400">
        {organizations.map((organization) => (
          <DropdownItem
            asChild
            key={organization.id}
          >
            <Link to={`/organizations/${organization.id}`}>
              <Avatar src={organization.logoUrl} name={organization.name} />
              {organization.name}
            </Link>
          </DropdownItem>
        ))}
        {hasNext && (
          <div className="px-3 py-1 flex justify-center">
            <Button
              variant="tertiary"
              onClick={handleLoadMore}
              onMouseDown={(e: React.MouseEvent) => {
                e.preventDefault();
                e.stopPropagation();
              }}
              className="mx-auto"
              icon={IconChevronDown}
              disabled={isLoadingMore}
            >
              {isLoadingMore ? __("Loading...") : __("Show More")}
            </Button>
          </div>
        )}
      </div>
      <DropdownSeparator />
      <DropdownItem asChild icon={IconPlusLarge}>
        <Link to="/organizations/new">
          <IconPlusLarge size={16} />
          {__("Add organization")}
        </Link>
      </DropdownItem>
    </Dropdown>
  );
}
