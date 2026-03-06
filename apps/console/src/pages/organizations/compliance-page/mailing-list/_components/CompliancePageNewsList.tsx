import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconChevronDown, IconPencil, IconSend, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr, useConfirm } from "@probo/ui";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNewsListDeleteMutation } from "#/__generated__/core/CompliancePageNewsListDeleteMutation.graphql";
import type { CompliancePageNewsListFragment$key } from "#/__generated__/core/CompliancePageNewsListFragment.graphql";
import type { CompliancePageNewsListQuery } from "#/__generated__/core/CompliancePageNewsListQuery.graphql";
import type { CompliancePageNewsListSendMutation } from "#/__generated__/core/CompliancePageNewsListSendMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteMutation = graphql`
  mutation CompliancePageNewsListDeleteMutation(
    $input: DeleteMailingListUpdateInput!
    $connections: [ID!]!
  ) {
    deleteMailingListUpdate(input: $input) {
      deletedMailingListUpdateId @deleteEdge(connections: $connections)
    }
  }
`;

const sendMutation = graphql`
  mutation CompliancePageNewsListSendMutation($input: UpdateMailingListUpdateInput!) {
    updateMailingListUpdate(input: $input) {
      mailingListUpdate {
        id
        title
        body
        status
        updatedAt
      }
    }
  }
`;

const fragment = graphql`
  fragment CompliancePageNewsListFragment on TrustCenter
  @argumentDefinitions(
    first: { type: Int, defaultValue: 20 }
    after: { type: CursorKey, defaultValue: null }
  )
  @refetchable(queryName: "CompliancePageNewsListQuery") {
    mailingListUpdates(
      first: $first
      after: $after
    ) @connection(key: "CompliancePageNewsList_mailingListUpdates") {
      __id
      pageInfo {
        hasNextPage
        endCursor
      }
      edges {
        node {
          id
          title
          body
          status
          createdAt
          updatedAt
        }
      }
    }
  }
`;

type NewsNode = {
  id: string;
  title: string;
  body: string;
  status: "DRAFT" | "SENT";
  createdAt: string;
  updatedAt: string;
};

export function CompliancePageNewsList(props: {
  fragmentRef: CompliancePageNewsListFragment$key;
  onEdit: (news: NewsNode) => void;
}) {
  const { fragmentRef, onEdit } = props;
  const { __ } = useTranslate();
  const confirm = useConfirm();

  const { data, hasNext, loadNext, isLoadingNext } = usePaginationFragment<
    CompliancePageNewsListQuery,
    CompliancePageNewsListFragment$key
  >(fragment, fragmentRef);

  const connection = data.mailingListUpdates;

  const [deleteNews, isDeleting]
    = useMutationWithToasts<CompliancePageNewsListDeleteMutation>(deleteMutation, {
      successMessage: __("News deleted successfully"),
      errorMessage: __("Failed to delete news"),
    });

  const [sendNews, isSending]
    = useMutationWithToasts<CompliancePageNewsListSendMutation>(sendMutation, {
      successMessage: __("News marked as sent"),
      errorMessage: __("Failed to send news"),
    });

  const handleDelete = (id: string) => {
    void deleteNews({
      variables: {
        input: { id },
        connections: [connection.__id],
      },
    });
  };

  const handleSend = (node: NewsNode) => {
    confirm(
      () =>
        sendNews({
          variables: {
            input: {
              id: node.id,
              title: node.title,
              body: node.body,
              status: "SENT",
            },
          },
        }),
      {
        title: __("Send this news to all subscribers?"),
        message: `"${node.title}" — ${node.body.length > 120 ? node.body.slice(0, 120) + "…" : node.body}`,
        label: __("Send"),
        variant: "primary",
      },
    );
  };

  return (
    <>
      {connection.edges.length === 0
        ? (
            <Table>
              <Tbody>
                <Tr>
                  <Td className="text-center text-txt-tertiary py-8">
                    {__("No news yet")}
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
                    <Th>{__("Title")}</Th>
                    <Th>{__("Status")}</Th>
                    <Th>{__("Created")}</Th>
                    <Th />
                  </Tr>
                </Thead>
                <Tbody>
                  {connection.edges.map(({ node }) => (
                    <Tr key={node.id}>
                      <Td>{node.title}</Td>
                      <Td>
                        <Badge variant={node.status === "SENT" ? "success" : "warning"}>
                          {node.status === "SENT" ? __("Sent") : __("Draft")}
                        </Badge>
                      </Td>
                      <Td className="text-txt-tertiary text-sm">
                        {new Date(node.createdAt).toLocaleDateString()}
                      </Td>
                      <Td className="w-auto">
                        <div className="flex gap-1 justify-end">
                          {node.status === "DRAFT" && (
                            <Button
                              icon={IconSend}
                              disabled={isSending}
                              onClick={() => handleSend(node as NewsNode)}
                              className="bg-green-600 text-white hover:bg-green-700 active:bg-green-800 shadow-none"
                              aria-label={__("Mark as sent")}
                            >
                              {__("Send")}
                            </Button>
                          )}
                          <Button
                            variant="tertiary"
                            icon={IconPencil}
                            onClick={() => onEdit(node as NewsNode)}
                            aria-label={__("Edit news")}
                          />
                          <Button
                            variant="tertiary"
                            icon={IconTrashCan}
                            disabled={isDeleting}
                            onClick={() => handleDelete(node.id)}
                            aria-label={__("Delete news")}
                          />
                        </div>
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
