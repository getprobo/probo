import {
  ConnectionHandler,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import type { PeopleGraphNodeQuery } from "/hooks/graph/__generated__/PeopleGraphNodeQuery.graphql";
import {
  PeopleConnectionKey,
  peopleNodeQuery,
  useDeletePeople,
} from "/hooks/graph/PeopleGraph";
import {
  ActionDropdown,
  Avatar,
  Breadcrumb,
  DropdownItem,
  IconTrashCan,
  TabLink,
  Tabs,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Outlet } from "react-router";

type Props = {
  queryRef: PreloadedQuery<PeopleGraphNodeQuery>;
};

export default function PeopleDetailPage(props: Props) {
  const data = usePreloadedQuery(peopleNodeQuery, props.queryRef);
  const people = data.node;
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const deletePeople = useDeletePeople(
    people,
    ConnectionHandler.getConnectionID(organizationId, PeopleConnectionKey)
  );

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("People"),
            to: `/organizations/${organizationId}/people`,
          },
          {
            label: data.node.fullName ?? "",
          },
        ]}
      />
      <div className="flex justify-between">
        <div className="space-y-4">
          <Avatar name={people.fullName ?? ""} size="xl" />
          <div className="text-2xl">{people.fullName}</div>
        </div>
        <ActionDropdown variant="secondary">
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onClick={deletePeople}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </div>

      <Tabs>
        <TabLink
          to={`/organizations/${organizationId}/people/${people.id}/tasks`}
        >
          {__("Tasks")}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/people/${people.id}/role`}
        >
          {__("Role & access")}
        </TabLink>
        <TabLink
          to={`/organizations/${organizationId}/people/${people.id}/profile`}
        >
          {__("General information")}
        </TabLink>
      </Tabs>

      <Outlet context={{ people }} />
    </div>
  );
}
