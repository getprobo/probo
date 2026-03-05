import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconChevronDown, IconPencil, IconSend, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr, useConfirm } from "@probo/ui";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageUpdatesListDeleteMutation } from "#/__generated__/core/CompliancePageUpdatesListDeleteMutation.graphql";
import type { CompliancePageUpdatesListFragment$data, CompliancePageUpdatesListFragment$key } from "#/__generated__/core/CompliancePageUpdatesListFragment.graphql";
import type { CompliancePageUpdatesListQuery } from "#/__generated__/core/CompliancePageUpdatesListQuery.graphql";
import type { CompliancePageUpdatesListSendMutation } from "#/__generated__/core/CompliancePageUpdatesListSendMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const deleteMutation = graphql`
  mutation CompliancePageUpdatesListDeleteMutation(
    $input: DeleteMailingListUpdateInput!
    $connections: [ID!]!
  ) {
    deleteMailingListUpdate(input: $input) {
      deletedMailingListUpdateId @deleteEdge(connections: $connections)
    }
  }
`;

const sendMutation = graphql`
  mutation CompliancePageUpdatesListSendMutation($input: SendMailingListUpdateInput!) {
    sendMailingListUpdate(input: $input) {
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
  fragment CompliancePageUpdatesListFragment on MailingList
  @argumentDefinitions(
    first: { type: Int, defaultValue: 20 }
    after: { type: CursorKey, defaultValue: null }
  )
  @refetchable(queryName: "CompliancePageUpdatesListQuery") {
    updates(
      first: $first
      after: $after
    ) @connection(key: "CompliancePageUpdatesList_updates") {
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

export type UpdateNode = CompliancePageUpdatesListFragment$data["updates"]["edges"][number]["node"];

export function CompliancePageUpdatesList(props: {
  fragmentRef: CompliancePageUpdatesListFragment$key;
  onEdit: (update: UpdateNode) => void;
}) {
  const { fragmentRef, onEdit } = props;
  const { __ } = useTranslate();
  const confirm = useConfirm();

  const { data, hasNext, loadNext, isLoadingNext } = usePaginationFragment<
    CompliancePageUpdatesListQuery,
    CompliancePageUpdatesListFragment$key
  >(fragment, fragmentRef);

  const connection = data.updates;

  const [deleteUpdate, isDeleting]
    = useMutationWithToasts<CompliancePageUpdatesListDeleteMutation>(deleteMutation, {
      successMessage: __("Update deleted successfully"),
      errorMessage: __("Failed to delete update"),
    });

  const [sendUpdate, isSending]
    = useMutationWithToasts<CompliancePageUpdatesListSendMutation>(sendMutation, {
      successMessage: __("Update enqueued for delivery"),
      errorMessage: __("Failed to enqueue update for delivery"),
    });

  const handleDelete = (id: string) => {
    void deleteUpdate({
      variables: {
        input: { id },
        connections: [connection.__id],
      },
    });
  };

  const handleSend = (node: UpdateNode) => {
    confirm(
      () =>
        sendUpdate({
          variables: {
            input: {
              id: node.id,
            },
          },
        }),
      {
        title: __("Send this update to all subscribers?"),
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
                    {__("No updates yet")}
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
                        <Badge variant={node.status === "SENT" ? "success" : node.status === "DRAFT" ? "warning" : "info"}>
                          {node.status === "SENT" ? __("Sent") : node.status === "ENQUEUED" ? __("Queued") : node.status === "PROCESSING" ? __("Processing…") : __("Draft")}
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
                              onClick={() => handleSend(node)}
                              className="bg-green-600 text-white hover:bg-green-700 active:bg-green-800 shadow-none"
                              aria-label={__("Send")}
                            >
                              {__("Send")}
                            </Button>
                          )}
                          <Button
                            variant="tertiary"
                            icon={IconPencil}
                            onClick={() => onEdit(node)}
                            aria-label={__("Edit update")}
                          />
                          <Button
                            variant="tertiary"
                            icon={IconTrashCan}
                            disabled={isDeleting}
                            onClick={() => handleDelete(node.id)}
                            aria-label={__("Delete update")}
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
