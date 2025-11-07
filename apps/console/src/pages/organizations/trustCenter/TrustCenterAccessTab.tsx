import {
  Badge,
  Button,
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
  IconPlusLarge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { formatDate } from "@probo/helpers";
import { useOutletContext } from "react-router";
import { useState, useCallback, useEffect, useRef } from "react";
import { useQueryLoader, usePreloadedQuery } from 'react-relay';
import z from "zod";
import {
  useTrustCenterAccesses,
  createTrustCenterAccessMutation,
  updateTrustCenterAccessMutation,
  deleteTrustCenterAccessMutation,
  loadTrustCenterAccessDocumentAccessesQuery
} from "/hooks/graph/TrustCenterAccessGraph";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { IfAuthorized } from "/permissions/IfAuthorized";

type ContextType = {
  organization: {
    id: string;
    trustCenter?: {
      id: string;
    };
    documents?: {
      edges: Array<{
        node: {
          id: string;
          title: string;
          documentType: string;
          trustCenterVisibility: string;
        };
      }>;
    };
    audits?: {
      edges: Array<{
        node: {
          id: string;
          filename: string;
          trustCenterVisibility: string;
          framework: {
            name: string;
          };
        };
      }>;
    };
    trustCenterFiles?: {
      edges: Array<{
        node: {
          id: string;
          name: string;
          category: string;
          trustCenterVisibility: string;
        };
      }>;
    };
  };
};

type DocumentAccessInfo = {
  id: string;
  active: boolean;
  requested: boolean;
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
  trustCenterFile?: {
    id: string;
    name: string;
    category: string;
  } | null;
};

function DocumentAccessesLoader({
  queryReference,
  onDataLoaded
}: {
  queryReference: any;
  onDataLoaded: (documentAccesses: DocumentAccessInfo[]) => void;
}) {
  const data = usePreloadedQuery(loadTrustCenterAccessDocumentAccessesQuery, queryReference);

  useEffect(() => {
    if (data && typeof data === 'object' && 'node' in data) {
      const node = (data as any).node;
      if (node?.availableDocumentAccesses?.edges) {
        const documentAccesses = node.availableDocumentAccesses.edges.map((edge: any) => edge.node);
        onDataLoaded(documentAccesses);
      }
    }
  }, [data, onDataLoaded]);

  return null;
}

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
    errorMessage: __("Failed to create access"),
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
  const [editingDocumentAccesses, setEditingDocumentAccesses] = useState<DocumentAccessType[]>([]);
  const [selectedDocumentAccesses, setSelectedDocumentAccesses] = useState<Set<string>>(new Set());
  const [pendingEditEmail, setPendingEditEmail] = useState<string | null>(null);
  const [documentAccessesQueryReference, loadDocumentAccessesQuery] = useQueryLoader(loadTrustCenterAccessDocumentAccessesQuery);
  const loadedAccessIdRef = useRef<string | null>(null);
  const [isLoadingDocumentAccesses, setIsLoadingDocumentAccesses] = useState(false);

  useEffect(() => {
    if (editingAccess?.id && loadedAccessIdRef.current !== editingAccess.id) {
      loadedAccessIdRef.current = editingAccess.id;
      setIsLoadingDocumentAccesses(true);
      loadDocumentAccessesQuery({ accessId: editingAccess.id }, { fetchPolicy: 'network-only' });
    }
  }, [editingAccess?.id, loadDocumentAccessesQuery]);

  const handleDocumentAccessesLoaded = useCallback((documentAccesses: DocumentAccessInfo[]) => {
    setEditingDocumentAccesses(documentAccesses);

    const activeIds = new Set<string>(
      documentAccesses
        .filter((docAccess) => docAccess.active)
        .map((docAccess) => docAccess.document?.id || docAccess.report?.id || docAccess.trustCenterFile?.id)
        .filter((id: unknown): id is string => typeof id === 'string')
    );
    setSelectedDocumentAccesses(activeIds);
    setIsLoadingDocumentAccesses(false);
  }, []);

  const formattedDocumentAccesses = editingDocumentAccesses
    ?.map((docAccess) => getDocumentAccessInfo(docAccess, __))
    ?.filter((info) => info !== null) ?? [];

  const inviteForm = useFormWithSchema(inviteSchema, {
    defaultValues: { name: "", email: "" },
  });

  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: "", active: false },
  });

  type DocumentAccessType = {
    id: string;
    active: boolean;
    requested: boolean;
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
    trustCenterFile?: {
      id: string;
      name: string;
      category: string;
    } | null;
  };

  function getDocumentAccessInfo(
    docAccess: DocumentAccessType,
    __: (key: string) => string
  ) {
    if (!!docAccess.document) {
      return {
        variant: "info" as const,
        name: docAccess.document?.title,
        type: __("Document"),
        category: docAccess.document?.documentType,
        id: docAccess.document?.id,
        requested: docAccess.requested,
        active: docAccess.active,
      };
    }
    if (!!docAccess.report) {
      return {
        variant: "success" as const,
        name: docAccess.report?.filename,
        type: __("Report"),
        category: docAccess.report?.audit?.framework?.name,
        id: docAccess.report?.id,
        requested: docAccess.requested,
        active: docAccess.active,
      };
    }
    if (!!docAccess.trustCenterFile) {
      return {
        variant: "highlight" as const,
        name: docAccess.trustCenterFile?.name,
        type: __("File"),
        category: docAccess.trustCenterFile?.category,
        id: docAccess.trustCenterFile?.id,
        requested: docAccess.requested,
        active: docAccess.active,
      };
    }

    return null;
  }

  type AccessType = {
    id: string;
    email: string;
    name: string;
    active: boolean;
    hasAcceptedNonDisclosureAgreement: boolean;
    createdAt: string;
    lastTokenExpiresAt: string | null;
    pendingRequestCount: number;
    activeCount: number;
    documentAccesses?: DocumentAccessType[];
  };

  const { data: trustCenterData, loadMore, hasNext, isLoadingNext } = useTrustCenterAccesses(organization.trustCenter?.id || "");

  const accesses: AccessType[] = trustCenterData?.node?.accesses?.edges?.map((edge: any) => ({
    id: edge.node.id,
    email: edge.node.email,
    name: edge.node.name,
    active: edge.node.active,
    hasAcceptedNonDisclosureAgreement: edge.node.hasAcceptedNonDisclosureAgreement,
    createdAt: edge.node.createdAt,
    lastTokenExpiresAt: edge.node.lastTokenExpiresAt,
    pendingRequestCount: edge.node.pendingRequestCount || 0,
    activeCount: edge.node.activeCount || 0,
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



  const handleEditAccess = useCallback((access: AccessType) => {
    loadedAccessIdRef.current = null;
    setEditingAccess(access);
    setEditingDocumentAccesses([]);
    setSelectedDocumentAccesses(new Set());
    setIsLoadingDocumentAccesses(false);
    editForm.reset({ name: access.name, active: access.active });
    editDialogRef.current?.open();
  }, [editDialogRef, editForm]);

  useEffect(() => {
    if (pendingEditEmail && accesses.length > 0) {
      const newAccess = accesses.find(access => access.email === pendingEditEmail);
      if (newAccess) {
        setPendingEditEmail(null);
        loadedAccessIdRef.current = null;
        setEditingAccess(newAccess);
        editForm.reset({ name: newAccess.name, active: true });
        setEditingDocumentAccesses([]);
        setSelectedDocumentAccesses(new Set());
        editDialogRef.current?.open();
        setTimeout(() => {
          dialogRef.current?.close();
        }, 50);
        setTimeout(() => {
          inviteForm.reset();
        }, 300);
      }
    }
  }, [accesses, pendingEditEmail, editForm, dialogRef, editDialogRef, inviteForm]);

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

    const { documentIds, reportIds, trustCenterFileIds } = editingDocumentAccesses.reduce(
      (acc, docAccess) => {
        const id = docAccess.document?.id || docAccess.report?.id || docAccess.trustCenterFile?.id;
        if (id && selectedDocumentAccesses.has(id)) {
          if (docAccess.document?.id) {
            acc.documentIds.push(docAccess.document.id);
          } else if (docAccess.report?.id) {
            acc.reportIds.push(docAccess.report.id);
          } else if (docAccess.trustCenterFile?.id) {
            acc.trustCenterFileIds.push(docAccess.trustCenterFile.id);
          }
        }
        return acc;
      },
      { documentIds: [] as string[], reportIds: [] as string[], trustCenterFileIds: [] as string[] }
    );

    await updateInvitation({
      variables: {
        input: {
          id: editingAccess.id,
          name: data.name.trim(),
          active: data.active,
          documentIds,
          reportIds,
          trustCenterFileIds,
        },
      },
      onSuccess: () => {
        editDialogRef.current?.close();
        setEditingAccess(null);
        editForm.reset();
        setEditingDocumentAccesses([]);
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
          <IfAuthorized entity="TrustCenter" action="update">
            <Button icon={IconPlusLarge} onClick={() => {
              inviteForm.reset();
              dialogRef.current?.open();
            }}>
              {__("Add Access")}
            </Button>
          </IfAuthorized>
        )}
      </div>

      {!organization.trustCenter?.id ? (
        <Table>
          <Tbody>
            <Tr>
              <Td className="text-center text-txt-tertiary py-8">
                <Spinner />
              </Td>
            </Tr>
          </Tbody>
        </Table>
      ) : accesses.length === 0 ? (
        <Table>
          <Tbody>
            <Tr>
              <Td className="text-center text-txt-tertiary py-8">
                {__("No external access granted yet")}
              </Td>
            </Tr>
          </Tbody>
        </Table>
      ) : (
        <>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Email")}</Th>
                <Th>{__("Date")}</Th>
                <Th>{__("Expires")}</Th>
                <Th className="text-center">{__("Active")}</Th>
                <Th className="text-center">{__("Access")}</Th>
                <Th className="text-center">{__("Requests")}</Th>
                <Th className="text-center">{__("NDA")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {accesses.map((access) => {
                const isExpired = access.lastTokenExpiresAt ? new Date(access.lastTokenExpiresAt) < new Date() : false;

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
                    <Td className={isExpired ? "text-txt-danger" : ""}>
                      {access.lastTokenExpiresAt ? formatDate(access.lastTokenExpiresAt) : "-"}
                    </Td>
                    <Td>
                      <div className="flex justify-center">
                        {access.active ? (
                          <IconCheckmark1 size={16} className="text-txt-success" />
                        ) : (
                          <IconCrossLargeX size={16} className="text-txt-danger" />
                        )}
                      </div>
                    </Td>
                    <Td className="text-center">
                      {access.activeCount}
                    </Td>
                    <Td className="text-center">
                      {access.pendingRequestCount > 0 ? access.pendingRequestCount : ""}
                    </Td>
                    <Td>
                      <div className="flex justify-center">
                        {access.hasAcceptedNonDisclosureAgreement && (
                          <IconCheckmark1 size={16} className="text-txt-success" />
                        )}
                      </div>
                    </Td>
                  <Td noLink width={160} className="text-end">
                    <div
                      className="flex gap-2 justify-end"
                      onClick={(e) => e.stopPropagation()}
                    >
                      <IfAuthorized entity="TrustCenter" action="update">
                        <Button
                          variant="secondary"
                          onClick={() => handleEditAccess(access)}
                          disabled={isUpdating}
                          icon={IconPencil}
                        />
                      </IfAuthorized>
                      <IfAuthorized entity="TrustCenter" action="delete">
                        <Button
                          variant="danger"
                          onClick={() => handleDelete(access.id)}
                          disabled={isDeleting}
                          icon={IconTrashCan}
                        />
                      </IfAuthorized>
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

      <Dialog
        ref={dialogRef}
        title={__("Invite External Access")}
      >
        <form onSubmit={handleInvite}>
          <DialogContent padded className="space-y-6">
            <div>
              <p className="text-txt-secondary text-sm mb-4">
                {__("Send a 30-day access token to an external person to view your trust center")}
              </p>

              <Field
                label={__("Full Name")}
                required
                error={inviteForm.formState.errors.name?.message}
                {...inviteForm.register("name")}
                placeholder={__("John Doe")}
              />

              <div className="mt-4">
                <Field
                  label={__("Email Address")}
                  required
                  error={inviteForm.formState.errors.email?.message}
                  type="email"
                  {...inviteForm.register("email")}
                  placeholder={__("john@example.com")}
                />
              </div>
            </div>
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
        {documentAccessesQueryReference && (
          <DocumentAccessesLoader
            queryReference={documentAccessesQueryReference}
            onDataLoaded={handleDocumentAccessesLoaded}
          />
        )}
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

            <div>
              <div className="flex justify-between items-center mb-4">
                <h4 className="font-medium text-txt-primary">
                  {__("Document Access Permissions")}
                </h4>
                {!isLoadingDocumentAccesses && formattedDocumentAccesses.length > 0 && (
                  <Button
                    type="button"
                    variant="tertiary"
                    onClick={() => {
                      if (selectedDocumentAccesses.size === formattedDocumentAccesses.length) {
                        setSelectedDocumentAccesses(new Set());
                      } else {
                        const allIds = new Set(formattedDocumentAccesses.map(doc => doc.id).filter((id): id is string => !!id));
                        setSelectedDocumentAccesses(allIds);
                      }
                    }}
                    className="text-xs h-7 min-h-7"
                  >
                    {selectedDocumentAccesses.size === formattedDocumentAccesses.length
                      ? __("Clear All")
                      : __("Select All")}
                  </Button>
                )}
              </div>

              {isLoadingDocumentAccesses ? (
                <div className="flex justify-center items-center py-12">
                  <Spinner />
                </div>
              ) : formattedDocumentAccesses.length > 0 ? (
                <div className="bg-bg-secondary rounded-lg overflow-hidden">
                  <Table>
                    <Thead>
                      <Tr>
                        <Th>{__("Name")}</Th>
                        <Th>{__("Type")}</Th>
                        <Th>{__("Category")}</Th>
                        <Th></Th>
                        <Th>
                          <div className="flex justify-end">
                            {__("Access")}
                          </div>
                        </Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {formattedDocumentAccesses.map((info) => {
                        const { variant, name, type, category, id, requested } = info;

                        return (
                          <Tr key={id}>
                            <Td>
                              <div className="font-medium text-txt-primary">
                                {name}
                              </div>
                            </Td>
                            <Td>
                              <Badge variant={variant}>
                                {type}
                              </Badge>
                            </Td>
                            <Td>
                              <div className="text-txt-secondary">
                                {category || "-"}
                              </div>
                            </Td>
                            <Td>
                              {requested && (
                                <Badge variant="warning">
                                  {__("Requested")}
                                </Badge>
                              )}
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
              ) : (
                <div className="text-center text-txt-tertiary py-8">
                  {__("No documents available")}
                </div>
              )}
            </div>
          </DialogContent>

          <DialogFooter>
            <Button type="submit" disabled={isUpdating || isLoadingDocumentAccesses}>
              {(isUpdating || isLoadingDocumentAccesses) && <Spinner />}
              {__("Update Access")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  );
}
