"use client";

import {
  ChevronRight,
  SquareCheck,
  Inbox,
  Store,
  type LucideIcon,
  Flame,
} from "lucide-react";
import { Link, useLocation, useParams } from "react-router";

import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
import {
  SidebarGroup,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from "@/components/ui/sidebar";

import { BookOpen, FileText, Settings, Users, Box, Database } from "lucide-react";

interface NavItem {
  title: string;
  url?: string;
  icon: LucideIcon;
  isActive?: boolean;
  items?: {
    title: string;
    url: string;
    icon: LucideIcon;
  }[];
}

export function NavMain() {
  const location = useLocation();
  const { organizationId } = useParams();
  const items: NavItem[] = getNavItems(organizationId);

  if (!organizationId) {
    return null;
  }

  const isItemActive = (item: { url?: string; items?: { url: string }[] }) => {
    if (item.url && location.pathname.startsWith(item.url)) {
      return true;
    }

    if (item.items?.length) {
      return item.items.some((subItem) =>
        location.pathname.startsWith(subItem.url),
      );
    }

    return false;
  };

  return (
    <SidebarGroup>
      <SidebarMenu>
        {items.map((item) => {
          const active = isItemActive(item) || item.isActive;

          return (
            <Collapsible key={item.title} asChild defaultOpen={active}>
              <SidebarMenuItem>
                <SidebarMenuButton
                  asChild
                  tooltip={item.title}
                  data-active={active ? "true" : undefined}
                >
                  <Link to={item.url ?? "#"}>
                    <item.icon
                      className={active ? "text-lime-9" : "text-lime-6"}
                    />
                    <span className="font-medium">{item.title}</span>
                  </Link>
                </SidebarMenuButton>
                {item.items?.length ? (
                  <>
                    <CollapsibleTrigger asChild>
                      <SidebarMenuAction className="data-[state=open]:rotate-90">
                        <ChevronRight />
                        <span className="sr-only">Toggle</span>
                      </SidebarMenuAction>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                      <SidebarMenuSub>
                        {item.items?.map((subItem) => {
                          const subItemActive = location.pathname.startsWith(
                            subItem.url,
                          );

                          return (
                            <SidebarMenuSubItem key={subItem.title}>
                              <SidebarMenuSubButton
                                asChild
                                data-active={subItemActive ? "true" : undefined}
                              >
                                <Link to={subItem.url}>
                                  <subItem.icon
                                    className={
                                      subItemActive
                                        ? "text-lime-9"
                                        : "text-lime-6"
                                    }
                                  />
                                  <span className="font-medium">
                                    {subItem.title}
                                  </span>
                                </Link>
                              </SidebarMenuSubButton>
                            </SidebarMenuSubItem>
                          );
                        })}
                      </SidebarMenuSub>
                    </CollapsibleContent>
                  </>
                ) : null}
              </SidebarMenuItem>
            </Collapsible>
          );
        })}
      </SidebarMenu>
    </SidebarGroup>
  );
}

function getNavItems(organizationId?: string): NavItem[] {
  return [
    {
      title: "Tasks",
      icon: Inbox,
      url: organizationId
        ? `/organizations/${organizationId}/tasks`
        : undefined,
    },
    {
      title: "Measures",
      icon: SquareCheck,
      url: organizationId
        ? `/organizations/${organizationId}/measures`
        : undefined,
    },
    {
      title: "Risks",
      icon: Flame,
      url: organizationId
        ? `/organizations/${organizationId}/risks`
        : undefined,
    },
    {
      title: "Frameworks",
      url: organizationId
        ? `/organizations/${organizationId}/frameworks`
        : undefined,
      icon: BookOpen,
    },
    {
      title: "People",
      url: `/organizations/${organizationId}/people`,
      icon: Users,
    },
    {
      title: "Vendors",
      url: `/organizations/${organizationId}/vendors`,
      icon: Store,
    },
    {
      title: "Documents",
      url: organizationId
        ? `/organizations/${organizationId}/documents`
        : undefined,
      icon: FileText,
    },
    {
      title: "Assets",
      url: organizationId
        ? `/organizations/${organizationId}/assets`
        : undefined,
      icon: Box,
    },
    {
      title: "Data",
      url: organizationId
        ? `/organizations/${organizationId}/data`
        : undefined,
      icon: Database,
    },
    {
      title: "Settings",
      url: organizationId
        ? `/organizations/${organizationId}/settings`
        : undefined,
      icon: Settings,
    },
  ];
}
