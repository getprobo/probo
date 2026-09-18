// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { Form } from "@base-ui/react/form";
import { TrashIcon } from "@phosphor-icons/react";
import { toFieldErrors } from "@probo/helpers";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { WorkspaceIdentitySection_deleteHorizontalLogoMutation } from "#/__generated__/iam/WorkspaceIdentitySection_deleteHorizontalLogoMutation.graphql";
import type { WorkspaceIdentitySection_updateHorizontalLogoMutation } from "#/__generated__/iam/WorkspaceIdentitySection_updateHorizontalLogoMutation.graphql";
import type { WorkspaceIdentitySection_updateLogoMutation } from "#/__generated__/iam/WorkspaceIdentitySection_updateLogoMutation.graphql";
import type { WorkspaceIdentitySection_updateMutation } from "#/__generated__/iam/WorkspaceIdentitySection_updateMutation.graphql";
import type { WorkspaceIdentitySectionFragment$key } from "#/__generated__/iam/WorkspaceIdentitySectionFragment.graphql";
import { ImageDropzone } from "#/components/ImageDropzone/ImageDropzone";
import { useMutation } from "#/lib/relay/useMutation";
import type { FileDropzoneError } from "#/lib/useFileDropzone";

import { identitySection } from "../variants";

const SETTINGS_NS = "iam/organizations/settings";
const NAME_MAX_LENGTH = 255;

const fragment = graphql`
  fragment WorkspaceIdentitySectionFragment on Organization {
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
  mutation WorkspaceIdentitySection_updateMutation($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      organization {
        ...WorkspaceIdentitySectionFragment
      }
    }
  }
`;

const updateLogoMutation = graphql`
  mutation WorkspaceIdentitySection_updateLogoMutation($input: UpdateOrganizationInput!) {
    updateOrganization(input: $input) {
      organization {
        ...WorkspaceIdentitySectionFragment
      }
    }
  }
`;

const updateHorizontalLogoMutation = graphql`
  mutation WorkspaceIdentitySection_updateHorizontalLogoMutation(
    $input: UpdateOrganizationInput!
  ) {
    updateOrganization(input: $input) {
      organization {
        ...WorkspaceIdentitySectionFragment
      }
    }
  }
`;

