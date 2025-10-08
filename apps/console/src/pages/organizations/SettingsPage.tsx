import {
  ActionDropdown,
  Avatar,
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  DropdownItem,
  Field,
  FileButton,
  IconTrashCan,
  Label,
  PageHeader,
  Spinner,
  Textarea,
  useConfirm,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import type { PreloadedQuery } from "react-relay";
import type { OrganizationGraph_ViewQuery } from "/hooks/graph/__generated__/OrganizationGraph_ViewQuery.graphql";
import { useFragment, useMutation, usePreloadedQuery } from "react-relay";
import { organizationViewQuery } from "/hooks/graph/OrganizationGraph";
import { graphql } from "relay-runtime";
import type {
  SettingsPageFragment$data,
  SettingsPageFragment$key,
} from "./__generated__/SettingsPageFragment.graphql";
import { useState, type ChangeEventHandler, useEffect } from "react";
import { sprintf } from "@probo/helpers";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { z } from "zod";
import type { NodeOf } from "/types";
import clsx from "clsx";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { InviteUserDialog } from "/components/organizations/InviteUserDialog";
import { useDeleteOrganizationMutation } from "/hooks/graph/OrganizationGraph";
import { useNavigate } from "react-router";
import { DeleteOrganizationDialog } from "/components/organizations/DeleteOrganizationDialog";
import { CustomDomainManager } from "/components/customDomains/CustomDomainManager";

const organizationSchema = z.object({
  name: z.string().min(1, "Organization name is required"),
  description: z.string().optional(),
  websiteUrl: z.string().optional(),
  email: z.string().optional(),
  headquarterAddress: z.string().optional(),
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

type Props = {
  queryRef: PreloadedQuery<OrganizationGraph_ViewQuery>;
};

const organizationFragment = graphql`
  fragment SettingsPageFragment on Organization {
    id
    name
    logoUrl
    horizontalLogoUrl
    description
    websiteUrl
    email
    headquarterAddress
    customDomain {
      id
      domain
      sslStatus
      dnsRecords {
        type
        name
        value
        ttl
        purpose
      }
      createdAt
      updatedAt
      sslExpiresAt
    }
    users(first: 100) {
      edges {
        node {
          id
          fullName
          email
          createdAt
        }
      }
    }
    connectors(first: 100) {
      edges {
        node {
          id
          name
          type
          createdAt
        }
      }
    }
  }
`;

const updateOrganizationMutation = graphql`
  mutation SettingsPage_UpdateMutation($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      organization {
        id
        name
        logoUrl
        horizontalLogoUrl
        description
        websiteUrl
        email
        headquarterAddress
      }
    }
  }
`;

const deleteHorizontalLogoMutation = graphql`
  mutation SettingsPage_DeleteHorizontalLogoMutation($input: DeleteOrganizationHorizontalLogoInput!) {
    deleteOrganizationHorizontalLogo(input: $input) {
      organization {
        id
        horizontalLogoUrl
      }
    }
  }
`;

export default function SettingsPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const organizationKey = usePreloadedQuery(
    organizationViewQuery,
    queryRef
  ).node;
  const { toast } = useToast();
  const organization = useFragment<SettingsPageFragment$key>(
    organizationFragment,
    organizationKey
  );
  const [updateOrganization] = useMutation(updateOrganizationMutation);
  const [deleteHorizontalLogo, isDeletingHorizontalLogo] = useMutationWithToasts(
    deleteHorizontalLogoMutation,
    {
      successMessage: __("Horizontal logo deleted successfully"),
      errorMessage: __("Failed to delete horizontal logo"),
    }
  );
  const [deleteOrganization, isDeleting] = useDeleteOrganizationMutation();
  const users = organization.users.edges.map((edge) => edge.node);

  const { formState, handleSubmit, register, reset } = useFormWithSchema(
    organizationSchema,
    {
      defaultValues: {
        name: organization.name || "",
        description: organization.description || "",
        websiteUrl: organization.websiteUrl || "",
        email: organization.email || "",
        headquarterAddress: organization.headquarterAddress || "",
      },
    }
  );

  useEffect(() => {
    reset({
      name: organization.name || "",
      description: organization.description || "",
      websiteUrl: organization.websiteUrl || "",
      email: organization.email || "",
      headquarterAddress: organization.headquarterAddress || "",
    });
  }, [organization, reset]);

  const onSubmit = handleSubmit((data: OrganizationFormData) => {
    updateOrganization({
      variables: {
        input: {
          organizationId: organization.id,
          name: data.name,
          description: data.description || null,
          websiteUrl: data.websiteUrl || null,
          email: data.email || null,
          headquarterAddress: data.headquarterAddress || null,
        },
      },
      onError() {
        toast({
          title: __("Error"),
          description: __("Failed to update organization."),
          variant: "error",
        });
      },
      onCompleted() {
        toast({
          title: __("Organization updated"),
          description: __(
            "Your organization details have been updated successfully."
          ),
          variant: "success",
        });
      },
    });
  });

  const updateOrganizationLogo: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) {
      return;
    }
    updateOrganization({
      variables: {
        input: {
          organizationId: organization.id,
          logo: null,
        },
      },
      uploadables: {
        "input.logo": file,
      },
      onError() {
        toast({
          title: __("Error"),
          description: __("Failed to update logo"),
          variant: "error",
        });
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Your organization logo has been updated successfully."),
          variant: "success",
        });
      },
    });
  };

  const updateHorizontalLogo: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) {
      return;
    }
    updateOrganization({
      variables: {
        input: {
          organizationId: organization.id,
          horizontalLogoFile: null,
        },
      },
      uploadables: {
        "input.horizontalLogoFile": file,
      },
      onError() {
        toast({
          title: __("Error"),
          description: __("Failed to update horizontal logo."),
          variant: "error",
        });
      },
      onCompleted() {
        toast({
          title: __("Success"),
          description: __("Your organization horizontal logo has been updated successfully."),
          variant: "success",
        });
      },
    });
  };

  const deleteDialogRef = useDialogRef();

  const handleDeleteHorizontalLogo = () => {
    deleteHorizontalLogo({
      variables: {
        input: {
          organizationId: organization.id,
        },
      },
      onSuccess: () => {
        deleteDialogRef.current?.close();
      },
    });
  };

  const handleDeleteOrganization = () => {
    return deleteOrganization({
      variables: {
        input: {
          organizationId: organization.id,
        },
        connections: [],
      },
      onSuccess: () => {
        navigate("/", { replace: true });
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader title={__("Settings")} />

      {/* Organization settings */}
      <form onSubmit={onSubmit} className="space-y-6">
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-medium">
              {__("Organization details")}
            </h2>
            {formState.isSubmitting && <Spinner />}
          </div>
          <Card padded className="space-y-4">
            <div>
              <Label>{__("Organization logo")}</Label>
              <div className="flex w-max items-center gap-4">
                <Avatar
                  src={organization.logoUrl}
                  name={organization.name}
                  size="xl"
                />
                <FileButton
                  disabled={formState.isSubmitting}
                  onChange={updateOrganizationLogo}
                  variant="secondary"
                  className="ml-auto"
                  accept="image/png,image/jpeg,image/jpg"
                >
                  {__("Change logo")}
                </FileButton>
              </div>
            </div>
            <div>
              <Label>{__("Horizontal logo")}</Label>
              <p className="text-sm text-txt-tertiary mb-2">
                {__("Upload a horizontal version of your logo for use in documents")}
              </p>
              <div className="flex items-center gap-4">
                {organization.horizontalLogoUrl && (
                  <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                    <img
                      src={organization.horizontalLogoUrl}
                      alt={__("Horizontal logo")}
                      className="h-12 max-w-xs object-contain"
                    />
                  </div>
                )}
                <FileButton
                  disabled={formState.isSubmitting}
                  onChange={updateHorizontalLogo}
                  variant="secondary"
                  accept="image/png,image/jpeg,image/jpg"
                >
                  {organization.horizontalLogoUrl ? __("Change horizontal logo") : __("Upload horizontal logo")}
                </FileButton>
                {organization.horizontalLogoUrl && (
                  <Dialog
                    ref={deleteDialogRef}
                    trigger={
                      <Button
                        variant="quaternary"
                        icon={IconTrashCan}
                        aria-label={__("Delete horizontal logo")}
                        className="text-red-600 hover:text-red-700"
                      />
                    }
                    title={__("Delete Horizontal Logo")}
                    className="max-w-md"
                  >
                    <DialogContent padded>
                      <p className="text-txt-secondary">
                        {__("Are you sure you want to delete the horizontal logo?")}
                      </p>
                      <p className="text-txt-secondary mt-2">
                        {__("This action cannot be undone.")}
                      </p>
                    </DialogContent>

                    <DialogFooter>
                      <Button
                        variant="danger"
                        onClick={handleDeleteHorizontalLogo}
                        disabled={isDeletingHorizontalLogo}
                        icon={isDeletingHorizontalLogo ? Spinner : IconTrashCan}
                      >
                        {isDeletingHorizontalLogo ? __("Deleting...") : __("Delete")}
                      </Button>
                    </DialogFooter>
                  </Dialog>
                )}
              </div>
            </div>
            <Field
              {...register("name")}
              readOnly={formState.isSubmitting}
              name="name"
              type="text"
              label={__("Organization name")}
              placeholder={__("Organization name")}
            />
            <div>
              <Label>{__("Description")}</Label>
              <Textarea
                {...register("description")}
                readOnly={formState.isSubmitting}
                name="description"
                placeholder={__("Brief description of your organization")}
                rows={3}
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field
                {...register("websiteUrl")}
                readOnly={formState.isSubmitting}
                name="websiteUrl"
                type="url"
                label={__("Website URL")}
                placeholder={__("https://example.com")}
              />
              <Field
                {...register("email")}
                readOnly={formState.isSubmitting}
                name="email"
                type="email"
                label={__("Email")}
                placeholder={__("contact@example.com")}
              />
            </div>
            <div>
              <Label>{__("Headquarter Address")}</Label>
              <Textarea
                {...register("headquarterAddress")}
                readOnly={formState.isSubmitting}
                name="headquarterAddress"
                placeholder={__("123 Main St, City, Country")}
              />
            </div>

            {formState.isDirty && (
              <div className="flex justify-end pt-6">
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting
                    ? __("Updating...")
                    : __("Update Organization")}
                </Button>
              </div>
            )}
          </Card>
        </div>
      </form>

      {/* Integrations */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Workspace members")}</h2>
          <InviteUserDialog>
            <Button variant="secondary">{__("Invite member")}</Button>
          </InviteUserDialog>
        </div>
        <Card className="divide-y divide-border-solid">
          {users.map((user) => (
            <UserRow key={user.id} user={user} />
          ))}
        </Card>
      </div>

      {/* Integrations */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Integrations")}</h2>
        <Card padded>
          <Connectors
            organizationId={organization.id}
            connectors={organization.connectors.edges.map((edge) => edge.node)}
          />
        </Card>
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Custom Domain")}</h2>
        <CustomDomainManager
          organizationId={organization.id}
          customDomain={organization.customDomain}
        />
      </div>

      <div className="space-y-4">
        <h2 className="text-base font-medium text-red-600">
          {__("Danger Zone")}
        </h2>
        <Card padded className="border-red-200 flex items-center gap-3">
          <div className="mr-auto">
            <h3 className="text-base font-semibold text-red-700">
              {__("Delete Organization")}
            </h3>
            <p className="text-sm text-txt-tertiary">
              {__("Permanently delete this organization and all its data.")}{" "}
              <span className="text-red-600 font-medium">
                {__("This action cannot be undone.")}
              </span>
            </p>
          </div>
          <DeleteOrganizationDialog
            organizationName={organization.name}
            onConfirm={handleDeleteOrganization}
            isDeleting={isDeleting}
          >
            <Button variant="danger" icon={IconTrashCan} disabled={isDeleting}>
              {isDeleting ? __("Deleting...") : __("Delete Organization")}
            </Button>
          </DeleteOrganizationDialog>
        </Card>
      </div>
    </div>
  );
}

function Connectors(props: {
  organizationId: string;
  connectors: NodeOf<SettingsPageFragment$data["connectors"]>[];
}) {
  const { __, dateTimeFormat } = useTranslate();
  const fakeconnectors = [
    {
      id: "github",
      name: "GitHub",
      type: "oauth2",
      createdAt: new Date(),
    },
  ] satisfies typeof props.connectors;
  const connectors = [
    {
      id: "github",
      name: "GitHub",
      type: "oauth2",
      description: __("Connect to GitHub repositories and issues"),
      ...fakeconnectors.find((connector) => connector.id === "github"),
    },
    {
      id: "slack",
      name: "Slack",
      type: "oauth2",
      description: __("Connect to Slack workspace and channels"),
      ...fakeconnectors.find((connector) => connector.id === "slack"),
    },
  ];

  const getUrl = (connectorId: string) => {
    const baseUrl = import.meta.env.VITE_API_URL || window.location.origin;
    const url = new URL("/api/console/v1/connectors/initiate", baseUrl);
    url.searchParams.append("organization_id", props.organizationId);
    url.searchParams.append("connector_id", connectorId);
    url.searchParams.append("continue", window.location.href);
    return url.toString();
  };

  return (
    <div className="space-y-2">
      {connectors.map((connector) => (
        <Card key={connector.id} padded className="flex items-center gap-3">
          <div>
            <img src={`/${connector.id}.png`} alt="" />
          </div>
          <div className="mr-auto">
            <h3 className="text-base font-semibold">{connector.name}</h3>
            <p className="text-sm text-txt-tertiary">
              {connector.createdAt
                ? sprintf(
                    __("Connected on %s"),
                    dateTimeFormat(connector.createdAt)
                  )
                : connector.description}
            </p>
          </div>
          {connector.createdAt ? (
            <div>
              <Badge variant="success" size="md">
                {__("Connected")}
              </Badge>
            </div>
          ) : (
            <Button variant="secondary" asChild>
              <a href={getUrl(connector.id)}>{__("Connect")}</a>
            </Button>
          )}
        </Card>
      ))}
    </div>
  );
}

const removeUserMutation = graphql`
  mutation SettingsPage_RemoveUserMutation($input: RemoveUserInput!) {
    removeUser(input: $input) {
      success
    }
  }
`;

function UserRow(props: { user: NodeOf<SettingsPageFragment$data["users"]> }) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const [removeUser, isRemoving] = useMutationWithToasts(removeUserMutation, {
    successMessage: sprintf(
      __("User %s removed successfully"),
      props.user.fullName
    ),
    errorMessage: sprintf(__("Failed to remove user %s"), props.user.fullName),
  });
  const confirm = useConfirm();
  const [isRemoved, setIsRemoved] = useState(false);

  if (isRemoved) {
    return null;
  }

  const onRemove = async () => {
    confirm(
      () => {
        return removeUser({
          variables: {
            input: {
              userId: props.user.id,
              organizationId: organizationId,
            },
          },
          onSuccess: () => {
            setIsRemoved(true);
          },
        });
      },
      {
        message: sprintf(
          __("Are you sure you want to remove %s?"),
          props.user.fullName
        ),
      }
    );
  };

  return (
    <div
      className={clsx(
        "flex justify-between items-center py-4 px-4",
        isRemoving && "opacity-60 pointer-events-none"
      )}
    >
      <div className="flex items-center gap-4">
        <Avatar name={props.user.fullName} size="l" />
        <div>
          <h3 className="text-base font-semibold">{props.user.fullName}</h3>
          <p className="text-sm text-txt-tertiary">{props.user.email}</p>
        </div>
      </div>
      <div className="flex gap-2 items-center">
        <Badge>{__("Owner")}</Badge>
        {isRemoving ? (
          <Spinner size={16} />
        ) : (
          <ActionDropdown>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={onRemove}
            >
              {__("Remove")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>
    </div>
  );
}
