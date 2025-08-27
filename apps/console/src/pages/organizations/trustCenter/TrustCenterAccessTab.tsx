import {
  Button,
  Card,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useDialogRef,
  IconTrashCan,
  IconPencil,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOutletContext } from "react-router";
import { useState, useCallback } from "react";
import z from "zod";
import {
  useTrustCenterAccesses,
  createTrustCenterAccessMutation,
  updateTrustCenterAccessMutation,
  deleteTrustCenterAccessMutation
} from "/hooks/graph/TrustCenterAccessGraph";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

type ContextType = {
  organization: {
    id: string;
    trustCenter?: {
      id: string;
    };
  };
};

export default function TrustCenterAccessTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<ContextType>();

  const inviteSchema = z.object({
    name: z.string().min(1, __("Name is required")).min(2, __("Name must be at least 2 characters long")),
    email: z.string().min(1, __("Email is required")).email(__("Please enter a valid email address")),
  });

  const editSchema = z.object({
    name: z.string().min(1, __("Name is required")).min(2, __("Name must be at least 2 characters long")),
  });

  const [createInvitation, isCreating] = useMutationWithToasts(createTrustCenterAccessMutation, {
    successMessage: __("Access invitation sent successfully"),
    errorMessage: __("Failed to send invitation. Please try again."),
  });
  const [updateInvitation, isUpdating] = useMutationWithToasts(updateTrustCenterAccessMutation, {
    successMessage: __("Access updated successfully"),
    errorMessage: __("Failed to update access. Please try again."),
  });
  const [deleteInvitation, isDeleting] = useMutationWithToasts(deleteTrustCenterAccessMutation, {
    successMessage: __("Access deleted successfully"),
    errorMessage: __("Failed to delete access. Please try again."),
  });

  const dialogRef = useDialogRef();
  const editDialogRef = useDialogRef();
  const [editingAccess, setEditingAccess] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const inviteForm = useFormWithSchema(inviteSchema, {
    defaultValues: { name: "", email: "" },
  });

  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: "" },
  });

  type AccessType = {
    id: string;
    email: string;
    name: string;
    active: boolean;
    createdAt: Date;
    };

  const trustCenterData = useTrustCenterAccesses(organization.trustCenter?.id || "");

  const accesses: AccessType[] = trustCenterData?.node?.accesses?.edges?.map(edge => ({
    id: edge.node.id,
    email: edge.node.email,
    name: edge.node.name,
    active: edge.node.active,
    createdAt: new Date(edge.node.createdAt)
  })) ?? [];

  const handleInvite = inviteForm.handleSubmit(async (data) => {
    if (!organization.trustCenter?.id) {
      return;
    }

    const connectionId = trustCenterData?.node?.accesses?.__id;

    await createInvitation({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
          email: data.email.trim(),
          name: data.name.trim(),
          active: true,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onSuccess: () => {
        dialogRef.current?.close();
        inviteForm.reset();
      },
    });
  });

  const handleDelete = useCallback(async (id: string) => {
    const connectionId = trustCenterData?.node?.accesses?.__id;

    await deleteInvitation({
      variables: {
        input: { id },
        connections: connectionId ? [connectionId] : [],
      },
    });
  }, [deleteInvitation, trustCenterData]);

  const handleToggleActive = useCallback(async (id: string, active: boolean) => {
    await updateInvitation({
      variables: {
        input: { id, active },
      },
      successMessage: active ? __("Access activated") : __("Access deactivated"),
    });
  }, [updateInvitation, __]);

  const handleEditAccess = useCallback((access: AccessType) => {
    setEditingAccess({ id: access.id, name: access.name });
    editForm.reset({ name: access.name });
    editDialogRef.current?.open();
  }, [editDialogRef, editForm]);

  const handleUpdateName = editForm.handleSubmit(async (data) => {
    if (!editingAccess) return;

    await updateInvitation({
      variables: {
        input: {
          id: editingAccess.id,
          name: data.name.trim(),
        },
      },
      successMessage: __("Name updated successfully"),
      onSuccess: () => {
        editDialogRef.current?.close();
        setEditingAccess(null);
        editForm.reset();
      },
    });
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("External Access")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__("Manage who can access your trust center with time-limited tokens")}
          </p>
        </div>
        {organization.trustCenter?.id && (
          <Button onClick={() => {
            inviteForm.reset();
            dialogRef.current?.open();
          }}>
            {__("Invite")}
          </Button>
        )}
      </div>

      <Card padded>
        {!organization.trustCenter?.id ? (
          <div className="text-center text-txt-tertiary py-8">
            <Spinner />
          </div>
        ) : accesses.length === 0 ? (
          <div className="text-center text-txt-tertiary py-8">
            {__("No external access granted yet")}
          </div>
        ) : (
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Email")}</Th>
                <Th>{__("Date")}</Th>
                <Th>{__("Active")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {accesses.map((access) => (
                <Tr key={access.id}>
                  <Td className="font-medium">{access.name}</Td>
                  <Td>{access.email}</Td>
                  <Td>
                    {access.createdAt.toLocaleDateString()}
                  </Td>
                  <Td>
                    <Checkbox
                      checked={access.active}
                      onChange={(active) => handleToggleActive(access.id, active)}
                    />
                  </Td>
                  <Td noLink width={160} className="text-end">
                    <div className="flex gap-2 justify-end">
                      <Button
                        variant="secondary"
                        onClick={() => handleEditAccess(access)}
                        disabled={isUpdating}
                        icon={IconPencil}
                      />
                      <Button
                        variant="secondary"
                        onClick={() => handleDelete(access.id)}
                        disabled={isDeleting}
                        icon={IconTrashCan}
                      />
                    </div>
                  </Td>
                </Tr>
              ))}
            </Tbody>
          </Table>
        )}
      </Card>

      <Dialog
        ref={dialogRef}
        title={__("Invite External Access")}
      >
        <form onSubmit={handleInvite}>
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {__("Send a 7-day access token to an external person to view your trust center")}
            </p>

            <Field
              label={__("Full Name")}
              required
              error={inviteForm.formState.errors.name?.message}
              {...inviteForm.register("name")}
              placeholder={__("John Doe")}
            />

            <Field
              label={__("Email Address")}
              required
              error={inviteForm.formState.errors.email?.message}
              type="email"
              {...inviteForm.register("email")}
              placeholder={__("john@example.com")}
            />
          </DialogContent>

          <DialogFooter>
            <Button type="submit" disabled={isCreating}>
              {isCreating && <Spinner />}
              {__("Send Invitation")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog
        ref={editDialogRef}
        title={__("Edit Access Name")}
      >
        <form onSubmit={handleUpdateName}>
          <DialogContent padded className="space-y-4">
            <p className="text-txt-secondary text-sm">
              {__("Update the display name for this access invitation")}
            </p>

            <Field
              label={__("Full Name")}
              required
              error={editForm.formState.errors.name?.message}
              {...editForm.register("name")}
              placeholder={__("John Doe")}
            />
          </DialogContent>

          <DialogFooter>
            <Button type="submit" disabled={isUpdating}>
              {isUpdating && <Spinner />}
              {__("Update Name")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  );
}
