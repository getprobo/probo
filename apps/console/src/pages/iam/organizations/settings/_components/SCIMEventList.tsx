import { useTranslate } from "@probo/i18n";
import { Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { graphql, usePaginationFragment } from "react-relay";

import type { SCIMEventListFragment$key } from "#/__generated__/iam/SCIMEventListFragment.graphql";
import type { SCIMEventListPaginationQuery } from "#/__generated__/iam/SCIMEventListPaginationQuery.graphql";
import { SortableTable } from "#/components/SortableTable";

import { SCIMEventListItem } from "./SCIMEventListItem";

const SCIMEventListFragment = graphql`
  fragment SCIMEventListFragment on SCIMConfiguration
  @refetchable(queryName: "SCIMEventListPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    events(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "SCIMEventListFragment_events") {
      edges {
        node {
          id
          ...SCIMEventListItemFragment
        }
      }
    }
  }
`;

export function SCIMEventList(props: { fKey: SCIMEventListFragment$key }) {
  const { fKey } = props;
  const { __ } = useTranslate();

  const eventsPagination = usePaginationFragment<
    SCIMEventListPaginationQuery,
    SCIMEventListFragment$key
  >(SCIMEventListFragment, fKey);

  return (
    <SortableTable
      {...eventsPagination}
      refetch={() => {
        eventsPagination.refetch({}, { fetchPolicy: "network-only" });
      }}
      pageSize={20}
    >
      <Thead>
        <Tr>
          <Th>{__("Time")}</Th>
          <Th>{__("Method")}</Th>
          <Th>{__("Path")}</Th>
          <Th>{__("Result")}</Th>
        </Tr>
      </Thead>
      <Tbody>
        {!eventsPagination.data.events?.edges
          || eventsPagination.data.events.edges.length === 0
          ? (
              <Tr>
                <Td colSpan={4} className="text-center text-txt-secondary">
                  {__("No SCIM events recorded yet.")}
                </Td>
              </Tr>
            )
          : (
              eventsPagination.data.events.edges.map(({ node: event }) => (
                <SCIMEventListItem key={event.id} fKey={event} />
              ))
            )}
      </Tbody>
    </SortableTable>
  );
}
