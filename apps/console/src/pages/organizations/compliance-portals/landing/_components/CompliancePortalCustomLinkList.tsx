// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Button, IconChevronRight, IconPlusLarge } from "@probo/ui";
import { useCallback, useRef, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalCustomLinkList_compliancePortalFragment$key } from "#/__generated__/core/CompliancePortalCustomLinkList_compliancePortalFragment.graphql";
import type { CompliancePortalCustomLinkList_compliancePortalRefetchQuery } from "#/__generated__/core/CompliancePortalCustomLinkList_compliancePortalRefetchQuery.graphql";
import type { CompliancePortalCustomLinkList_updateRankMutation } from "#/__generated__/core/CompliancePortalCustomLinkList_updateRankMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { CompliancePortalCustomLinkDialog, type CompliancePortalCustomLinkDialogRef } from "./CompliancePortalCustomLinkDialog";
import { CompliancePortalCustomLinkListItem } from "./CompliancePortalCustomLinkListItem";

const compliancePortalFragment = graphql`
  fragment CompliancePortalCustomLinkList_compliancePortalFragment on CompliancePortal
  @refetchable(queryName: "CompliancePortalCustomLinkList_compliancePortalRefetchQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 500 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceCustomLinkOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    id
    canUpdate: permission(action: "compliance-portal:portal:update")
    customLinks(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePortalCustomLinkList_customLinks", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          name
          url
          rank
          ...CompliancePortalCustomLinkListItem_customLink
          ...CompliancePortalCustomLinkDialog_customLink
        }
      }
    }
  }
`;

const updateRankMutation = graphql`
  mutation CompliancePortalCustomLinkList_updateRankMutation($input: UpdateComplianceCustomLinkInput!) {
    updateComplianceCustomLink(input: $input) {
      complianceCustomLink {
        id
        rank
      }
    }
  }
`;

export interface CompliancePortalCustomLinkListProps {
  compliancePortalRef: CompliancePortalCustomLinkList_compliancePortalFragment$key;
}

export function CompliancePortalCustomLinkList(props: CompliancePortalCustomLinkListProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const [, startTransition] = useTransition();
  const dialogRef = useRef<CompliancePortalCustomLinkDialogRef>(null);

  const [compliancePortal, refetch] = useRefetchableFragment<
    CompliancePortalCustomLinkList_compliancePortalRefetchQuery,
    CompliancePortalCustomLinkList_compliancePortalFragment$key
  >(compliancePortalFragment, props.compliancePortalRef);

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const [updateRank] = useMutation<CompliancePortalCustomLinkList_updateRankMutation>(
    updateRankMutation,
    { successMessage: t("externalUrls.messages.orderUpdated"), errorToast: t("externalUrls.errors.orderUpdate") },
  );

  const edges = compliancePortal.customLinks.edges;
  const readOnly = !compliancePortal.canUpdate;
  const connectionId = compliancePortal.customLinks.__id;
  const hasLinks = edges.length > 0;

  const handleCreate = () => {
    dialogRef.current?.openCreate(compliancePortal.id, connectionId);
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    if (draggedIndex !== index) setDragOverIndex(index);
  };

  const handleDrop = useCallback(
    async (targetIndex: number) => {
      if (draggedIndex === null || draggedIndex === targetIndex) {
        setDraggedIndex(null);
        setDragOverIndex(null);
        return;
      }

      const draggedEdge = edges[draggedIndex];
      const targetRank = edges[targetIndex].node.rank;
      const draggedId = draggedEdge.node.id;

      await updateRank({
        variables: {
          input: {
            id: draggedId,
            name: draggedEdge.node.name,
            url: draggedEdge.node.url,
            rank: targetRank,
          },
        },
        updater: (store) => {
          const connection = store.get(connectionId);
          if (!connection) return;
          const storeEdges = connection.getLinkedRecords("edges");
          if (!storeEdges) return;
          const fromIdx = storeEdges.findIndex(e => e.getLinkedRecord("node")?.getDataID() === draggedId);
          const toIdx = storeEdges.findIndex(e => e.getLinkedRecord("node")?.getDataID() === edges[targetIndex].node.id);
          if (fromIdx === -1 || toIdx === -1) return;
          const reordered = [...storeEdges];
          const [moved] = reordered.splice(fromIdx, 1);
          reordered.splice(toIdx, 0, moved);
          connection.setLinkedRecords(reordered, "edges");
        },
        onCompleted: (_, errors) => {
          startTransition(() => {
            refetch({}, { fetchPolicy: errors?.length ? "network-only" : "store-and-network" });
          });
        },
      });

      setDraggedIndex(null);
      setDragOverIndex(null);
    },
    [draggedIndex, edges, connectionId, updateRank, refetch, startTransition],
  );

  return (
    <div className="space-y-3">
      {!readOnly && hasLinks && (
        <div className="flex justify-end">
          <Button icon={IconPlusLarge} onClick={handleCreate}>
            {t("externalUrls.actions.add")}
          </Button>
        </div>
      )}

      {edges.map(({ node }, index) => (
        <CompliancePortalCustomLinkListItem
          key={node.id}
          customLinkKey={node}
          connectionId={connectionId}
          readOnly={readOnly}
          isDragging={draggedIndex === index}
          isDropTarget={dragOverIndex === index && draggedIndex !== index}
          onDragStart={() => setDraggedIndex(index)}
          onDragOver={e => handleDragOver(e, index)}
          onDrop={() => void handleDrop(index)}
          onDragEnd={() => {
            setDraggedIndex(null);
            setDragOverIndex(null);
          }}
          onEdit={() => dialogRef.current?.openEdit(node)}
        />
      ))}

      {!hasLinks && !readOnly && (
        <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border-solid px-4 py-8">
          <p className="max-w-md text-center text-sm text-txt-tertiary">
            {t("externalUrls.description")}
          </p>
          <Button iconAfter={IconChevronRight} onClick={handleCreate}>
            {t("externalUrls.actions.add")}
          </Button>
        </div>
      )}

      {edges.length > 1 && !readOnly && (
        <p className="text-sm text-txt-tertiary">
          {t("externalUrls.dragHint")}
        </p>
      )}

      <CompliancePortalCustomLinkDialog ref={dialogRef} />
    </div>
  );
}
