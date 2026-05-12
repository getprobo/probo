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
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useState } from "react";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageReferenceListFragment$key } from "#/__generated__/core/CompliancePageReferenceListFragment.graphql";
import type { CompliancePageReferenceListItemFragment$data } from "#/__generated__/core/CompliancePageReferenceListItemFragment.graphql";
import type { CompliancePageReferenceListQuery } from "#/__generated__/core/CompliancePageReferenceListQuery.graphql";
import { useUpdateTrustCenterReferenceRankMutation } from "#/hooks/graph/TrustCenterReferenceGraph";

import { CompliancePageReferenceListItem } from "./CompliancePageReferenceListItem";

const fragment = graphql`
  fragment CompliancePageReferenceListFragment on TrustCenter
  @refetchable(queryName: "CompliancePageReferenceListQuery")
  @argumentDefinitions (
    first: { type: Int defaultValue: 100 }
    after: { type: CursorKey defaultValue: null }
    order: { type: TrustCenterReferenceOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    references(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePageReferenceList_references", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          rank
          ...CompliancePageReferenceListItemFragment
        }
      }
    }
  }
`;

export function CompliancePageReferenceList(props: {
  fragmentRef: CompliancePageReferenceListFragment$key;
  onEdit: (r: CompliancePageReferenceListItemFragment$data, rank: number) => void;
}) {
  const { fragmentRef, onEdit } = props;

  const { __ } = useTranslate();

  const [{ references }, refetch] = useRefetchableFragment<
    CompliancePageReferenceListQuery,
    CompliancePageReferenceListFragment$key
  >(fragment, fragmentRef);
  const [updateRank] = useUpdateTrustCenterReferenceRankMutation();

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const handleDragStart = (index: number) => {
    setDraggedIndex(index);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex !== index) {
      setDragOverIndex(index);
    }
  };

  const handleDrop = async (targetIndex: number) => {
    if (draggedIndex === null || draggedIndex === targetIndex) {
      setDraggedIndex(null);
      setDragOverIndex(null);
      return;
    }

    const draggedRef = references.edges[draggedIndex];
    const targetRank = references.edges[targetIndex].node.rank;

    await updateRank({
      variables: {
        input: {
          id: draggedRef.node.id,
          rank: targetRank,
        },
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          return;
        }

        refetch({});
      },
    });

    setDraggedIndex(null);
    setDragOverIndex(null);
  };

  return (
    <>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("Description")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {references.edges.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No references available")}
              </Td>
            </Tr>
          )}
          {references.edges.map(({ node: reference }, index: number) => (
            <CompliancePageReferenceListItem
              key={reference.id}
              fragmentRef={reference}
              index={index}
              isDragging={draggedIndex === index}
              isDropTarget={dragOverIndex === index && draggedIndex !== index}
              onEdit={(r: CompliancePageReferenceListItemFragment$data) => onEdit(r, reference.rank)}
              connectionId={references.__id}
              onDragStart={() => handleDragStart(index)}
              onDragOver={e => handleDragOver(e, index)}
              onDrop={() => void handleDrop(index)}
            />
          ))}
        </Tbody>
      </Table>

      <p className="text-sm text-txt-tertiary">
        {__("Drag and drop references to change their displayed order")}
      </p>
    </>
  );
}
