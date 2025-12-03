import { useState, useEffect } from "react";
import {
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconTrashCan,
  IconPlusLarge,
  IconPencil,
  IconSquareBehindSquare2,
  Label,
  Select,
  Badge,
  useConfirm,
  useDialogRef,
  Option,
  useToast,
  Checkbox,
  Table,
  Thead,
  Tbody,
  Tr,
  Th,
  Td,
  Input,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { formatDate } from "@probo/helpers";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { Controller } from "react-hook-form";
import { z } from "zod";
import { UnAuthenticatedError } from "@probo/relay";

interface APIKey {
  id: string;
  name: string;
  expiresAt: string;
  createdAt: string;
  organizations: APIKeyOrganization[];
}

interface APIKeyOrganization {
  organizationId: string;
  organizationName: string;
  role: string;
}

interface Organization {
  id: string;
  name: string;
  authStatus: "authenticated" | "unauthenticated" | "expired";
}

const createSchema = z.object({
  name: z.string().min(1, "Name is required"),
  expiresIn: z.enum(["1month", "3months", "6months", "1year"]),
  organizations: z
    .array(
      z.object({
        organizationId: z.string(),
        role: z.string(),
      })
    )
    .min(1, "At least one organization is required"),
});

type CreateFormData = z.infer<typeof createSchema>;

export default function APIKeysPage() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const confirm = useConfirm();
  const dialogRef = useDialogRef();
  const editDialogRef = useDialogRef();
  const keyDialogRef = useDialogRef();
  const [currentKey, setCurrentKey] = useState<string | null>(null);
  const [isLoadingKey, setIsLoadingKey] = useState(false);
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [selectedOrganizations, setSelectedOrganizations] = useState<string[]>(
    []
  );
  const [organizationRoles, setOrganizationRoles] = useState<
    Record<string, string>
  >({});
  const [editingAPIKey, setEditingAPIKey] = useState<APIKey | null>(null);
  const [editingName, setEditingName] = useState<string>("");
  const [error, setError] = useState<Error | null>(null);

  const { formState, handleSubmit, reset, control, setValue, register } =
    useFormWithSchema(createSchema, {
      defaultValues: {
        name: new Date().toISOString().split("T")[0],
        expiresIn: "1month",
        organizations: [],
      },
    });

  if (error) {
    throw error;
  }

  const fetchAPIKeys = async () => {
    try {
      const response = await fetch("/connect/api-keys", {
        credentials: "include",
      });
      if (!response.ok) {
        throw new Error("Failed to fetch API keys");
      }
      const data: { apiKeys: APIKey[] } = await response.json();
      setApiKeys(data.apiKeys);
    } catch (err) {
      console.error("Failed to fetch API keys:", err);
      toast({
        title: __("Error"),
        description: __("Failed to load API keys"),
        variant: "error",
      });
    }
  };

  useEffect(() => {
    const fetchData = async () => {
      try {
        const [apiKeysResponse, organizationsResponse] = await Promise.all([
          fetch("/connect/api-keys", { credentials: "include" }),
          fetch("/connect/organizations?role=OWNER", {
            credentials: "include",
          }),
        ]);

        if (
          apiKeysResponse.status === 401 ||
          organizationsResponse.status === 401
        ) {
          setError(new UnAuthenticatedError());
          return;
        }

        if (!apiKeysResponse.ok) {
          throw new Error("Failed to fetch API keys");
        }

        if (!organizationsResponse.ok) {
          throw new Error("Failed to fetch organizations");
        }

        const apiKeysData: { apiKeys: APIKey[] } = await apiKeysResponse.json();
        const orgsData: { organizations: Organization[] } =
          await organizationsResponse.json();

        const authenticatedOrgs = orgsData.organizations.filter(
          (org) => org.authStatus === "authenticated"
        );

        setApiKeys(apiKeysData.apiKeys);
        setOrganizations(authenticatedOrgs);
      } catch (err) {
        console.error("Failed to fetch data:", err);
        toast({
          title: __("Error"),
          description: __("Failed to load data"),
          variant: "error",
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [__, toast]);

  const handleCreate = async (formData: CreateFormData) => {
    const now = new Date();
    const expiresAt = new Date(now);

    switch (formData.expiresIn) {
      case "1month":
        expiresAt.setMonth(now.getMonth() + 1);
        break;
      case "3months":
        expiresAt.setMonth(now.getMonth() + 3);
        break;
      case "6months":
        expiresAt.setMonth(now.getMonth() + 6);
        break;
      case "1year":
        expiresAt.setFullYear(now.getFullYear() + 1);
        break;
    }

    setIsCreating(true);
    try {
      const response = await fetch("/connect/api-keys", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          name: formData.name,
          expiresAt: expiresAt.toISOString(),
          organizations: formData.organizations,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to create API key");
      }

      const data: { apiKey: APIKey; key: string } = await response.json();

      await fetchAPIKeys();
      dialogRef.current?.close();
      reset();
      setSelectedOrganizations([]);
      setOrganizationRoles({});
      setCurrentKey(data.key);
      keyDialogRef.current?.open();
      toast({
        title: __("Success"),
        description: __("API Key created successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: (error as Error).message,
        variant: "error",
      });
    } finally {
      setIsCreating(false);
    }
  };

  const handleEdit = (apiKey: APIKey) => {
    setEditingAPIKey(apiKey);
    setEditingName(apiKey.name);
    const orgIds = apiKey.organizations.map((org) => org.organizationId);
    const roles: Record<string, string> = {};
    apiKey.organizations.forEach((org) => {
      roles[org.organizationId] = org.role;
    });
    setSelectedOrganizations(orgIds);
    setOrganizationRoles(roles);
    editDialogRef.current?.open();
  };

  const handleUpdate = async () => {
    if (!editingAPIKey) return;

    setIsUpdating(true);
    try {
      const response = await fetch("/connect/api-keys", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          id: editingAPIKey.id,
          name: editingName,
          organizations: selectedOrganizations.map((id) => ({
            organizationId: id,
            role: organizationRoles[id] || "FULL",
          })),
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to update API key");
      }

      await fetchAPIKeys();
      editDialogRef.current?.close();
      setEditingAPIKey(null);
      setEditingName("");
      setSelectedOrganizations([]);
      setOrganizationRoles({});
      toast({
        title: __("Success"),
        description: __("API Key updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: (error as Error).message,
        variant: "error",
      });
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDelete = (id: string, name: string) => {
    confirm(
      async () => {
        setIsDeleting(true);
        try {
          const response = await fetch("/connect/api-keys", {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
            },
            credentials: "include",
            body: JSON.stringify({ id }),
          });

          if (!response.ok) {
            throw new Error("Failed to delete API key");
          }

          setApiKeys(apiKeys.filter((key) => key.id !== id));
          toast({
            title: __("Success"),
            description: __("API Key deleted successfully"),
            variant: "success",
          });
        } catch (error) {
          toast({
            title: __("Error"),
            description: (error as Error).message,
            variant: "error",
          });
          throw error;
        } finally {
          setIsDeleting(false);
        }
      },
      {
        message: __(
          `Are you sure you want to delete the API key "${name}"? This action cannot be undone.`
        ),
      }
    );
  };

  const handleShowToken = async (id: string) => {
    setIsLoadingKey(true);
    try {
      const response = await fetch(`/connect/api-keys/${id}`, {
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Failed to load API key");
      }

      const data: { key: string } = await response.json();
      setCurrentKey(data.key);
      keyDialogRef.current?.open();
    } catch {
      toast({
        title: __("Error"),
        description: __("Failed to load API key"),
        variant: "error",
      });
    } finally {
      setIsLoadingKey(false);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({
        title: __("Success"),
        description: __("API key copied to clipboard"),
        variant: "success",
      });
    } catch {
      toast({
        title: __("Error"),
        description: __("Failed to copy to clipboard"),
        variant: "error",
      });
    }
  };

  const isExpired = (expiresAt: string) => {
    return new Date(expiresAt) < new Date();
  };

  if (isLoading) {
    return (
      <div className="space-y-6 w-full py-6">
        <h1 className="text-3xl font-bold text-center">{__("API Keys")}</h1>
        <Card padded>
          <div className="text-center py-8">
            <p className="text-txt-tertiary">{__("Loading...")}</p>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 w-full py-6">
      <h1 className="text-3xl font-bold text-center">{__("API Keys")}</h1>

      <div className="space-y-4 w-full">
        {apiKeys.length === 0 ? (
          <Card padded>
            <div className="text-center py-8">
              <p className="text-txt-tertiary mb-4">
                {__("No API keys yet. Create one to get started.")}
              </p>
            </div>
          </Card>
        ) : (
          apiKeys.map((apiKey) => {
            const expired = isExpired(apiKey.expiresAt);

            return (
              <Card key={apiKey.id} padded className="w-full">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 space-y-2">
                    <div className="flex items-center gap-2">
                      <h3 className="text-lg font-semibold">{apiKey.name}</h3>
                      {expired ? (
                        <Badge variant="danger">{__("Expired")}</Badge>
                      ) : (
                        <Badge variant="success">{__("Active")}</Badge>
                      )}
                    </div>
                    <div className="flex items-center gap-4 text-sm text-txt-secondary">
                      <span>
                        {__("Created")}: {formatDate(apiKey.createdAt)}
                      </span>
                      <span>•</span>
                      <span>
                        {__("Expires")}: {formatDate(apiKey.expiresAt)}
                      </span>
                    </div>
                    {apiKey.organizations.length > 0 && (
                      <div className="flex flex-wrap gap-2 mt-2">
                        {apiKey.organizations.map((org) => (
                          <Badge key={org.organizationId} variant="neutral">
                            {org.organizationName} ({org.role})
                          </Badge>
                        ))}
                      </div>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="secondary"
                      onClick={() => handleShowToken(apiKey.id)}
                      disabled={isLoadingKey}
                      title={__("Show Token")}
                    >
                      {__("Show")}
                    </Button>
                    <Button
                      variant="secondary"
                      onClick={() => handleEdit(apiKey)}
                      disabled={isLoadingKey}
                      icon={IconPencil}
                      title={__("Edit")}
                    />
                    <Button
                      variant="danger"
                      onClick={() => handleDelete(apiKey.id, apiKey.name)}
                      disabled={isDeleting}
                      icon={IconTrashCan}
                      title={__("Delete")}
                    />
                  </div>
                </div>
              </Card>
            );
          })
        )}

        <Card padded>
          <h2 className="text-xl font-semibold mb-1">
            {__("Create an API key")}
          </h2>
          <p className="text-txt-tertiary mb-4">
            {__(
              "Generate a new API key for programmatic access to your organization"
            )}
          </p>
          <Button
            onClick={() => dialogRef.current?.open()}
            variant="quaternary"
            icon={IconPlusLarge}
            className="w-full"
          >
            {__("Create API Key")}
          </Button>
        </Card>
      </div>

      <Dialog ref={dialogRef} title={__("Create API Key")}>
        <form onSubmit={handleSubmit(handleCreate)}>
          <DialogContent padded className="space-y-4">
            <Field error={formState.errors.name?.message}>
              <Label>{__("Name")}</Label>
              <Input
                {...register("name")}
                placeholder={__("e.g., Production API Key")}
              />
            </Field>
            <Field error={formState.errors.expiresIn?.message}>
              <Label>{__("Expires In")}</Label>
              <Controller
                control={control}
                name="expiresIn"
                render={({ field }) => (
                  <Select
                    {...field}
                    onValueChange={field.onChange}
                    value={field.value}
                  >
                    <Option value="1month">{__("1 Month")}</Option>
                    <Option value="3months">{__("3 Months")}</Option>
                    <Option value="6months">{__("6 Months")}</Option>
                    <Option value="1year">{__("1 Year")}</Option>
                  </Select>
                )}
              />
            </Field>
            <Field error={formState.errors.organizations?.message}>
              <div className="flex justify-between items-center mb-4">
                <h4 className="font-medium text-txt-primary">
                  {__("Organizations")}
                </h4>
                {organizations.length > 0 && (
                  <Button
                    type="button"
                    variant="tertiary"
                    onClick={() => {
                      const allSelected =
                        selectedOrganizations.length === organizations.length;
                      if (allSelected) {
                        setSelectedOrganizations([]);
                        setOrganizationRoles({});
                        setValue("organizations", []);
                      } else {
                        const allOrgIds = organizations.map((org) => org.id);
                        const newRoles: Record<string, string> = {};
                        allOrgIds.forEach((id) => {
                          newRoles[id] = organizationRoles[id] || "FULL";
                        });
                        setSelectedOrganizations(allOrgIds);
                        setOrganizationRoles(newRoles);
                        setValue(
                          "organizations",
                          allOrgIds.map((id) => ({
                            organizationId: id,
                            role: newRoles[id],
                          }))
                        );
                      }
                    }}
                    className="text-xs h-7 min-h-7"
                  >
                    {selectedOrganizations.length === organizations.length
                      ? __("Clear All")
                      : __("Select All")}
                  </Button>
                )}
              </div>
              {organizations.length === 0 ? (
                <div className="text-center text-txt-tertiary py-8">
                  {__("No organizations available")}
                </div>
              ) : (
                <div className="bg-bg-secondary rounded-lg overflow-hidden">
                  <Table>
                    <Thead>
                      <Tr>
                        <Th>{__("Name")}</Th>
                        <Th width={180}>{__("Role")}</Th>
                        <Th width={100}>
                          <div className="flex justify-end">{__("Access")}</div>
                        </Th>
                      </Tr>
                    </Thead>
                    <Tbody>
                      {organizations.map((org) => (
                        <Tr key={org.id}>
                          <Td>
                            <div className="font-medium text-txt-primary">
                              {org.name}
                            </div>
                          </Td>
                          <Td>
                            <div className="min-h-[36px] flex items-center">
                              {selectedOrganizations.includes(org.id) ? (
                                <Select
                                  value={organizationRoles[org.id] || "FULL"}
                                  onValueChange={(role) => {
                                    const newRoles = {
                                      ...organizationRoles,
                                      [org.id]: role,
                                    };
                                    setOrganizationRoles(newRoles);
                                    setValue(
                                      "organizations",
                                      selectedOrganizations.map((id) => ({
                                        organizationId: id,
                                        role: newRoles[id] || "FULL",
                                      }))
                                    );
                                  }}
                                >
                                  <Option value="FULL">{__("Full")}</Option>
                                </Select>
                              ) : (
                                <span className="text-txt-tertiary">—</span>
                              )}
                            </div>
                          </Td>
                          <Td>
                            <div className="flex justify-end">
                              <Checkbox
                                checked={selectedOrganizations.includes(org.id)}
                                onChange={(checked: boolean) => {
                                  let newSelected: string[];
                                  const newRoles = { ...organizationRoles };
                                  if (checked) {
                                    newSelected = [
                                      ...selectedOrganizations,
                                      org.id,
                                    ];
                                    if (!newRoles[org.id]) {
                                      newRoles[org.id] = "FULL";
                                    }
                                  } else {
                                    newSelected = selectedOrganizations.filter(
                                      (id) => id !== org.id
                                    );
                                    delete newRoles[org.id];
                                  }
                                  setSelectedOrganizations(newSelected);
                                  setOrganizationRoles(newRoles);
                                  setValue(
                                    "organizations",
                                    newSelected.map((id) => ({
                                      organizationId: id,
                                      role: newRoles[id] || "FULL",
                                    }))
                                  );
                                }}
                              />
                            </div>
                          </Td>
                        </Tr>
                      ))}
                    </Tbody>
                  </Table>
                </div>
              )}
            </Field>
          </DialogContent>
          <DialogFooter>
            <Button type="submit" disabled={isCreating}>
              {isCreating ? __("Creating...") : __("Create")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>

      <Dialog ref={editDialogRef} title={__("Edit API Key")}>
        <DialogContent padded className="space-y-6">
          <Field>
            <Label>{__("Name")}</Label>
            <Input
              value={editingName}
              onChange={(e) => setEditingName(e.target.value)}
              placeholder={__("API Key Name")}
            />
          </Field>

          <div>
            <div className="flex justify-between items-center mb-4">
              <h4 className="font-medium text-txt-primary">
                {__("Organizations")}
              </h4>
              {organizations.length > 0 && (
                <Button
                  type="button"
                  variant="tertiary"
                  onClick={() => {
                    const allSelected =
                      selectedOrganizations.length === organizations.length;
                    if (allSelected) {
                      setSelectedOrganizations([]);
                      setOrganizationRoles({});
                    } else {
                      const allOrgIds = organizations.map((org) => org.id);
                      const newRoles: Record<string, string> = {};
                      allOrgIds.forEach((id) => {
                        newRoles[id] = organizationRoles[id] || "FULL";
                      });
                      setSelectedOrganizations(allOrgIds);
                      setOrganizationRoles(newRoles);
                    }
                  }}
                  className="text-xs h-7 min-h-7"
                >
                  {selectedOrganizations.length === organizations.length
                    ? __("Clear All")
                    : __("Select All")}
                </Button>
              )}
            </div>
            {organizations.length === 0 ? (
              <div className="text-center text-txt-tertiary py-8">
                {__("No organizations available")}
              </div>
            ) : (
              <div className="bg-bg-secondary rounded-lg overflow-hidden">
                <Table>
                  <Thead>
                    <Tr>
                      <Th>{__("Name")}</Th>
                      <Th width={180}>{__("Role")}</Th>
                      <Th width={100}>
                        <div className="flex justify-end">{__("Access")}</div>
                      </Th>
                    </Tr>
                  </Thead>
                  <Tbody>
                    {organizations.map((org) => (
                      <Tr key={org.id}>
                        <Td>
                          <div className="font-medium text-txt-primary">
                            {org.name}
                          </div>
                        </Td>
                        <Td>
                          <div className="min-h-[36px] flex items-center">
                            {selectedOrganizations.includes(org.id) ? (
                              <Select
                                value={organizationRoles[org.id] || "FULL"}
                                onValueChange={(role) => {
                                  const newRoles = {
                                    ...organizationRoles,
                                    [org.id]: role,
                                  };
                                  setOrganizationRoles(newRoles);
                                }}
                              >
                                <Option value="FULL">{__("Full")}</Option>
                              </Select>
                            ) : (
                              <span className="text-txt-tertiary">—</span>
                            )}
                          </div>
                        </Td>
                        <Td>
                          <div className="flex justify-end">
                            <Checkbox
                              checked={selectedOrganizations.includes(org.id)}
                              onChange={(checked: boolean) => {
                                let newSelected: string[];
                                const newRoles = { ...organizationRoles };
                                if (checked) {
                                  newSelected = [
                                    ...selectedOrganizations,
                                    org.id,
                                  ];
                                  if (!newRoles[org.id]) {
                                    newRoles[org.id] = "FULL";
                                  }
                                } else {
                                  newSelected = selectedOrganizations.filter(
                                    (id) => id !== org.id
                                  );
                                  delete newRoles[org.id];
                                }
                                setSelectedOrganizations(newSelected);
                                setOrganizationRoles(newRoles);
                              }}
                            />
                          </div>
                        </Td>
                      </Tr>
                    ))}
                  </Tbody>
                </Table>
              </div>
            )}
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            onClick={handleUpdate}
            disabled={
              isUpdating ||
              (selectedOrganizations.length === 0 && !editingName.trim())
            }
          >
            {isUpdating ? __("Updating...") : __("Update")}
          </Button>
        </DialogFooter>
      </Dialog>

      <Dialog ref={keyDialogRef} title={__("API Key")}>
        <DialogContent padded className="space-y-4">
          <p className="text-sm text-txt-tertiary">
            {__("Please save this API key securely.")}
          </p>
          <div className="bg-gray-100 p-4 rounded-lg flex items-center gap-2">
            <code className="text-sm font-mono break-all flex-1">
              {currentKey || ""}
            </code>
            <Button
              onClick={() => {
                if (currentKey) {
                  copyToClipboard(currentKey);
                }
              }}
              variant="secondary"
              disabled={!currentKey}
              title={__("Copy to Clipboard")}
            >
              <IconSquareBehindSquare2 size={16} />
            </Button>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            onClick={() => {
              keyDialogRef.current?.close();
              setCurrentKey(null);
            }}
          >
            {__("Done")}
          </Button>
        </DialogFooter>
      </Dialog>
    </div>
  );
}
