import { graphql, usePaginationFragment } from "react-relay";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { SortableTable, SortableTh } from "/components/SortableTable";
import type { MemberListFragment_RefetchQuery } from "/__generated__/iam/MemberListFragment_RefetchQuery.graphql";
import { MemberListItem } from "./MemberListItem";
import type { MemberListFragment$key } from "/__generated__/iam/MemberListFragment.graphql";
import type { MemberListItem_permissionsFragment$key } from "/__generated__/iam/MemberListItem_permissionsFragment.graphql";

const fragment = graphql`
  fragment MemberListFragment on Organization
  @refetchable(queryName: "MemberListFragment_RefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: {
      type: "MembershipOrder"
      defaultValue: { direction: ASC, field: CREATED_AT }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    ...MemberListItem_currentRoleFragment
    members(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MemberListFragment_members") @required(action: THROW) {
      __id
      totalCount
      edges @required(action: THROW) {
        node {
          id
          ...MemberListItemFragment
        }
      }
    }
  }
`;

export function MemberList(props: {
  fKey: MemberListFragment$key;
  permissionsFKey: MemberListItem_permissionsFragment$key;
}) {
  const { fKey, permissionsFKey } = props;

  const { __ } = useTranslate();

  const membersPagination = usePaginationFragment<
    MemberListFragment_RefetchQuery,
    MemberListFragment$key
  >(fragment, fKey);

  const refetchMemberships = () => {
    membersPagination.refetch({}, { fetchPolicy: "network-only" });
  };

  return (
    <SortableTable
      {...membersPagination}
      refetch={({ order }: { order: { direction: string; field: string } }) => {
        membersPagination.refetch({
          order: {
            direction: order.direction as "ASC" | "DESC",
            field: order.field as
              | "CREATED_AT"
              // FIXME: add those back
              // | "FULL_NAME"
              // | "EMAIL_ADDRESS"
              | "ROLE",
          },
        });
      }}
      pageSize={20}
    >
      <Thead>
        <Tr>
          <SortableTh field="FULL_NAME">{__("Name")}</SortableTh>
          <SortableTh field="EMAIL_ADDRESS">{__("Email")}</SortableTh>
          <SortableTh field="ROLE">{__("Role")}</SortableTh>
          <SortableTh field="CREATED_AT">{__("Joined")}</SortableTh>
          <Th></Th>
        </Tr>
      </Thead>
      <Tbody>
        {membersPagination.data.members.totalCount === 0 ? (
          <Tr>
            <Td colSpan={5} className="text-center text-txt-secondary">
              {__("No members")}
            </Td>
          </Tr>
        ) : (
          membersPagination.data.members.edges.map(({ node: membership }) => (
            <MemberListItem
              connectionId={membersPagination.data.members.__id}
              key={membership.id}
              fKey={membership}
              onRefetch={refetchMemberships}
              permissionsFKey={permissionsFKey}
              viewerFKey={membersPagination.data}
            />
          ))
        )}
      </Tbody>
    </SortableTable>
  );
}
