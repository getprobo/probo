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
  IconCalendar2,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { PeopleGraphPaginatedQuery } from "/__generated__/core/PeopleGraphPaginatedQuery.graphql";
import { type PreloadedQuery } from "react-relay";
import { useDeletePeople, usePeopleQuery } from "/hooks/graph/PeopleGraph";
import { SortableTable, SortableTh } from "/components/SortableTable";
import type { PeopleGraphPaginatedFragment$data } from "/__generated__/core/PeopleGraphPaginatedFragment.graphql";
import type { NodeOf } from "/types";
import { usePageTitle } from "@probo/hooks";
import { getRole } from "@probo/helpers";
import { CreatePeopleDialog } from "./dialogs/CreatePeopleDialog";
import { SetEndOfContractDialog, type SetEndOfContractDialogRef } from "./dialogs/SetEndOfContractDialog";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PermissionsContext } from "/providers/PermissionsContext";
import { use, useRef } from "react";

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
  const { isAuthorized } = use(PermissionsContext);
  const { people, refetch, connectionId, hasNext, loadNext, isLoadingNext } =
    usePeopleQuery(queryRef);

  usePageTitle(__("Members"));

  const hasAnyAction =
    isAuthorized("People", "updatePeople") ||
    isAuthorized("People", "deletePeople");

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Members")}
        description={__(
          "Keep track of your company's workforce and their progress towards completing tasks assigned to them.",
        )}
      >
        {isAuthorized("Organization", "createPeople") && (
          <CreatePeopleDialog connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add member")}</Button>
          </CreatePeopleDialog>
        )}
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
  const { isAuthorized } = use(PermissionsContext);
  const dialogRef = useRef<SetEndOfContractDialogRef>(null);

  return (
    <>
      <SetEndOfContractDialog
        peopleId={people.id}
        currentContractEndDate={people.contractEndDate}
        ref={dialogRef}
      />
      <Tr
        to={`/organizations/${organizationId}/people/${people.id}/profile`}
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
              {isAuthorized("People", "updatePeople") && (
                <DropdownItem
                  icon={IconCalendar2}
                  onClick={() => dialogRef.current?.open()}
                >
                  {__("Set end of contract")}
                </DropdownItem>
              )}
              {isAuthorized("People", "deletePeople") && (
                <DropdownItem
                  icon={IconTrashCan}
                  variant="danger"
                  onClick={deletePeople}
                >
                  {__("Delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          </Td>
        )}
      </Tr>
    </>
  );
}