const deleteHorizontalLogoMutation = graphql`
  mutation WorkspaceIdentitySection_deleteHorizontalLogoMutation(
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

interface WorkspaceIdentitySectionProps {
  organizationKey: WorkspaceIdentitySectionFragment$key;
}

export function WorkspaceIdentitySection({ organizationKey }: WorkspaceIdentitySectionProps) {
  const { t } = useTranslation(SETTINGS_NS);
  const { toast } = useToast();
  const { form, logos, logoCell, horizontalCell, actions, root, intro } = identitySection();
  const organization = useFragment(fragment, organizationKey);

  const [name, setName] = useState(organization.name);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [horizontalLogoPreview, setHorizontalLogoPreview] = useState<string | null>(null);
  const [clearOpen, setClearOpen] = useState(false);

  const [updateOrganization, isUpdating] = useMutation<WorkspaceIdentitySection_updateMutation>(
    updateOrganizationMutation,
    {
      successMessage: t("identity.messages.updated"),
      errorToast: t("identity.errors.update"),
    },
  );
  const [updateLogo, isUploadingLogo] = useMutation<WorkspaceIdentitySection_updateLogoMutation>(
    updateLogoMutation,
    {
      successMessage: t("identity.messages.updated"),
      errorToast: t("identity.errors.update"),
    },
  );
  const [updateHorizontalLogo, isUploadingHorizontalLogo]
    = useMutation<WorkspaceIdentitySection_updateHorizontalLogoMutation>(
      updateHorizontalLogoMutation,
      {
        successMessage: t("identity.messages.updated"),
        errorToast: t("identity.errors.update"),
      },
    );
  const [deleteHorizontalLogo, isDeletingHorizontalLogo]
    = useMutation<WorkspaceIdentitySection_deleteHorizontalLogoMutation>(
      deleteHorizontalLogoMutation,
      {
        successMessage: t("identity.messages.horizontalLogoDeleted"),
        errorToast: t("identity.errors.deleteHorizontalLogo"),
      },
    );

  const canUpdate = organization.canUpdate;
  const busy = isUpdating || isUploadingLogo || isUploadingHorizontalLogo
    || isDeletingHorizontalLogo;
  const dropzoneDisabled = !canUpdate || busy;
  const nameDirty = name !== organization.name;
  const logoSrc = logoPreview ?? organization.logo?.downloadUrl;
  const horizontalLogoSrc = horizontalLogoPreview ?? organization.horizontalLogo?.downloadUrl;

  function handleReject(error: FileDropzoneError) {
    toast({
      title: t(`identity.errors.${error}.title`),
      description: t(`identity.errors.${error}.description`),
      variant: "error",
    });
  }

  function handleLogoFile(file: File) {
    const preview = URL.createObjectURL(file);
    setLogoPreview(preview);
    void updateLogo({
      variables: {
        input: {
          organizationId: organization.id,
          logoFile: null,
        },
      },
      uploadables: {
        "input.logoFile": file,
      },
    }).then(
      () => {
        URL.revokeObjectURL(preview);
        setLogoPreview(null);
      },
      () => {
        URL.revokeObjectURL(preview);
        setLogoPreview(null);
      },
    );
  }

  function handleHorizontalLogoFile(file: File) {
    const preview = URL.createObjectURL(file);
    setHorizontalLogoPreview(preview);
    void updateHorizontalLogo({
      variables: {
        input: {
          organizationId: organization.id,
          horizontalLogoFile: null,
        },
      },
      uploadables: {
        "input.horizontalLogoFile": file,
      },
    }).then(
      () => {
        URL.revokeObjectURL(preview);
        setHorizontalLogoPreview(null);
      },
      () => {
        URL.revokeObjectURL(preview);
        setHorizontalLogoPreview(null);
      },
    );
  }

  function handleClearHorizontalLogo() {
    void deleteHorizontalLogo({
      variables: {
        input: {
          organizationId: organization.id,
        },
      },
    }).then(
      () => {
        setClearOpen(false);
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  function handleSubmit() {
    const nextName = name.trim();
    if (nextName.length === 0) {
      setErrors({ name: t("identity.errors.nameRequired") });
      return;
    }

    void updateOrganization({
      variables: {
        input: {
          organizationId: organization.id,
          name: nextName,
        },
      },
      onCompleted(_response, payloadErrors) {
        const fieldErrors = toFieldErrors(payloadErrors);
        if (fieldErrors != null) {
          setErrors(fieldErrors);
        }
      },
    }).then(
      () => {
        setName(nextName);
        setErrors({});
      },
      () => {
        // Field errors are mapped in onCompleted; other failures toast.
      },
    );
  }

  return (
    <>
      <section className={root()}>
        <div className={intro()}>
          <Heading level={2} size={4} weight="medium" highContrast>
            {t("identity.title")}
          </Heading>
          <Text size={2} color="neutral">
            {t("identity.description")}
          </Text>
        </div>
        <Card size={2} variant="soft">
          <Form className={form()} errors={errors} onFormSubmit={handleSubmit}>
            <div className={logos()}>
              <div className={logoCell()}>
                <Text size={2} weight="medium" highContrast>
                  {t("identity.fields.logo")}
                </Text>
                <ImageDropzone
                  ratio="square"
                  src={logoSrc}
                  disabled={dropzoneDisabled}
                  uploading={isUploadingLogo}
                  placeholder={t("identity.fields.logoPlaceholder")}
                  onFile={handleLogoFile}
                  onReject={handleReject}
                />
              </div>
              <div className={horizontalCell()}>
                <Text size={2} weight="medium" highContrast>
                  {t("identity.fields.horizontalLogo")}
                </Text>
                <ImageDropzone
                  ratio="wide"
                  src={horizontalLogoSrc}
                  disabled={dropzoneDisabled}
                  uploading={isUploadingHorizontalLogo}
                  placeholder={t("identity.fields.horizontalLogoPlaceholder")}
                  clearLabel={t("identity.actions.clearHorizontalLogo")}
                  onFile={handleHorizontalLogoFile}
                  onClear={canUpdate ? () => setClearOpen(true) : undefined}
                  onReject={handleReject}
                />
              </div>
            </div>
            <Callout color="sky">
              {t("identity.fields.acceptedFiles")}
            </Callout>
            <Field label={t("identity.fields.name")} error={errors.name}>
              <TextField
                name="name"
                required
                maxLength={NAME_MAX_LENGTH}
                value={name}
                disabled={!canUpdate || busy}
                onValueChange={(value) => {
                  setName(value);
                  setErrors({});
                }}
              />
            </Field>
            {nameDirty && canUpdate && (
              <div className={actions()}>
                <Button
                  type="submit"
                  variant="solid"
                  color="neutral"
                  highContrast
                  loading={isUpdating}
                  disabled={busy}
                >
                  {t("identity.actions.save")}
                </Button>
              </div>
            )}
          </Form>
        </Card>
      </section>
      <Dialog open={clearOpen} onOpenChange={setClearOpen}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("identity.deleteHorizontalLogo.title")}</DialogTitle>
            <DialogDescription>
              {t("identity.deleteHorizontalLogo.description")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral">
                  {t("identity.deleteHorizontalLogo.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="button"
              variant="solid"
              color="red"
              iconStart={<TrashIcon />}
              loading={isDeletingHorizontalLogo}
              onClick={handleClearHorizontalLogo}
            >
              {t("identity.deleteHorizontalLogo.actions.confirm")}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </>
  );
}
