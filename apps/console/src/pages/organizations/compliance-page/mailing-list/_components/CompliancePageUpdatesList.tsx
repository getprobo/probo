// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconChevronDown, IconPageTextLine, IconPencil, IconSend, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr, useDialogRef } from "@probo/ui";
import { useState } from "react";
import { usePaginationFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageUpdatesListDeleteMutation } from "#/__generated__/core/CompliancePageUpdatesListDeleteMutation.graphql";
import type { CompliancePageUpdatesListFragment$data, CompliancePageUpdatesListFragment$key } from "#/__generated__/core/CompliancePageUpdatesListFragment.graphql";
import type { CompliancePageUpdatesListQuery } from "#/__generated__/core/CompliancePageUpdatesListQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { SendUpdateDialog } from "./SendUpdateDialog";

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
          # eslint-disable-next-line relay/unused-fields
          body
          status
          createdAt
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

  const sendDialogRef = useDialogRef();
  const [updateToSend, setUpdateToSend] = useState<UpdateNode | null>(null);

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

  const handleDelete = (id: string) => {
    void deleteUpdate({
      variables: {
        input: { id },
        connections: [connection.__id],
      },
    });
  };

  const handleSend = (node: UpdateNode) => {
    setUpdateToSend(node);
    sendDialogRef.current?.open();
  };

  return (
    <>
      <SendUpdateDialog ref={sendDialogRef} update={updateToSend} />
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
                        <div className="flex gap-2 justify-end">
                          {node.status === "DRAFT" && (
                            <Button
                              variant="secondary"
                              icon={IconSend}
                              onClick={() => handleSend(node)}
                              aria-label={__("Send")}
                            >
                              {__("Send")}
                            </Button>
                          )}
                          <Button
                            variant="secondary"
                            icon={node.status === "DRAFT" ? IconPencil : IconPageTextLine}
                            onClick={() => onEdit(node)}
                            aria-label={node.status === "DRAFT" ? __("Edit update") : __("View update")}
                          />
                          <Button
                            variant="danger"
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
