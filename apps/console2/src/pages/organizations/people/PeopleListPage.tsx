import {
  Button,
  IconPlusLarge,
  PageHeader,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Avatar,
  ActionDropdown,
  DropdownItem,
  IconTrashCan,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { PeopleGraphPaginatedQuery } from "/hooks/graph/__generated__/PeopleGraphPaginatedQuery.graphql";
import { type PreloadedQuery } from "react-relay";
import { useDeletePeople, usePeopleQuery } from "/hooks/graph/PeopleGraph";
import { SortableTable, SortableTh } from "/components/SortableTable";
import type { PeopleGraphPaginatedFragment$data } from "/hooks/graph/__generated__/PeopleGraphPaginatedFragment.graphql";
import type { NodeOf } from "/types";
import { usePageTitle } from "@probo/hooks";
import { getRole } from "@probo/helpers";
import { CreatePeopleDialog } from "./dialogs/CreatePeopleDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";

type People = NodeOf<PeopleGraphPaginatedFragment$data["peoples"]>;

export default function PeopleListPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<PeopleGraphPaginatedQuery>;
}) {
  const { __ } = useTranslate();
  const { people, refetch, connectionId } = usePeopleQuery(queryRef);

  usePageTitle(__("Members"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Members")}
        description={__(
          "Keep track of your company's workforce and their progress towards completing tasks assigned to them."
        )}
      >
        <CreatePeopleDialog connectionId={connectionId}>
          <Button icon={IconPlusLarge}>{__("Add member")}</Button>
        </CreatePeopleDialog>
      </PageHeader>
      <SortableTable refetch={refetch}>
        <Thead>
          <Tr>
            <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
            <SortableTh field="KIND">{__("Role")}</SortableTh>
            <Th>{__("Actions")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {people.map((person) => (
            <PeopleRow
              key={person.id}
              people={person}
              connectionId={connectionId}
            />
          ))}
        </Tbody>
      </SortableTable>
    </div>
  );
}

function PeopleRow({
  people,
  connectionId,
}: {
  people: People;
  connectionId: string;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deletePeople = useDeletePeople(people, connectionId);

  return (
    <Tr to={`/organizations/${organizationId}/people/${people.id}/tasks`}>
      <Td>
        <div className="flex gap-3 items-center">
          <Avatar name={people.fullName} />
          <div>
            <div className="text-sm">{people.fullName}</div>
            <div className="text-xs text-txt-tertiary">
              {people.primaryEmailAddress}
            </div>
          </div>
        </div>
      </Td>
      <Td className="text-sm">{getRole(__, people.kind)}</Td>
      <Td noLink width={50} className="text-end">
        <ActionDropdown>
          <DropdownItem
            icon={IconTrashCan}
            variant="danger"
            onClick={deletePeople}
          >
            {__("Delete")}
          </DropdownItem>
        </ActionDropdown>
      </Td>
    </Tr>
  );
}
