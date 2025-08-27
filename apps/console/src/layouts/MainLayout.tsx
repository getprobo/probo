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
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type { MainLayoutQuery as MainLayoutQueryType } from "./__generated__/MainLayoutQuery.graphql";
import { Suspense } from "react";
import { useToast } from "@probo/ui";
import { ErrorBoundary } from "react-error-boundary";
import { PageError } from "/components/PageError";
import { buildEndpoint } from "/providers/RelayProviders";

const MainLayoutQuery = graphql`
  query MainLayoutQuery {
    viewer {
      id
      user {
        fullName
        email
      }
      organizations(first: 25) @connection(key: "MainLayout_organizations") {
        __id
        edges {
          node {
            id
            name
            logoUrl
          }
        }
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
              <OrganizationSelector organizationId={organizationId} />
            </Suspense>
          </div>
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <UserDropdown />
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
            label={__("Nonconformity Registries")}
            icon={IconCrossLargeX}
            to={`${prefix}/nonconformity-registries`}
          />
          <SidebarItem
            label={__("Compliance Registries")}
            icon={IconBook}
            to={`${prefix}/compliance-registries`}
          />
           <SidebarItem
            label={__("Continual Improvement Registries")}
            icon={IconRotateCw}
            to={`${prefix}/continual-improvement-registries`}
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

function UserDropdown() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const user = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, {}).viewer
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

function OrganizationSelector({ organizationId }: { organizationId: string }) {
  const organizations = useLazyLoadQuery<MainLayoutQueryType>(
    MainLayoutQuery,
    {}
  ).viewer.organizations.edges.map((edge) => edge.node);
  const currentOrganization = organizations.find(
    (organization) => organization.id === organizationId
  );
  const { __ } = useTranslate();

  return (
    <Dropdown
      toggle={
        <Button
          className="-ml-3"
          variant="tertiary"
          iconAfter={IconChevronGrabberVertical}
        >
          {currentOrganization?.name}
        </Button>
      }
    >
      {organizations.map((organization) => (
        <DropdownItem
          asChild
          key={organization.id}
          icon={IconCircleQuestionmark}
        >
          <Link to={`/organizations/${organization.id}`}>
            <Avatar src={organization.logoUrl} name={organization.name} />
            {organization.name}
          </Link>
        </DropdownItem>
      ))}
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
