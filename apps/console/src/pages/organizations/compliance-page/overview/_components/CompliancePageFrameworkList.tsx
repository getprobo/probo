import { useTranslate } from "@probo/i18n";
import { Badge, Field, FrameworkLogo, Option, Table, Tbody, Td, Th, Thead, Tr } from "@probo/ui";
import { useCallback, useState, useTransition } from "react";
import { useRefetchableFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";

import type { CompliancePageFrameworkList_compliancePageFragment$key } from "#/__generated__/core/CompliancePageFrameworkList_compliancePageFragment.graphql";
import type { CompliancePageFrameworkList_compliancePageFragment$data } from "#/__generated__/core/CompliancePageFrameworkList_compliancePageFragment.graphql";
import type { CompliancePageFrameworkList_compliancePageRefetchQuery } from "#/__generated__/core/CompliancePageFrameworkList_compliancePageRefetchQuery.graphql";
import type { CompliancePageFrameworkList_createMutation } from "#/__generated__/core/CompliancePageFrameworkList_createMutation.graphql";
import type { CompliancePageFrameworkList_deleteMutation } from "#/__generated__/core/CompliancePageFrameworkList_deleteMutation.graphql";
import type { CompliancePageFrameworkList_updateRankMutation } from "#/__generated__/core/CompliancePageFrameworkList_updateRankMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const compliancePageFragment = graphql`
  fragment CompliancePageFrameworkList_compliancePageFragment on TrustCenter
  @refetchable(queryName: "CompliancePageFrameworkList_compliancePageRefetchQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 100 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceFrameworkOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    id
    canUpdate: permission(action: "core:trust-center:update")
    complianceFrameworks(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePageFrameworkList_complianceFrameworks", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          rank
          visibility
          framework {
            id
            name
            lightLogoURL
            darkLogoURL
          }
        }
      }
    }
  }
`;

const createMutation = graphql`
  mutation CompliancePageFrameworkList_createMutation($input: CreateComplianceFrameworkInput!) {
    createComplianceFramework(input: $input) {
      complianceFrameworkEdge {
        node {
          id
        }
      }
    }
  }
