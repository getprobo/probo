import { detectSocialName, safeOpenUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconArrowLink,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  SocialIcon,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useCallback, useImperativeHandle, useRef, useState, useTransition } from "react";
import { useRefetchableFragment } from "react-relay";
import { ConnectionHandler, graphql } from "relay-runtime";
import { z } from "zod";

import type { CompliancePageExternalUrlsSection_createMutation } from "#/__generated__/core/CompliancePageExternalUrlsSection_createMutation.graphql";
import type { CompliancePageExternalUrlsSection_deleteMutation } from "#/__generated__/core/CompliancePageExternalUrlsSection_deleteMutation.graphql";
import type { CompliancePageExternalUrlsSection_trustCenterFragment$key } from "#/__generated__/core/CompliancePageExternalUrlsSection_trustCenterFragment.graphql";
import type { CompliancePageExternalUrlsSection_trustCenterRefetchQuery } from "#/__generated__/core/CompliancePageExternalUrlsSection_trustCenterRefetchQuery.graphql";
import type { CompliancePageExternalUrlsSection_updateMutation } from "#/__generated__/core/CompliancePageExternalUrlsSection_updateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const trustCenterFragment = graphql`
  fragment CompliancePageExternalUrlsSection_trustCenterFragment on TrustCenter
  @refetchable(queryName: "CompliancePageExternalUrlsSection_trustCenterRefetchQuery")
  @argumentDefinitions(
    first: { type: Int, defaultValue: 100 }
    after: { type: CursorKey, defaultValue: null }
    order: { type: ComplianceExternalURLOrder, defaultValue: { field: RANK, direction: ASC } }
  ) {
    id
    canUpdate: permission(action: "core:trust-center:update")
    externalUrls(first: $first, after: $after, orderBy: $order)
    @connection(key: "CompliancePageExternalUrlsSection_externalUrls", filters: ["orderBy"]) {
      __id
      edges {
        node {
          id
          name
          url
          rank
        }
      }
    }
  }
`;

const createMutation = graphql`
  mutation CompliancePageExternalUrlsSection_createMutation($input: CreateComplianceExternalURLInput!) {
    createComplianceExternalURL(input: $input) {
      complianceExternalUrlEdge {
        node {
          id
          name
          url
          rank
        }
      }
    }
  }
