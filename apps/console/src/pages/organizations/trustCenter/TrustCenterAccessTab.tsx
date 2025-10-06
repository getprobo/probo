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
  IconCheckmark1,
  IconCrossLargeX,
  IconChevronDown,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { formatDate } from "@probo/helpers";
import { useOutletContext } from "react-router";
import { useState, useCallback, useEffect } from "react";
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
    active: z.boolean(),
  });

  const [createInvitation, isCreating] = useMutationWithToasts(createTrustCenterAccessMutation, {
    successMessage: __("Access created successfully"),
    errorMessage: __("Failed to create access. Please try again."),
  });
  const [updateInvitation, isUpdating] = useMutationWithToasts(updateTrustCenterAccessMutation, {
    successMessage: __("Access updated successfully"),
    errorMessage: __("Failed to update access"),
  });
  const [deleteInvitation, isDeleting] = useMutationWithToasts(deleteTrustCenterAccessMutation, {
    successMessage: __("Access deleted successfully"),
    errorMessage: __("Failed to delete access"),
  });

  const dialogRef = useDialogRef();
  const editDialogRef = useDialogRef();
  const [editingAccess, setEditingAccess] = useState<AccessType | null>(null);
  const [selectedDocumentAccesses, setSelectedDocumentAccesses] = useState<Set<string>>(new Set());
  const [pendingEditEmail, setPendingEditEmail] = useState<string | null>(null);

  const inviteForm = useFormWithSchema(inviteSchema, {
    defaultValues: { name: "", email: "" },
  });

  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: "", active: false },
  });

  type DocumentAccessType = {
    id: string;
    active: boolean;
    document?: {
      id: string;
      title: string;
      documentType: string;
    } | null;
    report?: {
      id: string;
      filename: string;
      audit: {
        id: string;
        framework: {
          name: string;
        };
      };
    } | null;
  };

  type AccessType = {
    id: string;
    email: string;
    name: string;
    active: boolean;
    hasAcceptedNonDisclosureAgreement: boolean;
    createdAt: string;
    documentAccesses: DocumentAccessType[];
  };

  const { data: trustCenterData, loadMore, hasNext, isLoadingNext } = useTrustCenterAccesses(organization.trustCenter?.id || "");

  const accesses: AccessType[] = trustCenterData?.node?.accesses?.edges?.map((edge: any) => ({
    id: edge.node.id,
    email: edge.node.email,
    name: edge.node.name,
    active: edge.node.active,
    hasAcceptedNonDisclosureAgreement: edge.node.hasAcceptedNonDisclosureAgreement,
    createdAt: edge.node.createdAt,
    documentAccesses: edge.node.documentAccesses?.edges?.map((docEdge: any) => ({
      id: docEdge.node.id,
      active: docEdge.node.active,
      document: docEdge.node.document,
      report: docEdge.node.report
    })) ?? []
  })) ?? [];


  const handleInvite = inviteForm.handleSubmit(async (data) => {
    if (!organization.trustCenter?.id) {
      return;
    }

    const connectionId = trustCenterData?.node?.accesses?.__id;
    const email = data.email.trim();

    await createInvitation({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
          email: email,
          name: data.name.trim(),
          active: false,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onSuccess: () => {
        dialogRef.current?.close();
        inviteForm.reset();
        setPendingEditEmail(email);
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


  const getActiveDocumentIds = useCallback((access: AccessType) => {
    return access.documentAccesses
      .filter(docAccess => docAccess.active)
      .map(docAccess => docAccess.document?.id || docAccess.report?.id)
      .filter((id): id is string => id !== undefined);
  }, []);

  const handleEditAccess = useCallback((access: AccessType) => {
    setEditingAccess(access);
    editForm.reset({ name: access.name, active: access.active });
    setSelectedDocumentAccesses(new Set(getActiveDocumentIds(access)));
    editDialogRef.current?.open();
  }, [editDialogRef, editForm, getActiveDocumentIds]);

  useEffect(() => {
    if (pendingEditEmail && accesses.length > 0) {
      const newAccess = accesses.find(access => access.email === pendingEditEmail);
      if (newAccess) {
        setPendingEditEmail(null);
        setEditingAccess(newAccess);
        editForm.reset({ name: newAccess.name, active: true });
        setSelectedDocumentAccesses(new Set(getActiveDocumentIds(newAccess)));
        editDialogRef.current?.open();
      }
    }
  }, [accesses, pendingEditEmail, editForm, editDialogRef, getActiveDocumentIds]);

  const handleToggleDocumentAccess = useCallback((documentId: string, active: boolean) => {
    setSelectedDocumentAccesses(prev => {
      const newSet = new Set(prev);
      if (active) {
        newSet.add(documentId);
      } else {
        newSet.delete(documentId);
      }
      return newSet;
    });
  }, []);

  const handleUpdateName = editForm.handleSubmit(async (data) => {
    if (!editingAccess) return;

    const { documentIds, reportIds } = editingAccess.documentAccesses.reduce(
      (acc, docAccess) => {
        const id = docAccess.document?.id || docAccess.report?.id;
        if (id && selectedDocumentAccesses.has(id)) {
          if (docAccess.document?.id) {
            acc.documentIds.push(docAccess.document.id);
          } else if (docAccess.report?.id) {
            acc.reportIds.push(docAccess.report.id);
          }
        }
        return acc;
      },
      { documentIds: [] as string[], reportIds: [] as string[] }
    );

    await updateInvitation({
      variables: {
        input: {
          id: editingAccess.id,
          name: data.name.trim(),
          active: data.active,
          documentIds,
          reportIds,
        },
      },
      onSuccess: () => {
        editDialogRef.current?.close();
        setEditingAccess(null);
        editForm.reset();
        setSelectedDocumentAccesses(new Set());
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
            {__("Add Access")}
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
          <>
            <Table>
              <Thead>
                <Tr>
                  <Th>{__("Name")}</Th>
                  <Th>{__("Email")}</Th>
                  <Th>{__("Date")}</Th>
                  <Th>{__("Active")}</Th>
                  <Th>{__("Documents")}</Th>
                  <Th>{__("NDA")}</Th>
                  <Th></Th>
                </Tr>
              </Thead>
              <Tbody>
                {accesses.map((access) => {
                  const activeDocuments = access.documentAccesses.filter(doc => doc.active).length;
                  const totalDocuments = access.documentAccesses.length;

                  return (
                    <Tr
                      key={access.id}
                      onClick={() => handleEditAccess(access)}
                      className="cursor-pointer hover:bg-bg-secondary transition-colors"
                    >
                      <Td className="font-medium">{access.name}</Td>
                      <Td>{access.email}</Td>
                      <Td>
                        {formatDate(access.createdAt)}
                      </Td>
                      <Td>
                        {access.active ? (
                          <IconCheckmark1 size={16} className="text-txt-success" />
                        ) : (
                          <IconCrossLargeX size={16} className="text-txt-danger" />
                        )}
                      </Td>
                      <Td>
                        {totalDocuments > 0 ? `${activeDocuments}/${totalDocuments}` : '0/0'}
                      </Td>
                      <Td>
                        {access.hasAcceptedNonDisclosureAgreement && (
                          <IconCheckmark1 size={16} className="text-txt-success" />
                        )}
                      </Td>
                    <Td noLink width={160} className="text-end">
                      <div
                        className="flex gap-2 justify-end"
                        onClick={(e) => e.stopPropagation()}
                      >
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
                  );
                })}
              </Tbody>
            </Table>
            {hasNext && (
              <Button
                variant="tertiary"
                onClick={loadMore}
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
              {__("Create Access")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog
        ref={editDialogRef}
        title={__("Edit Access")}
      >
        <form onSubmit={handleUpdateName}>
          <DialogContent padded className="space-y-6">
            <div>
              <p className="text-txt-secondary text-sm mb-4">
                {__("Update access settings and document permissions")}
              </p>

              <Field
                label={__("Full Name")}
                required
                error={editForm.formState.errors.name?.message}
                {...editForm.register("name")}
                placeholder={__("John Doe")}
              />

              <div className="flex items-center justify-between mt-6">
                <div>
                  <label className="font-medium text-txt-primary">
                    {__("Active Status")}
                  </label>
                  <p className="text-sm text-txt-secondary">
                    {__("Enable or disable access for this user")}
                  </p>
                </div>
                <Checkbox
                  checked={editForm.watch("active")}
                  onChange={(checked) => editForm.setValue("active", checked)}
                />
              </div>
            </div>

            {editingAccess && editingAccess.documentAccesses.length > 0 && (
              <div>
                <h4 className="font-medium text-txt-primary mb-4">
                  {__("Document Access Permissions")}
                </h4>
                <div className="bg-bg-secondary rounded-lg overflow-hidden">
                  <Table>
                    <Thead>
                      <Tr>
                        <Th>{__("Name")}</Th>
                        <Th>{__("Type")}</Th>
                        <Th>{__("Category")}</Th>
                        <Th>
                          <div className="flex justify-end">
                            {__("Access")}
                          </div>
                        </Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {editingAccess.documentAccesses.map((docAccess) => {
                        const getDocumentInfo = () => {
                          const isDocument = !!docAccess.document;
                          return {
                            isDocument,
                            name: docAccess.document?.title || docAccess.report?.filename || __("Unknown Item"),
                            type: isDocument ? __("Document") : __("Report"),
                            category: isDocument
                              ? docAccess.document?.documentType
                              : docAccess.report?.audit?.framework?.name || __("Compliance Report"),
                            id: docAccess.document?.id || docAccess.report?.id || ''
                          };
                        };

                        const { isDocument, name, type, category, id } = getDocumentInfo();

                        return (
                          <Tr key={docAccess.id}>
                            <Td>
                              <div className="font-medium text-txt-primary">
                                {name}
                              </div>
                            </Td>
                            <Td>
                              <div className="flex items-center space-x-2">
                                <div className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                  isDocument
                                    ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
                                    : 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                                }`}>
                                  {type}
                                </div>
                              </div>
                            </Td>
                            <Td>
                              <div className="text-txt-secondary">
                                {category || "-"}
                              </div>
                            </Td>
                            <Td>
                              <div className="flex justify-end">
                                <Checkbox
                                  checked={selectedDocumentAccesses.has(id)}
                                  onChange={(active) => {
                                    if (id) handleToggleDocumentAccess(id, active);
                                  }}
                                />
                              </div>
                            </Td>
                          </Tr>
                        );
                      })}
                    </Tbody>
                  </Table>
                </div>
              </div>
            )}
          </DialogContent>

          <DialogFooter>
            <Button type="submit" disabled={isUpdating}>
              {isUpdating && <Spinner />}
              {__("Update Access")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  );
}