`;

const updateRankMutation = graphql`
  mutation CompliancePageFrameworkList_updateRankMutation($input: UpdateComplianceFrameworkInput!) {
    updateComplianceFramework(input: $input) {
      complianceFramework {
        id
        rank
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation CompliancePageFrameworkList_deleteMutation($input: DeleteComplianceFrameworkInput!) {
    deleteComplianceFramework(input: $input) {
      deletedComplianceFrameworkId
    }
  }
`;

type Edge = CompliancePageFrameworkList_compliancePageFragment$data["complianceFrameworks"]["edges"][number];

function CompliancePageFrameworkListItem(props: {
  edge: Edge;
  compliancePage: CompliancePageFrameworkList_compliancePageFragment$data;
  draggedCfId: string | null;
  dragOverCfId: string | null;
  onDragStart: (cfId: string) => void;
  onDragEnd: () => void;
  onDragOver: (e: React.DragEvent, cfId: string) => void;
  onDrop: (cfId: string) => void;
  onRefetch: () => void;
}) {
  const { edge, draggedCfId, dragOverCfId, onDragStart, onDragEnd, onDragOver, onDrop, onRefetch } = props;
  const { __ } = useTranslate();
  const [isMouseDown, setIsMouseDown] = useState(false);

  const compliancePage = props.compliancePage;
  const { id, visibility, framework } = edge.node;

  const isPublic = visibility === "PUBLIC";
  const canDrag = isPublic && compliancePage.canUpdate;

  const isDragging = draggedCfId === id;
  const isDropTarget = dragOverCfId === id && draggedCfId !== id;

  const [createComplianceFramework, isCreating] = useMutationWithToasts<CompliancePageFrameworkList_createMutation>(
    createMutation,
    {
      successMessage: __("Framework visibility updated successfully."),
      errorMessage: __("Failed to update framework visibility"),
    },
  );

  const [deleteComplianceFramework, isDeleting] = useMutationWithToasts<CompliancePageFrameworkList_deleteMutation>(
    deleteMutation,
    {
      successMessage: __("Framework visibility updated successfully."),
      errorMessage: __("Failed to update framework visibility"),
    },
  );

  const isLoading = isCreating || isDeleting;

  const handleVisibilityChange = useCallback(
    async (value: string) => {
      if (value === "PUBLIC" && !isPublic) {
        await createComplianceFramework({
          variables: {
            input: {
              trustCenterId: compliancePage.id,
              frameworkId: framework.id,
            },
          },
          onCompleted: (_, errors) => {
            if (!errors?.length) {
              onRefetch();
            }
          },
        });
      } else if (value === "NONE" && isPublic) {
        await deleteComplianceFramework({
          variables: {
            input: { id },
          },
          onCompleted: (_, errors) => {
            if (!errors?.length) {
              onRefetch();
            }
          },
        });
      }
    },
    [compliancePage.id, framework.id, id, isPublic, onRefetch, createComplianceFramework, deleteComplianceFramework],
  );

  const visibilityOptions = [
    { value: "PUBLIC", label: __("Public"), variant: "success" as const },
    { value: "NONE", label: __("None"), variant: "neutral" as const },
  ];

  const rowClassName = [
    canDrag && isDragging && "opacity-50 cursor-grabbing",
    canDrag && !isDragging && !isMouseDown && "cursor-grab",
    canDrag && !isDragging && isMouseDown && "cursor-grabbing",
    isDropTarget && "!bg-primary-50 border-y-2 border-primary-500",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <Tr
      draggable={canDrag}
      onDragStart={canDrag ? () => onDragStart(id) : undefined}
      onDragEnd={canDrag ? onDragEnd : undefined}
      onDragOver={canDrag ? e => onDragOver(e, id) : undefined}
      onDrop={
        canDrag
          ? (e) => {
              e.preventDefault();
              onDrop(id);
            }
          : undefined
      }
      onMouseDown={canDrag ? () => setIsMouseDown(true) : undefined}
      onMouseUp={canDrag ? () => setIsMouseDown(false) : undefined}
      onMouseLeave={canDrag ? () => setIsMouseDown(false) : undefined}
      className={rowClassName}
    >
      <Td>
        <div className="flex items-center gap-3">
          <FrameworkLogo
            className="size-8"
            lightLogoURL={framework.lightLogoURL}
            darkLogoURL={framework.darkLogoURL}
            name={framework.name}
          />
          {framework.name}
        </div>
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={visibility}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={isLoading || !compliancePage.canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
    </Tr>
  );
}

export function CompliancePageFrameworkList(props: {
  compliancePageRef: CompliancePageFrameworkList_compliancePageFragment$key;
}) {
  const { __ } = useTranslate();
  const [, startTransition] = useTransition();

  const [compliancePage, refetch] = useRefetchableFragment<
    CompliancePageFrameworkList_compliancePageRefetchQuery,
    CompliancePageFrameworkList_compliancePageFragment$key
  >(compliancePageFragment, props.compliancePageRef);

  const [updateRank] = useMutationWithToasts<CompliancePageFrameworkList_updateRankMutation>(
    updateRankMutation,
    {
      successMessage: __("Order updated successfully"),
      errorMessage: __("Failed to update order"),
    },
  );

  const [draggedCfId, setDraggedCfId] = useState<string | null>(null);
  const [dragOverCfId, setDragOverCfId] = useState<string | null>(null);

  const allEdges = compliancePage.complianceFrameworks.edges;
  const publicEdges = allEdges.filter(e => e.node.visibility === "PUBLIC");

  const handleDragStart = (cfId: string) => {
    setDraggedCfId(cfId);
  };

  const handleDragEnd = () => {
    setDraggedCfId(null);
    setDragOverCfId(null);
  };

  const handleDragOver = (e: React.DragEvent, cfId: string) => {
    e.preventDefault();
    if (draggedCfId !== cfId) {
      setDragOverCfId(cfId);
    }
  };

  const handleDrop = async (targetCfId: string) => {
    if (!draggedCfId || draggedCfId === targetCfId) {
      setDraggedCfId(null);
      setDragOverCfId(null);
      return;
    }

    const targetEdge = publicEdges.find(e => e.node.id === targetCfId);
    if (!targetEdge) {
      setDraggedCfId(null);
      setDragOverCfId(null);
      return;
    }

    const draggedId = draggedCfId;

    await updateRank({
      variables: {
        input: {
          id: draggedId,
          rank: targetEdge.node.rank,
        },
      },
      updater: (store) => {
        const trustCenterRecord = store.get(compliancePage.id);
        if (!trustCenterRecord) return;

        const connection = ConnectionHandler.getConnection(
          trustCenterRecord,
          "CompliancePageFrameworkList_complianceFrameworks",
          { orderBy: { field: "RANK", direction: "ASC" } },
        );
        if (!connection) return;

        const edges = connection.getLinkedRecords("edges");
        if (!edges) return;

        const fromIdx = edges.findIndex(e => e.getLinkedRecord("node")?.getDataID() === draggedId);
        const toIdx = edges.findIndex(e => e.getLinkedRecord("node")?.getDataID() === targetCfId);
        if (fromIdx === -1 || toIdx === -1) return;

        const reordered = [...edges];
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

    setDraggedCfId(null);
    setDragOverCfId(null);
  };

  const hasMultiplePublic = publicEdges.length > 1;

  return (
    <>
      <Table>
        <Thead>
          <Tr>
            <Th>{__("Framework")}</Th>
            <Th>{__("Visibility")}</Th>
          </Tr>
        </Thead>
        <Tbody>
          {allEdges.length === 0 && (
            <Tr>
              <Td colSpan={2} className="text-center text-txt-secondary">
                {__("No frameworks available")}
              </Td>
            </Tr>
          )}
          {allEdges.map(edge => (
            <CompliancePageFrameworkListItem
              key={edge.node.id}
              edge={edge}
              compliancePage={compliancePage}
              draggedCfId={draggedCfId}
              dragOverCfId={dragOverCfId}
              onDragStart={handleDragStart}
              onDragEnd={handleDragEnd}
              onDragOver={handleDragOver}
              onDrop={cfId => void handleDrop(cfId)}
              onRefetch={() => startTransition(() => { refetch({}, { fetchPolicy: "store-and-network" }); })}
            />
          ))}
        </Tbody>
      </Table>

      {hasMultiplePublic && compliancePage.canUpdate && (
        <p className="text-sm text-txt-tertiary">
          {__("Drag and drop public frameworks to change their displayed order")}
        </p>
      )}
    </>
  );
}