`;

const updateMutation = graphql`
  mutation CompliancePageExternalUrlsSection_updateMutation($input: UpdateComplianceExternalURLInput!) {
    updateComplianceExternalURL(input: $input) {
      complianceExternalUrl {
        id
        name
        url
        rank
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation CompliancePageExternalUrlsSection_deleteMutation($input: DeleteComplianceExternalURLInput!) {
    deleteComplianceExternalURL(input: $input) {
      deletedComplianceExternalUrlId
    }
  }
`;

const urlSchema = z.object({
  name: z.string().min(1, "Name is required"),
  url: z.string().url("Please enter a valid URL"),
});

type UrlFormData = z.infer<typeof urlSchema>;

type UrlNode = { id: string; name: string; url: string; rank: number };

type ExternalUrlDialogRef = {
  openCreate: (trustCenterId: string, connectionId: string) => void;
  openEdit: (node: UrlNode) => void;
};

const ExternalUrlDialog = forwardRef<ExternalUrlDialogRef>(
  function ExternalUrlDialog(_, ref) {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [mode, setMode] = useState<"create" | "edit">("create");
    const [trustCenterId, setTrustCenterId] = useState("");
    const [connectionId, setConnectionId] = useState("");
    const [editNode, setEditNode] = useState<UrlNode | null>(null);

    const [create, isCreating] = useMutationWithToasts<CompliancePageExternalUrlsSection_createMutation>(
      createMutation,
      { successMessage: __("Link added successfully."), errorMessage: __("Failed to add link.") },
    );

    const [update, isUpdating] = useMutationWithToasts<CompliancePageExternalUrlsSection_updateMutation>(
      updateMutation,
      { successMessage: __("Link updated successfully."), errorMessage: __("Failed to update link.") },
    );

    const { register, handleSubmit, formState: { errors }, reset, setValue, watch } = useFormWithSchema(urlSchema, {
      defaultValues: { name: "", url: "" },
    });

    const [nameAutoDetected, setNameAutoDetected] = useState(false);

    const handleUrlChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const url = e.target.value;
      const detected = detectSocialName(url);
      if (detected && (nameAutoDetected || watch("name") === "")) {
        setValue("name", detected, { shouldValidate: true });
        setNameAutoDetected(true);
      } else if (!detected && nameAutoDetected) {
        setValue("name", "", { shouldValidate: false });
        setNameAutoDetected(false);
      }
    };

    useImperativeHandle(ref, () => ({
      openCreate: (tId, cId) => {
        setMode("create");
        setTrustCenterId(tId);
        setConnectionId(cId);
        setEditNode(null);
        setNameAutoDetected(false);
        reset({ name: "", url: "" });
        dialogRef.current?.open();
      },
      openEdit: (node) => {
        setMode("edit");
        setEditNode(node);
        setNameAutoDetected(false);
        reset({ name: node.name, url: node.url });
        dialogRef.current?.open();
      },
    }));

    const onSubmit = async (data: UrlFormData) => {
      if (mode === "create") {
        await create({
          variables: {
            input: { trustCenterId, name: data.name, url: data.url },
          },
          updater: (store) => {
            const payload = store.getRootField("createComplianceExternalURL");
            const edge = payload?.getLinkedRecord("complianceExternalUrlEdge");
            if (!edge) return;
            const connection = store.get(connectionId);
            if (!connection) return;
            ConnectionHandler.insertEdgeAfter(connection, edge);
          },
          onCompleted: (_, errs) => {
            if (!errs?.length) {
              reset();
              dialogRef.current?.close();
            }
          },
        });
      } else if (editNode) {
        await update({
          variables: { input: { id: editNode.id, name: data.name, url: data.url } },
          onCompleted: (_, errs) => {
            if (!errs?.length) {
              reset();
              dialogRef.current?.close();
            }
          },
        });
      }
    };

    const isSubmitting = isCreating || isUpdating;
    const title = mode === "create" ? __("Add link") : __("Edit link");

    return (
      <Dialog ref={dialogRef} title={title} onClose={() => reset()}>
        <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
          <DialogContent padded className="space-y-6">
            <Field
              {...register("url", { onChange: handleUrlChange })}
              label={__("URL")}
              type="url"
              required
              placeholder="https://example.com"
              error={errors.url?.message}
            />
            <Field
              {...register("name")}
              label={__("Name")}
              type="text"
              required
              placeholder={__("e.g. Twitter, LinkedIn")}
              error={errors.name?.message}
            />
          </DialogContent>
          <DialogFooter>
            <Button type="submit" disabled={isSubmitting} icon={isSubmitting ? Spinner : undefined}>
              {mode === "create" ? __("Add link") : __("Save changes")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    );
  },
);

function ExternalUrlRow(props: {
  node: UrlNode;
  canEdit: boolean;
  connectionId: string;
  isDragging: boolean;
  isDropTarget: boolean;
  onDragStart: () => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: () => void;
  onDragEnd: () => void;
  onEdit: (node: UrlNode) => void;
}) {
  const {
    node, canEdit, connectionId, isDragging, isDropTarget,
    onDragStart, onDragOver, onDrop, onDragEnd, onEdit,
  } = props;
  const { __ } = useTranslate();
  const [isMouseDown, setIsMouseDown] = useState(false);

  const [deleteUrl] = useMutationWithToasts<CompliancePageExternalUrlsSection_deleteMutation>(
    deleteMutation,
    { successMessage: __("Link removed."), errorMessage: __("Failed to remove link.") },
  );

  const handleDelete = () => {
    void deleteUrl({
      variables: { input: { id: node.id } },
      updater: (store) => {
        const connection = store.get(connectionId);
        if (!connection) return;
        ConnectionHandler.deleteNode(connection, node.id);
      },
    });
  };

  const className = [
    isDragging && "opacity-50 cursor-grabbing",
    !isDragging && !isMouseDown && "cursor-grab",
    !isDragging && isMouseDown && "cursor-grabbing",
    isDropTarget && "!bg-primary-50 border-y-2 border-primary-500",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <Tr
      draggable={canEdit}
      onDragStart={canEdit ? onDragStart : undefined}
      onDragOver={canEdit ? onDragOver : undefined}
      onDrop={canEdit ? onDrop : undefined}
      onDragEnd={canEdit ? onDragEnd : undefined}
      onMouseDown={canEdit ? () => setIsMouseDown(true) : undefined}
      onMouseUp={canEdit ? () => setIsMouseDown(false) : undefined}
      onMouseLeave={canEdit ? () => setIsMouseDown(false) : undefined}
      className={className}
    >
      <Td>
        <div className="flex items-center gap-3 text-txt-secondary">
          <SocialIcon socialName={detectSocialName(node.url)} size={16} className="shrink-0" />
          <span className="font-medium">{node.name}</span>
        </div>
      </Td>
      <Td>
        <span className="text-txt-secondary truncate">{node.url}</span>
      </Td>
      <Td noLink width={canEdit ? 144 : 56} className="text-end">
        <div className="flex gap-2 justify-end">
          <Button variant="secondary" icon={IconArrowLink} onClick={() => safeOpenUrl(node.url)} />
          {canEdit && (
            <>
              <Button variant="secondary" icon={IconPencil} onClick={() => onEdit(node)} />
              <Button variant="danger" icon={IconTrashCan} onClick={handleDelete} />
            </>
          )}
        </div>
      </Td>
    </Tr>
  );
}

export function CompliancePageExternalUrlsSection(props: {
  trustCenterRef: CompliancePageExternalUrlsSection_trustCenterFragment$key;
}) {
  const { __ } = useTranslate();
  const [, startTransition] = useTransition();
  const dialogRef = useRef<ExternalUrlDialogRef>(null);

  const [trustCenter, refetch] = useRefetchableFragment<
    CompliancePageExternalUrlsSection_trustCenterRefetchQuery,
    CompliancePageExternalUrlsSection_trustCenterFragment$key
  >(trustCenterFragment, props.trustCenterRef);

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);

  const [updateRank] = useMutationWithToasts<CompliancePageExternalUrlsSection_updateMutation>(
    updateMutation,
    { successMessage: __("Order updated."), errorMessage: __("Failed to update order.") },
  );

  const edges = trustCenter.externalUrls.edges;
  const canEdit = trustCenter.canUpdate;

  const connectionId = trustCenter.externalUrls.__id;

  const handleCreate = () => {
    dialogRef.current?.openCreate(trustCenter.id, connectionId);
  };

  const handleEdit = (node: UrlNode) => {
    dialogRef.current?.openEdit(node);
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
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("Custom links")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Add external URLs to display on your compliance page")}
          </p>
        </div>
        {canEdit && (
          <Button icon={IconPlusLarge} onClick={handleCreate}>
            {__("Add link")}
          </Button>
        )}
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th>{__("Name")}</Th>
            <Th>{__("URL")}</Th>
            <Th />
          </Tr>
        </Thead>
        <Tbody>
          {edges.length === 0 && (
            <Tr>
              <Td colSpan={3} className="text-center text-txt-secondary">
                {__("No custom links yet")}
              </Td>
            </Tr>
          )}
          {edges.map(({ node }, index) => (
            <ExternalUrlRow
              key={node.id}
              node={node}
              canEdit={canEdit}
              connectionId={connectionId}
              isDragging={draggedIndex === index}
              isDropTarget={dragOverIndex === index && draggedIndex !== index}
              onDragStart={() => setDraggedIndex(index)}
              onDragOver={e => handleDragOver(e, index)}
              onDrop={() => void handleDrop(index)}
              onDragEnd={() => {
                setDraggedIndex(null);
                setDragOverIndex(null);
              }}
              onEdit={handleEdit}
            />
          ))}
        </Tbody>
      </Table>

      {edges.length > 1 && canEdit && (
        <p className="text-sm text-txt-tertiary">
          {__("Drag and drop to change the displayed order")}
        </p>
      )}

      <ExternalUrlDialog ref={dialogRef} />
    </div>
  );
}
