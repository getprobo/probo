// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  FileButton,
  IconTrashCan,
  Label,
  Spinner,
  useToast,
} from "@probo/ui";
import { type ChangeEventHandler, useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageVisualIdentitySection_compliancePageFragment$key } from "#/__generated__/core/CompliancePageVisualIdentitySection_compliancePageFragment.graphql";
import type { CompliancePageVisualIdentitySection_updateMutation } from "#/__generated__/core/CompliancePageVisualIdentitySection_updateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const compliancePageFragment = graphql`
  fragment CompliancePageVisualIdentitySection_compliancePageFragment on TrustCenter {
    id
    logo {
      downloadUrl
    }
    darkLogo {
      downloadUrl
    }
    canUpdate: permission(action: "compliance-portal:portal:update")
  }
`;

const updateMutation = graphql`
  mutation CompliancePageVisualIdentitySection_updateMutation($input: UpdateTrustCenterBrandInput!) {
    updateTrustCenterBrand(input: $input) {
      trustCenter {
        id
        logo {
          downloadUrl
        }
        darkLogo {
          downloadUrl
        }
      }
    }
  }
`;

const acceptImageMimeTypes = "image/png,image/jpeg,image/jpg,image/svg+xml,image/webp";
const maxLogoBytes = 5 * 1024 * 1024;

export interface CompliancePageVisualIdentitySectionProps {
  compliancePageRef: CompliancePageVisualIdentitySection_compliancePageFragment$key;
}

export function CompliancePageVisualIdentitySection(props: CompliancePageVisualIdentitySectionProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();

  const compliancePage = useFragment(compliancePageFragment, props.compliancePageRef);
  const compliancePageId = compliancePage.id;

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [darkLogoPreview, setDarkLogoPreview] = useState<string | null>(null);

  const [updateBrand, isUpdating] = useMutation<CompliancePageVisualIdentitySection_updateMutation>(
    updateMutation,
    {
      successMessage: __("Compliance page branding updated successfully"),
      errorToast: __("Failed to update compliance page branding"),
    },
  );

  const disabled = isUpdating || !compliancePage.canUpdate;

  const processLogoFile = (file: File, setPreview: (url: string) => void) => {
    const reader = new FileReader();
    reader.onload = () => {
      setPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
  };

  const isTooLarge = (file: File) => {
    if (file.size > maxLogoBytes) {
      toast({
        title: __("File size too large"),
        description: __("The file size is too large. Please upload a file smaller than 5MB."),
        variant: "error",
      });
      return true;
    }
    return false;
  };

  const handleLogoChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    if (!file || isTooLarge(file)) return;

    processLogoFile(file, setLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId: compliancePageId,
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
    if (!file || isTooLarge(file)) return;

    processLogoFile(file, setDarkLogoPreview);

    void updateBrand({
      variables: {
        input: {
          trustCenterId: compliancePageId,
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
          trustCenterId: compliancePageId,
          logoFile: null,
        },
      },
    });

    setLogoPreview(null);
  };

  const handleRemoveDarkLogo = async () => {
    await updateBrand({
      variables: {
        input: {
          trustCenterId: compliancePageId,
          darkLogoFile: null,
        },
      },
    });

    setDarkLogoPreview(null);
  };

  const currentLogoUrl = logoPreview || compliancePage.logo?.downloadUrl;
  const currentDarkLogoUrl = darkLogoPreview || compliancePage.darkLogo?.downloadUrl;

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">{__("Visual identity")}</h2>
          <p className="text-sm text-txt-tertiary">
            {__("Square logos displayed on your public compliance page.")}
          </p>
        </div>
        {isUpdating && <Spinner />}
      </div>

      <Card padded className="space-y-4">
        <div className="flex gap-6 items-start">
          <div className="flex-1">
            <Label>{__("Logo")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {__("Upload a square logo for your public compliance page.")}
            </p>

            <div className="flex items-center gap-4">
              {currentLogoUrl
                ? (
                    <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                      <img
                        src={currentLogoUrl}
                        alt={__("Compliance page logo")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                  )
                : (
                    <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-dashed border-border-solid bg-surface-secondary text-xs text-txt-tertiary">
                      {__("No logo")}
                    </div>
                  )}
              <div className="space-y-1">
                <FileButton
                  disabled={disabled}
                  onChange={handleLogoChange}
                  variant="secondary"
                  accept={acceptImageMimeTypes}
                >
                  {isUpdating
                    ? __("Uploading...")
                    : currentLogoUrl
                      ? __("Change logo")
                      : __("Upload logo")}
                </FileButton>
                {!currentLogoUrl && (
                  <p className="text-xs text-txt-tertiary">
                    {__("Square format. PNG, JPG, SVG, or WEBP up to 5MB")}
                  </p>
                )}
              </div>
              {currentLogoUrl && (
                <Button
                  type="button"
                  variant="quaternary"
                  icon={IconTrashCan}
                  onClick={() => void handleRemoveLogo()}
                  disabled={disabled}
                  aria-label={__("Remove logo")}
                  className="text-red-600 hover:text-red-700"
                />
              )}
            </div>
          </div>
          <div className="flex-1">
            <Label>{__("Dark mode logo")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {__("Upload a square logo for use when dark mode is enabled.")}
            </p>

            <div className="flex items-center gap-4">
              {currentDarkLogoUrl
                ? (
                    <div className="border border-border-solid rounded-md p-4 bg-gray-900">
                      <img
                        src={currentDarkLogoUrl}
                        alt={__("Compliance page dark logo")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                  )
                : (
                    <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-dashed border-border-solid bg-gray-900 text-xs text-txt-tertiary">
                      {__("No logo")}
                    </div>
                  )}
              <div className="space-y-1">
                <FileButton
                  disabled={disabled}
                  onChange={handleDarkLogoChange}
                  variant="secondary"
                  accept={acceptImageMimeTypes}
                >
                  {isUpdating
                    ? __("Uploading...")
                    : currentDarkLogoUrl
                      ? __("Change dark logo")
                      : __("Upload dark logo")}
                </FileButton>
                {!currentDarkLogoUrl && (
                  <p className="text-xs text-txt-tertiary">
                    {__("Square format. PNG, JPG, SVG, or WEBP up to 5MB")}
                  </p>
                )}
              </div>
              {currentDarkLogoUrl && (
                <Button
                  type="button"
                  variant="quaternary"
                  icon={IconTrashCan}
                  onClick={() => void handleRemoveDarkLogo()}
                  disabled={disabled}
                  aria-label={__("Remove dark logo")}
                  className="text-red-600 hover:text-red-700"
                />
              )}
            </div>
          </div>
        </div>
      </Card>
    </section>
  );
}
