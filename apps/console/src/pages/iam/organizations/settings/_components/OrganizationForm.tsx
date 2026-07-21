// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
  useDialogRef,
} from "@probo/ui";
import { type ChangeEventHandler, useState } from "react";
import { useTranslation } from "react-i18next";
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
    logo {
      downloadUrl
    }
    horizontalLogo {
      downloadUrl
    }
    canUpdate: permission(action: "iam:organization:update")
  }
`;

const updateOrganizationMutation = graphql`
  mutation OrganizationForm_updateMutation($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      organization {
        id
        name
        logo {
          downloadUrl
        }
        horizontalLogo {
          downloadUrl
        }
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
        horizontalLogo {
          downloadUrl
        }
      }
    }
  }
`;

const organizationSchema = z.object({
  name: z.string().min(1, "Organization name is required"),
});

type OrganizationFormData = z.infer<typeof organizationSchema>;

export function OrganizationForm(props: {
  fKey: OrganizationFormFragment$key;
}) {
  const { fKey } = props;
  const { t } = useTranslation();
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
      successMessage: t("organizationForm.messages.updated"),
      errorMessage: t("organizationForm.errors.update"),
    },
  );
  const [deleteHorizontalLogo, isDeletingHorizontalLogo]
    = useMutationWithToasts(deleteHorizontalLogoMutation, {
      successMessage: t("organizationForm.messages.horizontalLogoDeleted"),
      errorMessage: t("organizationForm.errors.deleteHorizontalLogo"),
    });

  const { formState, handleSubmit, register } = useFormWithSchema(
    organizationSchema,
    {
      defaultValues: {
        name: organization.name,
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
        },
      },
    });
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{t("organizationForm.title")}</h2>
        {formState.isSubmitting && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <div>
          <Label>{t("organizationForm.fields.logo")}</Label>
          <div className="flex w-max items-center gap-4">
            <Avatar
              className={logoPreview || organization.logo?.downloadUrl ? "bg-transparent" : undefined}
              src={logoPreview || organization.logo?.downloadUrl}
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
                  ? t("organizationForm.actions.uploading")
                  : t("organizationForm.actions.changeLogo")}
              </FileButton>
            )}
          </div>
        </div>
        <div>
          <Label>{t("organizationForm.fields.horizontalLogo")}</Label>
          <p className="text-sm text-txt-tertiary mb-2">
            {t("organizationForm.fields.horizontalLogoDescription")}
          </p>
          <div className="flex items-center gap-4">
            {(horizontalLogoPreview || organization.horizontalLogo?.downloadUrl) && (
              <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                <img
                  src={
                    horizontalLogoPreview
                    || organization.horizontalLogo?.downloadUrl
                    || undefined
                  }
                  alt={t("organizationForm.fields.horizontalLogo")}
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
                  ? t("organizationForm.actions.uploading")
                  : horizontalLogoPreview || organization.horizontalLogo?.downloadUrl
                    ? t("organizationForm.actions.changeHorizontalLogo")
                    : t("organizationForm.actions.uploadHorizontalLogo")}
              </FileButton>
            )}
            {canUpdate && organization.horizontalLogo?.downloadUrl && (
              <Dialog
                ref={deleteDialogRef}
                trigger={(
                  <Button
                    type="button"
                    variant="quaternary"
                    icon={IconTrashCan}
                    aria-label={t("organizationForm.actions.deleteHorizontalLogo")}
                    className="text-red-600 hover:text-red-700"
                  />
                )}
                title={t("organizationForm.deleteHorizontalLogo.title")}
                className="max-w-md"
              >
                <DialogContent padded>
                  <p className="text-txt-secondary">
                    {t("organizationForm.deleteHorizontalLogo.description")}
                  </p>
                  <p className="text-txt-secondary mt-2">
                    {t("organizationForm.deleteHorizontalLogo.warning")}
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
                      ? t("organizationForm.actions.deleting")
                      : t("organizationForm.actions.delete")}
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
          label={t("organizationForm.fields.name")}
          placeholder={t("organizationForm.fields.name")}
        />

        {formState.isDirty && canUpdate && (
          <div className="flex justify-end pt-6">
            <Button
              type="submit"
              disabled={formState.isSubmitting || isUpdatingOrganization}
            >
              {formState.isSubmitting || isUpdatingOrganization
                ? t("organizationForm.actions.updating")
                : t("organizationForm.actions.update")}
            </Button>
          </div>
        )}
      </Card>
    </form>
  );
}
