"use client";

import { useState, useEffect } from "react";
import { ChevronsUpDown, Plus } from "lucide-react";
import { graphql, useFragment } from "react-relay";
import { Link, useNavigate, useParams } from "react-router";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import type {
  OrganizationSwitcher_organizations$key,
  OrganizationSwitcher_organizations$data,
} from "./__generated__/OrganizationSwitcher_organizations.graphql";
import { cn } from "@/lib/utils";

export const organizationSwitcherFragment = graphql`
  fragment OrganizationSwitcher_organizations on Viewer {
    organizations(first: 25)
      @connection(key: "OrganizationSwitcher_organizations") {
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
`;

type Organization =
  OrganizationSwitcher_organizations$data["organizations"]["edges"][0]["node"];

const LogoComponent = ({
  org,
  className,
}: {
  org: Organization;
  className?: string;
}) => {
  if (org.logoUrl) {
    return (
      <img
        src={org.logoUrl}
        alt={org.name}
        className={cn("rounded-md", className)}
      />
    );
  }

  return org.name.substring(0, 2).toUpperCase();
};

export function OrganizationSwitcher({
  organizations,
}: {
  organizations: OrganizationSwitcher_organizations$key;
}) {
  const { isMobile } = useSidebar();
  const data = useFragment(organizationSwitcherFragment, organizations);
  const navigate = useNavigate();
  const { organizationId } = useParams();
  const [currentOrganization, setCurrentOrganization] =
    useState<Organization | null>(null);
  const hasOrganizations =
    data.organizations && data.organizations.edges.length > 0;

  useEffect(() => {
    if (hasOrganizations) {
      const org = data.organizations.edges.find(
        (edge) => edge.node.id === organizationId,
      );
      if (org) {
        setCurrentOrganization(org.node);
      }
    }
  }, [data.organizations, organizationId, hasOrganizations]);

  const handleOrganizationSwitch = (org: Organization) => {
    navigate(`/organizations/${org.id}`);
  };

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              size="lg"
              variant="ghost"
              className={cn(
                "gap-2.5",
                !currentOrganization && "border border-dashed",
              )}
            >
              <div className="flex aspect-square size-8 items-center justify-center rounded-sm bg-highlight-bg">
                {currentOrganization ? (
                  <LogoComponent org={currentOrganization} className="size-8" />
                ) : (
                  <div className="size-8 bg-highlight-bg rounded-sm" />
                )}
              </div>
              <div className="grid text-left leading-tight text-primary">
                {currentOrganization ? (
                  <>
                    <span className="truncate font-medium text-lg leading-5">
                      {currentOrganization.name}
                    </span>
                  </>
                ) : (
                  <span className="truncate text-secondary font-medium">
                    Select Organization
                  </span>
                )}
              </div>
              <ChevronsUpDown className="ml-auto" />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
            align="start"
            side={isMobile ? "bottom" : "right"}
            sideOffset={4}
          >
            <DropdownMenuLabel className="text-xs text-tertiary">
              Organizations
            </DropdownMenuLabel>
            {hasOrganizations &&
              data.organizations.edges.map((edge, index) => (
                <DropdownMenuItem
                  key={edge.node.id}
                  onClick={() => handleOrganizationSwitch(edge.node)}
                  className={`gap-2 p-2 ${
                    edge.node.id === organizationId ? "bg-subtle-bg" : ""
                  }`}
                >
                  <div className="flex size-6 items-center justify-center rounded-sm bg-highlight-bg">
                    <LogoComponent
                      org={edge.node}
                      className="size-4 shrink-0"
                    />
                  </div>
                  {edge.node.name}
                  <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
                </DropdownMenuItem>
              ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild>
              <Link
                to="/organizations/new"
                className="gap-2 p-2 cursor-pointer"
              >
                <div className="flex size-6 items-center justify-center rounded-md border bg-secondary-bg">
                  <Plus className="size-4" />
                </div>
                <div className="font-medium text-tertiary">
                  Add organization
                </div>
              </Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
