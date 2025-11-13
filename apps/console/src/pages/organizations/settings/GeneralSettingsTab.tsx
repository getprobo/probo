import { useState, useRef, useEffect, type ChangeEventHandler } from "react";
import { useOutletContext, useNavigate } from "react-router";
import { useFragment, graphql } from "react-relay";
import {
  Avatar,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  FileButton,
  IconTrashCan,
  Label,
  Spinner,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { z } from "zod";
import type { GeneralSettingsTabFragment$key } from "./__generated__/GeneralSettingsTabFragment.graphql";
import { DeleteOrganizationDialog } from "/components/organizations/DeleteOrganizationDialog";
import { useDeleteOrganizationMutation } from "/hooks/graph/OrganizationGraph";

const generalSettingsTabFragment = graphql`
  fragment GeneralSettingsTabFragment on Organization {
    id
    name
    logoUrl
    horizontalLogoUrl
    description
    websiteUrl
    email
    headquarterAddress
    slackId
    createdAt
    updatedAt
  }
`;

const updateOrganizationMutation = graphql`
  mutation GeneralSettingsTab_UpdateMutation($input: UpdateOrganizationInput!) {
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
        slackId
      }
    }
  }
`;

const deleteHorizontalLogoMutation = graphql`
  mutation GeneralSettingsTab_DeleteHorizontalLogoMutation(
    $input: DeleteOrganizationHorizontalLogoInput!
  ) {
    deleteOrganizationHorizontalLogo(input: $input) {
      organization {
        id
        horizontalLogoUrl
      }
    }
  }
`;

const organizationSchema = z.object({
  name: z.string().min(1, "Organization name is required"),
  description: z.string().optional(),
  websiteUrl: z.string().optional(),
  email: z.string().optional(),
  headquarterAddress: z.string().optional(),
  slackId: z.string().optional(),
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

type OutletContext = {
  organization: GeneralSettingsTabFragment$key;
};

export default function GeneralSettingsTab() {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { organization: organizationKey } = useOutletContext<OutletContext>();
  const organization = useFragment(generalSettingsTabFragment, organizationKey);
  const deleteDialogRef = useDialogRef();

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [horizontalLogoPreview, setHorizontalLogoPreview] = useState<
    string | null
  >(null);

  const [updateOrganization, isUpdatingOrganization] = useMutationWithToasts(
    updateOrganizationMutation,
    {
      successMessage: __("Organization updated successfully"),
      errorMessage: __("Failed to update organization"),
    }
  );
  const [deleteHorizontalLogo, isDeletingHorizontalLogo] =
    useMutationWithToasts(deleteHorizontalLogoMutation, {
      successMessage: __("Horizontal logo deleted successfully"),
      errorMessage: __("Failed to delete horizontal logo"),
    });
  const [deleteOrganization, isDeletingOrganization] =
    useDeleteOrganizationMutation();

  const { formState, handleSubmit, register, reset } = useFormWithSchema(
    organizationSchema,
    {
      defaultValues: {
        name: organization.name || "",
        description: organization.description || "",
        websiteUrl: organization.websiteUrl || "",
        email: organization.email || "",
        headquarterAddress: organization.headquarterAddress || "",
        slackId: organization.slackId || "",
      },
    }
  );

  const prevOrgDataRef = useRef({
    name: organization.name,
    description: organization.description,
    websiteUrl: organization.websiteUrl,
    email: organization.email,
    headquarterAddress: organization.headquarterAddress,
    slackId: organization.slackId,
  });

  useEffect(() => {
    const prevData = prevOrgDataRef.current;
    const currentData = {
      name: organization.name,
      description: organization.description,
      websiteUrl: organization.websiteUrl,
      email: organization.email,
      headquarterAddress: organization.headquarterAddress,
      slackId: organization.slackId,
    };

    if (JSON.stringify(prevData) !== JSON.stringify(currentData)) {
      reset({
        name: organization.name || "",
        description: organization.description || "",
        websiteUrl: organization.websiteUrl || "",
        email: organization.email || "",
        headquarterAddress: organization.headquarterAddress || "",
        slackId: organization.slackId || "",
      });
      prevOrgDataRef.current = currentData;
    }
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
          slackId: data.slackId || null,
        },
      },
    });
  });

  const handleLogoChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
      setLogoPreview(reader.result as string);
    };
    reader.readAsDataURL(file);

    updateOrganization({
      variables: {
        input: {
          organizationId: organization.id,
          logoFile: null,
        },
      },
      uploadables: {
        "input.logoFile": file,
      },
      onCompleted: () => {
        setLogoPreview(null);
      },
    });
  };

  const handleHorizontalLogoChange: ChangeEventHandler<HTMLInputElement> = (
    e
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
      setHorizontalLogoPreview(reader.result as string);
    };
    reader.readAsDataURL(file);

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
      onCompleted: () => {
        setHorizontalLogoPreview(null);
      },
    });
  };

  const handleDeleteHorizontalLogo = () => {
    deleteHorizontalLogo({
      variables: {
        input: {
          organizationId: organization.id,
        },
      },
      onCompleted: () => {
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
                src={logoPreview || organization.logoUrl}
                name={organization.name}
                size="xl"
              />
              <FileButton
                disabled={formState.isSubmitting || isUpdatingOrganization}
                onChange={handleLogoChange}
                variant="secondary"
                className="ml-auto"
                accept="image/png,image/jpeg,image/jpg"
              >
                {isUpdatingOrganization
                  ? __("Uploading...")
                  : __("Change logo")}
              </FileButton>
            </div>
          </div>
          <div>
            <Label>{__("Horizontal logo")}</Label>
            <p className="text-sm text-txt-tertiary mb-2">
              {__(
                "Upload a horizontal version of your logo for use in documents"
              )}
            </p>
            <div className="flex items-center gap-4">
              {(horizontalLogoPreview || organization.horizontalLogoUrl) && (
                <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                  <img
                    src={
                      horizontalLogoPreview ||
                      organization.horizontalLogoUrl ||
                      undefined
                    }
                    alt={__("Horizontal logo")}
                    className="h-12 max-w-xs object-contain"
                  />
                </div>
              )}
              <FileButton
                disabled={formState.isSubmitting || isUpdatingOrganization}
                onChange={handleHorizontalLogoChange}
                variant="secondary"
                accept="image/png,image/jpeg,image/jpg"
              >
                {isUpdatingOrganization
                  ? __("Uploading...")
                  : horizontalLogoPreview || organization.horizontalLogoUrl
                    ? __("Change horizontal logo")
                    : __("Upload horizontal logo")}
              </FileButton>
              {organization.horizontalLogoUrl && (
                <Dialog
                  ref={deleteDialogRef}
                  trigger={
                    <Button
                      type="button"
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
                      {__(
                        "Are you sure you want to delete the horizontal logo?"
                      )}
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
                      {isDeletingHorizontalLogo
                        ? __("Deleting...")
                        : __("Delete")}
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
          <div>
            <Field
              {...register("slackId")}
              readOnly={formState.isSubmitting}
              name="slackId"
              type="text"
              label={__("Slack ID")}
              placeholder={__("C1234567890")}
              />
          </div>

          {formState.isDirty && (
            <div className="flex justify-end pt-6">
              <Button
                type="submit"
                disabled={formState.isSubmitting || isUpdatingOrganization}
              >
                {formState.isSubmitting || isUpdatingOrganization
                  ? __("Updating...")
                  : __("Update Organization")}
              </Button>
            </div>
          )}
        </Card>
      </div>

      <div className="space-y-4 mt-12">
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
            isDeleting={isDeletingOrganization}
          >
            <Button
              variant="danger"
              icon={IconTrashCan}
              disabled={isDeletingOrganization}
            >
              {__("Delete Organization")}
            </Button>
          </DeleteOrganizationDialog>
        </Card>
      </div>
    </form>
  );
}
