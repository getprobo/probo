import { Outlet, useParams } from "react-router";
import {
  DropdownSeparator,
  IconArrowBoxLeft,
  IconBank,
  IconCircleQuestionmark,
  IconFire3,
  IconGroup1,
  IconInboxEmpty,
  IconPageTextLine,
  IconSettingsGear2,
  IconStore,
  IconTodo,
  Layout,
  SidebarItem,
  UserDropdown,
  UserDropdownItem,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useLazyLoadQuery } from "react-relay";
import type { MainLayoutQuery as MainLayoutQueryType } from "./__generated__/MainLayoutQuery.graphql";

const MainLayoutQuery = graphql`
  query MainLayoutQuery {
    viewer {
      id
      user {
        fullName
        email
      }
    }
  }
`;

export function MainLayout() {
  const { organizationId } = useParams();
  const { __ } = useTranslate();
  const user = useLazyLoadQuery<MainLayoutQueryType>(MainLayoutQuery, {}).viewer
    .user;

  const prefix = `/organizations/${organizationId}`;

  return (
    <Layout
      header={
        <>
          <div className="mr-auto"></div>
          <UserDropdown fullName={user.fullName} email={user.email}>
            <UserDropdownItem
              to="/settings"
              icon={IconSettingsGear2}
              label={__("Settings")}
            />
            <UserDropdownItem
              to="/settings"
              icon={IconCircleQuestionmark}
              label={__("Help")}
            />
            <DropdownSeparator />
            <UserDropdownItem
              variant="danger"
              to="/logout"
              icon={IconArrowBoxLeft}
              label="Logout"
            />
          </UserDropdown>
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
            label={__("Policies")}
            icon={IconPageTextLine}
            to={`${prefix}/policies`}
          />
          <SidebarItem
            label={__("Settings")}
            icon={IconSettingsGear2}
            to={`${prefix}/settings`}
          />
        </ul>
      }
    >
      <Outlet />
    </Layout>
  );
}
