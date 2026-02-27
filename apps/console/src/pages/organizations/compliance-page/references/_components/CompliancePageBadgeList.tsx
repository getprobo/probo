import { useTranslate } from "@probo/i18n";
import { Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { Fragment, startTransition } from "react";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageBadgeListFragment$key } from "#/__generated__/core/CompliancePageBadgeListFragment.graphql";
import type { CompliancePageBadgeListItemFragment$data } from "#/__generated__/core/CompliancePageBadgeListItemFragment.graphql";
import type { CompliancePageBadgeListQuery } from "#/__generated__/core/CompliancePageBadgeListQuery.graphql";
import { useUpdateComplianceBadgeRankMutation } from "#/hooks/graph/ComplianceBadgeGraph";
import { useDragReorder } from "#/hooks/useDragReorder";

import { CompliancePageBadgeListItem } from "./CompliancePageBadgeListItem";

const fragment = graphql`
  fragment CompliancePageBadgeListFragment on TrustCenter
  @refetchable(queryName: "CompliancePageBadgeListQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 100 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceBadgeOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    complianceBadges(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePageBadgeList_complianceBadges", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          rank
          ...CompliancePageBadgeListItemFragment
        }
      }
    }
  }
`;

function GhostBadgeRow({
  name,
  iconUrl,
  onDragOver,
  onDrop,
}: {
  name: string;
  iconUrl: string;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: (e: React.DragEvent) => void;
}) {
  return (
    <Tr className="opacity-50 bg-primary-50 cursor-default" onDragOver={onDragOver} onDrop={onDrop}>
      <Td width={32} className="pr-3 text-txt-tertiary">
        <svg width="10" height="14" viewBox="0 0 10 14" fill="currentColor" aria-hidden>
          <circle cx="2.5" cy="2" r="1.5" />
          <circle cx="7.5" cy="2" r="1.5" />
          <circle cx="2.5" cy="7" r="1.5" />
          <circle cx="7.5" cy="7" r="1.5" />
          <circle cx="2.5" cy="12" r="1.5" />
          <circle cx="7.5" cy="12" r="1.5" />
        </svg>
      </Td>
      <Td>
        <div className="flex items-center gap-4">
          <div className="size-8 flex-shrink-0 rounded bg-white flex items-center justify-center border border-border-low overflow-hidden">
            <img src={iconUrl} alt={name} className="size-7 object-contain" />
          </div>
          <span className="font-medium">{name}</span>
        </div>
      </Td>
      <Td></Td>
    </Tr>
  );
}

export function CompliancePageBadgeList(props: {
  fragmentRef: CompliancePageBadgeListFragment$key;
  onEdit: (b: CompliancePageBadgeListItemFragment$data) => void;
}) {
  const { fragmentRef, onEdit } = props;

  const { __ } = useTranslate();

  const [{ complianceBadges }, refetch] = useRefetchableFragment<
    CompliancePageBadgeListQuery,
    CompliancePageBadgeListFragment$key
  >(fragment, fragmentRef);
  const [updateRank] = useUpdateComplianceBadgeRankMutation();

  const {
    draggedIndex,
    dropIndex,
    setDropIndex,
    draggedData,
    isDragging,
    handleDragStart,
    handleDragEnter,
    handleDragOver,
    handleDragEnd,
    handleDrop,
  } = useDragReorder<{ name: string; iconUrl: string }>({
    onDrop: (fromIndex, targetIndex) => {
      const draggedId = complianceBadges.edges[fromIndex].node.id;
      const targetRank = complianceBadges.edges[targetIndex].node.rank;
      void (async () => {
        try {
          await updateRank({ variables: { input: { id: draggedId, rank: targetRank } } });
        } catch {
          // error toast handled by useMutationWithToasts
        } finally {
          startTransition(() => {
            refetch({}, { fetchPolicy: "network-only" });
          });
        }
      })();
    },
  });

  return (
    <>
      <Table>
        <Thead>
          <Tr>
            <Th width={32}></Th>
            <Th>{__("Badge")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {complianceBadges.edges.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No badges yet")}
              </Td>
            </Tr>
          )}
          {complianceBadges.edges.map(({ node: badge }, index: number) => (
            <Fragment key={badge.id}>
              {isDragging && dropIndex === index && draggedData && (
                <GhostBadgeRow
                  name={draggedData.name}
                  iconUrl={draggedData.iconUrl}
                  onDragOver={handleDragOver}
                  onDrop={e => void handleDrop(e, index)}
                />
              )}
              <CompliancePageBadgeListItem
                fragmentRef={badge}
                isDragging={draggedIndex === index}
                onEdit={(b: CompliancePageBadgeListItemFragment$data) => onEdit(b)}
                connectionId={complianceBadges.__id}
                onDragStart={data => handleDragStart(index, data)}
                onDragEnter={e => handleDragEnter(e, index)}
                onDragOver={handleDragOver}
                onDragEnd={handleDragEnd}
                onDrop={e => void handleDrop(e, index)}
              />
            </Fragment>
          ))}
          {isDragging && dropIndex === complianceBadges.edges.length && draggedData && (
            <GhostBadgeRow
              name={draggedData.name}
              iconUrl={draggedData.iconUrl}
              onDragOver={handleDragOver}
              onDrop={e => void handleDrop(e, complianceBadges.edges.length - 1)}
            />
          )}
          {isDragging && dropIndex !== complianceBadges.edges.length && (
            <tr
              className="h-2"
              onDragEnter={(e) => {
                e.preventDefault();
                setDropIndex(complianceBadges.edges.length);
              }}
              onDragOver={handleDragOver}
            />
          )}
        </Tbody>
      </Table>

      <p className="text-xs text-txt-secondary flex items-center gap-1.5">
        <span aria-hidden="true">⠿</span>
        {__("Drag rows to reorder badges")}
      </p>
    </>
  );
}
