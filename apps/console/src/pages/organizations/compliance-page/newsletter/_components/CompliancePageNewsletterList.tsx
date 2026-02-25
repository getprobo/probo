import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, IconChevronDown, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr, useToast } from "@probo/ui";
import { useMutation, usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNewsletterListDeleteMutation } from "#/__generated__/core/CompliancePageNewsletterListDeleteMutation.graphql";
import type { CompliancePageNewsletterListFragment$key } from "#/__generated__/core/CompliancePageNewsletterListFragment.graphql";
import type { CompliancePageNewsletterListQuery } from "#/__generated__/core/CompliancePageNewsletterListQuery.graphql";

const deleteMutation = graphql`
  mutation CompliancePageNewsletterListDeleteMutation(
    $input: DeleteNewsletterSubscriberInput!
    $connections: [ID!]!
  ) {
    deleteNewsletterSubscriber(input: $input) {
      deletedNewsletterSubscriberId @deleteEdge(connections: $connections)
    }
  }
`;

const fragment = graphql`
  fragment CompliancePageNewsletterListFragment on TrustCenter
  @argumentDefinitions(
    first: { type: Int, defaultValue: 20 }
    after: { type: CursorKey, defaultValue: null }
  )
  @refetchable(queryName: "CompliancePageNewsletterListQuery") {
    newsletterSubscribers(
      first: $first
      after: $after
    ) @connection(key: "CompliancePageNewsletterList_newsletterSubscribers") {
      __id
      pageInfo {
        hasNextPage
        endCursor
      }
      edges {
        node {
          id
          email
          createdAt
        }
      }
    }
  }
`;

export function CompliancePageNewsletterList(props: {
  fragmentRef: CompliancePageNewsletterListFragment$key;
}) {
  const { fragmentRef } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();

  const {
    data: { newsletterSubscribers },
    hasNext,
    loadNext,
    isLoadingNext,
  } = usePaginationFragment<CompliancePageNewsletterListQuery, CompliancePageNewsletterListFragment$key>(
    fragment,
    fragmentRef,
  );

  const [deleteSubscriber, isDeleting] = useMutation<CompliancePageNewsletterListDeleteMutation>(deleteMutation);

  const handleDelete = (id: string) => {
    deleteSubscriber({
      variables: {
        input: { id },
        connections: [newsletterSubscribers.__id],
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot delete subscriber"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Deleted"),
          description: __("Subscriber removed successfully."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot delete subscriber"),
          variant: "error",
        });
      },
    });
  };

  if (newsletterSubscribers.edges.length === 0) {
    return (
      <Table>
        <Tbody>
          <Tr>
            <Td className="text-center text-txt-tertiary py-8">
              {__("No newsletter subscribers yet")}
            </Td>
          </Tr>
        </Tbody>
      </Table>
    );
  }

  return (
    <>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Email")}</Th>
            <Th>{__("Subscribed on")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {newsletterSubscribers.edges.map(({ node: subscriber }) => (
            <Tr key={subscriber.id}>
              <Td>{subscriber.email}</Td>
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
          onClick={() => loadNext(20)}
          disabled={isLoadingNext}
          className="mt-3 mx-auto"
          icon={IconChevronDown}
        >
          {isLoadingNext && <Spinner />}
          {__("Show More")}
        </Button>
      )}
    </>
  );
}
