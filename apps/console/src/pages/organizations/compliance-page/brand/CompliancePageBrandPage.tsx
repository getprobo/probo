import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Dropzone,
  FileButton,
  IconTrashCan,
  Label,
  Spinner,
  useToast,
} from "@probo/ui";
import { type ChangeEventHandler, useState } from "react";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageBrandPage_updateMutation } from "#/__generated__/core/CompliancePageBrandPage_updateMutation.graphql";
import type { CompliancePageBrandPageQuery } from "#/__generated__/core/CompliancePageBrandPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

export const compliancePageBrandPageQuery = graphql`
  query CompliancePageBrandPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        compliancePage: trustCenter @required(action: THROW) {
          id
          logoFileUrl
          darkLogoFileUrl
          canUpdate: permission(action: "core:trust-center:update")
        }
      }
    }
  }
`;

const updateTrustCenterBrandMutation = graphql`
  mutation CompliancePageBrandPage_updateMutation($input: UpdateTrustCenterBrandInput!) {
    updateTrustCenterBrand(input: $input) {
      trustCenter {
        id
        logoFileUrl
        darkLogoFileUrl
      }
    }
  }
`;

export function CompliancePageBrandPage(props: { queryRef: PreloadedQuery<CompliancePageBrandPageQuery> }) {
  const { queryRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const { organization } = usePreloadedQuery<CompliancePageBrandPageQuery>(compliancePageBrandPageQuery, queryRef);
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const trustCenterId = organization.compliancePage.id;
  const logoFileUrl = organization.compliancePage.logoFileUrl;
  const darkLogoFileUrl = organization.compliancePage.darkLogoFileUrl;

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [darkLogoPreview, setDarkLogoPreview] = useState<string | null>(null);

  const [updateBrand, isUpdating] = useMutationWithToasts<CompliancePageBrandPage_updateMutation>(
    updateTrustCenterBrandMutation,
    {
      successMessage: __("Compliance page branding updated successfully"),
      errorMessage: __("Failed to update compliance page branding"),
    },
  );
  const disabled = isUpdating || !organization.compliancePage.canUpdate;

  const processLogoFile = (file: File, setPreview: (url: string) => void) => {
    const reader = new FileReader();
    reader.onload = () => {
      setPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleLogoChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      toast({
        title: __("File size too large"),
        description: __("The file size is too large. Please upload a file smaller than 5MB."),
        variant: "error",
      });
      return;
    }

    processLogoFile(file, setLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId,
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

  const handleDarkLogoChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      toast({
        title: __("File size too large"),
        description: __("The file size is too large. Please upload a file smaller than 5MB."),
        variant: "error",
      });
      return;
    }

    processLogoFile(file, setDarkLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId,
          darkLogoFile: null,
        },
      },
      uploadables: {
        "input.darkLogoFile": file,
      },
      onCompleted: () => {
        setDarkLogoPreview(null);
      },
    });
  };

  const handleLogoDrop = (files: File[]) => {
    const file = files[0];
    if (!file) return;

    processLogoFile(file, setLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId,
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

  const handleDarkLogoDrop = (files: File[]) => {
    const file = files[0];
    if (!file) return;

    processLogoFile(file, setDarkLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId,
          darkLogoFile: null,
        },
      },
      uploadables: {
        "input.darkLogoFile": file,
      },
      onCompleted: () => {
        setDarkLogoPreview(null);
      },
    });
  };

  const handleRemoveLogo = async () => {
    await updateBrand({
      variables: {
        input: {
          trustCenterId,
          logoFile: null,
        },
      },
      onSuccess: () => {
        setLogoPreview(null);
      },
    });
  };

  const handleRemoveDarkLogo = async () => {
    await updateBrand({
      variables: {
        input: {
          trustCenterId,
          darkLogoFile: null,
        },
      },
      onSuccess: () => {
        setDarkLogoPreview(null);
      },
    });
  };

  const currentLogoUrl = logoPreview || logoFileUrl;
  const currentDarkLogoUrl = darkLogoPreview || darkLogoFileUrl;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Branding")}</h2>
        {isUpdating && <Spinner />}
      </div>

      <Card padded className="space-y-4">
        <div className="flex gap-6 items-start">
          <div className="flex-1">
            <Label>{__("Logo")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {__("This logo will be displayed on your public compliance page.")}
            </p>

            {currentLogoUrl
              ? (
                  <div className="flex items-center gap-4">
                    <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                      <img
                        src={currentLogoUrl}
                        alt={__("Compliance page logo")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                    <FileButton
                      disabled={disabled}
                      onChange={handleLogoChange}
                      variant="secondary"
                      accept="image/png,image/jpeg,image/jpg,image/svg+xml,image/webp"
                    >
                      {isUpdating ? __("Uploading...") : __("Change logo")}
                    </FileButton>
                    <Button
                      type="button"
                      variant="quaternary"
                      icon={IconTrashCan}
                      onClick={() => void handleRemoveLogo()}
                      disabled={disabled}
                      aria-label={__("Remove logo")}
                      className="text-red-600 hover:text-red-700"
                    />
                  </div>
                )
              : (
                  <Dropzone
                    description={__("Upload logo image (PNG, JPG, SVG, WEBP up to 5MB)")}
                    isUploading={isUpdating}
                    onDrop={handleLogoDrop}
                    accept={{
                      "image/png": [".png"],
                      "image/jpeg": [".jpg", ".jpeg"],
                      "image/svg+xml": [".svg"],
                      "image/webp": [".webp"],
                    }}
                    maxSize={5}
                  />
                )}
          </div>
          <div className="flex-1">
            <Label>{__("Dark mode logo")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {__("This logo will be used when dark mode is enabled.")}
            </p>

            {currentDarkLogoUrl
              ? (
                  <div className="flex items-center gap-4">
                    <div className="border border-border-solid rounded-md p-4 bg-gray-900">
                      <img
                        src={currentDarkLogoUrl}
                        alt={__("Compliance page dark logo")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                    <FileButton
                      disabled={disabled}
                      onChange={handleDarkLogoChange}
                      variant="secondary"
                      accept="image/png,image/jpeg,image/jpg,image/svg+xml,image/webp"
                    >
                      {isUpdating ? __("Uploading...") : __("Change dark logo")}
                    </FileButton>
                    <Button
                      type="button"
                      variant="quaternary"
                      icon={IconTrashCan}
                      onClick={() => void handleRemoveDarkLogo()}
                      disabled={disabled}
                      aria-label={__("Remove dark logo")}
                      className="text-red-600 hover:text-red-700"
                    />
                  </div>
                )
              : (
                  <Dropzone
                    description={__("Upload dark logo image (PNG, JPG, SVG, WEBP up to 5MB)")}
                    isUploading={isUpdating}
                    onDrop={handleDarkLogoDrop}
                    accept={{
                      "image/png": [".png"],
                      "image/jpeg": [".jpg", ".jpeg"],
                      "image/svg+xml": [".svg"],
                      "image/webp": [".webp"],
                    }}
                    maxSize={5}
                  />
                )}
          </div>
        </div>
      </Card>
    </div>
  );
}
