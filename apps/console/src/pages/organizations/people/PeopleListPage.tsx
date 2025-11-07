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
import { Authorized } from "/permissions";
import { isAuthorized } from "/permissions";

type People = NodeOf<PeopleGraphPaginatedFragment$data["peoples"]>;

const isContractEnded = (person: People): boolean => {
  if (!person.contractEndDate) return false;
  const endDate = new Date(person.contractEndDate);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return endDate < today;
};

export default function PeopleListPage({
  queryRef,
}: {
  queryRef: PreloadedQuery<PeopleGraphPaginatedQuery>;
}) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { people, refetch, connectionId, hasNext, loadNext, isLoadingNext } =
    usePeopleQuery(queryRef);

  usePageTitle(__("Members"));

  const hasAnyAction = isAuthorized(organizationId, "People", "updatePeople") ||
    isAuthorized(organizationId, "People", "deletePeople");

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Members")}
        description={__(
          "Keep track of your company's workforce and their progress towards completing tasks assigned to them."
        )}
      >
        <Authorized entity="Organization" action="createPeople">
          <CreatePeopleDialog connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add member")}</Button>
          </CreatePeopleDialog>
        </Authorized>
      </PageHeader>
      <SortableTable
        refetch={refetch}
        hasNext={hasNext}
        loadNext={loadNext}
        isLoadingNext={isLoadingNext}
      >
        <Thead>
          <Tr>
            <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
            <SortableTh field="KIND">{__("Role")}</SortableTh>
            <Th>{__("Position")}</Th>
            {hasAnyAction && <Th>{__("Actions")}</Th>}
          </Tr>
        </Thead>
        <Tbody>
          {people.map((person) => (
            <PeopleRow
              key={person.id}
              people={person}
              connectionId={connectionId}
              hasAnyAction={hasAnyAction}
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
  hasAnyAction,
}: {
  people: People;
  connectionId: string;
  hasAnyAction: boolean;
}) {
  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const deletePeople = useDeletePeople(people, connectionId);
  const contractEnded = isContractEnded(people);

  return (
    <Tr
      to={`/organizations/${organizationId}/people/${people.id}/tasks`}
      className={contractEnded ? "opacity-50" : ""}
    >
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
      <Td className="text-sm">{people.position}</Td>
      {hasAnyAction && (
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <Authorized entity="People" action="deletePeople">
              <DropdownItem
                icon={IconTrashCan}
                variant="danger"
                onClick={deletePeople}
              >
                {__("Delete")}
              </DropdownItem>
            </Authorized>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
