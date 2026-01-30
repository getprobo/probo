import { useTranslate } from "@probo/i18n";
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
import { type ChangeEventHandler, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { OrganizationFormFragment$key } from "#/__generated__/iam/OrganizationFormFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const fragment = graphql`
  fragment OrganizationFormFragment on Organization {
    id
    name @required(action: THROW)
    logoUrl
    horizontalLogoUrl
    description
    websiteUrl
    email
    headquarterAddress
    canUpdate: permission(action: "iam:organization:update")
  }
`;

const updateOrganizationMutation = graphql`
  mutation OrganizationForm_updateMutation($input: UpdateOrganizationInput!) {
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
  mutation OrganizationForm_deleteHorizontalLogoMutation(
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
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

export function OrganizationForm(props: {
  fKey: OrganizationFormFragment$key;
}) {
  const { fKey } = props;
  const { __ } = useTranslate();
  const deleteDialogRef = useDialogRef();

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [horizontalLogoPreview, setHorizontalLogoPreview] = useState<
    string | null
  >(null);

  const { canUpdate, ...organization }
    = useFragment<OrganizationFormFragment$key>(fragment, fKey);

  const [updateOrganization, isUpdatingOrganization] = useMutationWithToasts(
    updateOrganizationMutation,
    {
      successMessage: __("Organization updated successfully"),
      errorMessage: __("Failed to update organization"),
    },
  );
  const [deleteHorizontalLogo, isDeletingHorizontalLogo]
    = useMutationWithToasts(deleteHorizontalLogoMutation, {
      successMessage: __("Horizontal logo deleted successfully"),
      errorMessage: __("Failed to delete horizontal logo"),
    });

  const { formState, handleSubmit, register } = useFormWithSchema(
    organizationSchema,
    {
      defaultValues: {
        name: organization.name,
        description: organization.description || "",
        websiteUrl: organization.websiteUrl || "",
        email: organization.email || "",
        headquarterAddress: organization.headquarterAddress || "",
      },
    },
  );

  const handleLogoChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
      setLogoPreview(reader.result as string);
    };
    reader.readAsDataURL(file);

    void updateOrganization({
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
    e,
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = () => {
      setHorizontalLogoPreview(reader.result as string);
    };
    reader.readAsDataURL(file);

    void updateOrganization({
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

  const handleDeleteHorizontalLogo = async () => {
    await deleteHorizontalLogo({
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

  const onSubmit = handleSubmit(async (data: OrganizationFormData) => {
    await updateOrganization({
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
    });
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Organization details")}</h2>
        {formState.isSubmitting && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <div>
          <Label>{__("Organization logo")}</Label>
          <div className="flex w-max items-center gap-4">
            <Avatar
              className={logoPreview || organization.logoUrl ? "bg-transparent" : undefined}
              src={logoPreview || organization.logoUrl}
              name={organization.name}
              size="xl"
            />
            {canUpdate && (
              <FileButton
                disabled={formState.isSubmitting || isUpdatingOrganization}
                onChange={handleLogoChange}
                variant="secondary"
                className="ml-auto"
                accept="image/png,image/jpeg,image/jpg,image/svg+xml"
              >
                {isUpdatingOrganization
                  ? __("Uploading...")
                  : __("Change logo")}
              </FileButton>
            )}
          </div>
        </div>
        <div>
          <Label>{__("Horizontal logo")}</Label>
          <p className="text-sm text-txt-tertiary mb-2">
            {__(
              "Upload a horizontal version of your logo for use in documents",
            )}
          </p>
          <div className="flex items-center gap-4">
            {(horizontalLogoPreview || organization.horizontalLogoUrl) && (
              <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                <img
                  src={
                    horizontalLogoPreview
                    || organization.horizontalLogoUrl
                    || undefined
                  }
                  alt={__("Horizontal logo")}
                  className="h-12 max-w-xs object-contain"
                />
              </div>
            )}
            {canUpdate && (
              <FileButton
                disabled={formState.isSubmitting || isUpdatingOrganization}
                onChange={handleHorizontalLogoChange}
                variant="secondary"
                accept="image/png,image/jpeg,image/jpg,image/svg+xml"
              >
                {isUpdatingOrganization
                  ? __("Uploading...")
                  : horizontalLogoPreview || organization.horizontalLogoUrl
                    ? __("Change horizontal logo")
                    : __("Upload horizontal logo")}
              </FileButton>
            )}
            {canUpdate && organization.horizontalLogoUrl && (
              <Dialog
                ref={deleteDialogRef}
                trigger={(
                  <Button
                    type="button"
                    variant="quaternary"
                    icon={IconTrashCan}
                    aria-label={__("Delete horizontal logo")}
                    className="text-red-600 hover:text-red-700"
                  />
                )}
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
                    onClick={() => void handleDeleteHorizontalLogo()}
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
          readOnly={formState.isSubmitting || !canUpdate}
          name="name"
          type="text"
          label={__("Organization name")}
          placeholder={__("Organization name")}
        />
        <div>
          <Label>{__("Description")}</Label>
          <Textarea
            {...register("description")}
            readOnly={formState.isSubmitting || !canUpdate}
            name="description"
            placeholder={__("Brief description of your organization")}
            rows={3}
          />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <Field
            {...register("websiteUrl")}
            readOnly={formState.isSubmitting || !canUpdate}
            name="websiteUrl"
            type="url"
            label={__("Website URL")}
            placeholder={__("https://example.com")}
          />
          <Field
            {...register("email")}
            readOnly={formState.isSubmitting || !canUpdate}
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
            readOnly={formState.isSubmitting || !canUpdate}
            name="headquarterAddress"
            placeholder={__("123 Main St, City, Country")}
          />
        </div>

        {formState.isDirty && canUpdate && (
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
    </form>
  );
}
