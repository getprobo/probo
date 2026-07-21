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
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageVisualIdentitySection_compliancePageFragment$key } from "#/__generated__/core/CompliancePageVisualIdentitySection_compliancePageFragment.graphql";
import type { CompliancePageVisualIdentitySection_updateMutation } from "#/__generated__/core/CompliancePageVisualIdentitySection_updateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

const compliancePageFragment = graphql`
  fragment CompliancePageVisualIdentitySection_compliancePageFragment on CompliancePortal {
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
  mutation CompliancePageVisualIdentitySection_updateMutation($input: UpdateCompliancePortalBrandInput!) {
    updateCompliancePortalBrand(input: $input) {
      compliancePortal {
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
  const { t } = useTranslation("organizations/compliance-page");
  const { toast } = useToast();

  const compliancePage = useFragment(compliancePageFragment, props.compliancePageRef);
  const compliancePageId = compliancePage.id;

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [darkLogoPreview, setDarkLogoPreview] = useState<string | null>(null);

  const [updateBrand, isUpdating] = useMutation<CompliancePageVisualIdentitySection_updateMutation>(
    updateMutation,
    {
      successMessage: t("brandPage.messages.updated"),
      errorToast: t("brandPage.errors.update"),
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
        title: t("brandPage.errors.fileTooLarge.title"),
        description: t("brandPage.errors.fileTooLarge.description"),
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
          compliancePortalId: compliancePageId,
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
          compliancePortalId: compliancePageId,
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
          compliancePortalId: compliancePageId,
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
          compliancePortalId: compliancePageId,
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
          <h2 className="text-base font-medium">{t("brandPage.visualIdentity.title")}</h2>
          <p className="text-sm text-txt-tertiary">
            {t("brandPage.visualIdentity.description")}
          </p>
        </div>
        {isUpdating && <Spinner />}
      </div>

      <Card padded className="space-y-4">
        <div className="flex gap-6 items-start">
          <div className="flex-1">
            <Label>{t("brandPage.logo.label")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {t("brandPage.logo.description")}
            </p>

            <div className="flex items-center gap-4">
              {currentLogoUrl
                ? (
                    <div className="border border-border-solid rounded-md p-4 bg-surface-secondary">
                      <img
                        src={currentLogoUrl}
                        alt={t("brandPage.logo.alt")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                  )
                : (
                    <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-dashed border-border-solid bg-surface-secondary text-xs text-txt-tertiary">
                      {t("brandPage.actions.noLogo")}
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
                    ? t("brandPage.actions.uploading")
                    : currentLogoUrl
                      ? t("brandPage.actions.changeLogo")
                      : t("brandPage.actions.uploadLogo")}
                </FileButton>
                {!currentLogoUrl && (
                  <p className="text-xs text-txt-tertiary">
                    {t("brandPage.logo.uploadDescription")}
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
                  aria-label={t("brandPage.actions.removeLogo")}
                  className="text-red-600 hover:text-red-700"
                />
              )}
            </div>
          </div>
          <div className="flex-1">
            <Label>{t("brandPage.darkLogo.label")}</Label>
            <p className="text-sm text-txt-tertiary mb-3">
              {t("brandPage.darkLogo.description")}
            </p>

            <div className="flex items-center gap-4">
              {currentDarkLogoUrl
                ? (
                    <div className="border border-border-solid rounded-md p-4 bg-gray-900">
                      <img
                        src={currentDarkLogoUrl}
                        alt={t("brandPage.darkLogo.alt")}
                        className="h-16 max-w-xs object-contain"
                      />
                    </div>
                  )
                : (
                    <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-dashed border-border-solid bg-gray-900 text-xs text-txt-tertiary">
                      {t("brandPage.actions.noLogo")}
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
                    ? t("brandPage.actions.uploading")
                    : currentDarkLogoUrl
                      ? t("brandPage.actions.changeDarkLogo")
                      : t("brandPage.actions.uploadDarkLogo")}
                </FileButton>
                {!currentDarkLogoUrl && (
                  <p className="text-xs text-txt-tertiary">
                    {t("brandPage.darkLogo.uploadDescription")}
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
                  aria-label={t("brandPage.actions.removeDarkLogo")}
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
