import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconChevronDown, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageMailingListDeleteMutation } from "#/__generated__/core/CompliancePageMailingListDeleteMutation.graphql";
import type { CompliancePageMailingListFragment$key } from "#/__generated__/core/CompliancePageMailingListFragment.graphql";
import type { CompliancePageMailingListQuery } from "#/__generated__/core/CompliancePageMailingListQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteMutation = graphql`
  mutation CompliancePageMailingListDeleteMutation(
    $input: DeleteMailingListSubscriberInput!
    $connections: [ID!]!
  ) {
    deleteMailingListSubscriber(input: $input) {
      deletedMailingListSubscriberId @deleteEdge(connections: $connections)
    }
  }
`;

const fragment = graphql`
  fragment CompliancePageMailingListFragment on TrustCenter
  @argumentDefinitions(
    first: { type: Int, defaultValue: 20 }
    after: { type: CursorKey, defaultValue: null }
  )
  @refetchable(queryName: "CompliancePageMailingListQuery") {
    mailingList {
      id
      subscribers(
        first: $first
        after: $after
      ) @connection(key: "CompliancePageMailingList_subscribers") {
        __id
        pageInfo {
          hasNextPage
          endCursor
        }
        edges {
          node {
            id
            fullName
            email
            status
            createdAt
          }
        }
      }
    }
  }
`;

export function CompliancePageMailingList(props: {
  fragmentRef: CompliancePageMailingListFragment$key;
}) {
  const { fragmentRef } = props;
  const { __ } = useTranslate();

  const {
    data,
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<CompliancePageMailingListQuery, CompliancePageMailingListFragment$key>(
    fragment,
    fragmentRef,
  );

  const subscribers = data.mailingList?.subscribers;

  const [deleteSubscriber, isDeleting] = useMutationWithToasts<CompliancePageMailingListDeleteMutation>(
    deleteMutation,
    {
      successMessage: __("Subscriber removed successfully"),
      errorMessage: __("Failed to delete subscriber"),
    },
  );

  const handleDelete = (id: string) => {
    if (!subscribers) return;
    void deleteSubscriber({
      variables: {
        input: { id },
        connections: [subscribers.__id],
      },
    });
  };

  return (
    <>
      {!subscribers || subscribers.edges.length === 0
        ? (
            <Table>
              <Tbody>
                <Tr>
                  <Td className="text-center text-txt-tertiary py-8">
                    {__("No mailing list subscribers yet")}
                  </Td>
                </Tr>
              </Tbody>
            </Table>
          )
        : (
            <>
              <Table>
                <Thead>
                  <Tr>
                    <Th>{__("Name")}</Th>
                    <Th>{__("Email")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Subscribed on")}</Th>
                    <Th />
                  </Tr>
                </Thead>
                <Tbody>
                  {subscribers.edges.map(({ node: subscriber }) => (
                    <Tr key={subscriber.id}>
                      <Td>{subscriber.fullName}</Td>
                      <Td>{subscriber.email}</Td>
                      <Td>
                        <Badge
                          variant={subscriber.status === "CONFIRMED" ? "success" : "warning"}
                        >
                          {subscriber.status === "CONFIRMED" ? __("Confirmed") : __("Pending")}
                        </Badge>
                      </Td>
                      <Td className="text-txt-tertiary text-sm">
                        {new Date(subscriber.createdAt).toLocaleDateString()}
                      </Td>
                      <Td className="w-10">
                        <Button
                          variant="tertiary"
                          icon={IconTrashCan}
                          disabled={isDeleting}
                          onClick={() => handleDelete(subscriber.id)}
                          aria-label={__("Delete subscriber")}
                        />
                      </Td>
                    </Tr>
                  ))}
                </Tbody>
              </Table>
              {hasNext && (
                <Button
                  variant="tertiary"
                  onClick={() => loadNext(10)}
                  disabled={isLoadingNext}
                  className="mt-3 mx-auto"
                  icon={IconChevronDown}
                >
                  {isLoadingNext && <Spinner />}
                  {__("Show More")}
                </Button>
              )}
            </>
          )}
    </>
  );
}
