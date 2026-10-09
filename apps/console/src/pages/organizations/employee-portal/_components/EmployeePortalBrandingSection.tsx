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

import { useToast } from "@probo/ui";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EmployeePortalBrandingSection_employeePortal$key } from "#/__generated__/core/EmployeePortalBrandingSection_employeePortal.graphql";
import type { EmployeePortalBrandingSection_updateMutation } from "#/__generated__/core/EmployeePortalBrandingSection_updateMutation.graphql";
import { ImageDropzone } from "#/components/ImageDropzone/ImageDropzone";
import { useMutation } from "#/lib/relay/useMutation";
import type { FileDropzoneError } from "#/lib/useFileDropzone";

import { employeePortalBrandingSection } from "../variants";

const NS = "organizations/employee-portal";

const employeePortalFragment = graphql`
  fragment EmployeePortalBrandingSection_employeePortal on EmployeePortal {
    id
    logo {
      downloadUrl
    }
    darkLogo {
      downloadUrl
    }
    canUploadBrand: permission(action: "employee-portal:portal:upload-brand")
    canDeleteBrand: permission(action: "employee-portal:portal:delete-brand")
  }
`;

const updateBrandMutation = graphql`
  mutation EmployeePortalBrandingSection_updateMutation(
    $input: UpdateEmployeePortalBrandInput!
  ) {
    updateEmployeePortalBrand(input: $input) {
      employeePortal {
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

function useRevokeObjectUrl(url: string | null) {
  useEffect(() => {
    if (url == null) {
      return;
    }

    return () => {
      URL.revokeObjectURL(url);
    };
  }, [url]);
}

export interface EmployeePortalBrandingSectionProps {
  employeePortalKey: EmployeePortalBrandingSection_employeePortal$key;
}

export function EmployeePortalBrandingSection({
  employeePortalKey,
}: EmployeePortalBrandingSectionProps) {
  const { t } = useTranslation(NS);
  const { toast } = useToast();
  const { root, intro, body, logos, logoCell, darkIsland } = employeePortalBrandingSection();
  const employeePortal = useFragment(employeePortalFragment, employeePortalKey);

  const [logoPreview, setLogoPreview] = useState<string | null>(null);
  const [darkLogoPreview, setDarkLogoPreview] = useState<string | null>(null);
  const [uploadingField, setUploadingField] = useState<"logo" | "darkLogo" | null>(null);
  useRevokeObjectUrl(logoPreview);
  useRevokeObjectUrl(darkLogoPreview);

  const [updateBrand] = useMutation<EmployeePortalBrandingSection_updateMutation>(
    updateBrandMutation,
    {
      successMessage: t("branding.messages.updated"),
      errorToast: t("branding.errors.update"),
    },
  );

  const logoSrc = logoPreview ?? employeePortal.logo?.downloadUrl;
  const darkLogoSrc = darkLogoPreview ?? employeePortal.darkLogo?.downloadUrl;
  const busy = uploadingField != null;

  function handleReject(error: FileDropzoneError) {
    toast({
      title: t(`branding.errors.${error}.title`),
      description: t(`branding.errors.${error}.description`),
      variant: "error",
    });
  }

  function uploadLogo(field: "logoFile" | "darkLogoFile", file: File) {
    const preview = URL.createObjectURL(file);
    const isDark = field === "darkLogoFile";
    const setPreview = isDark ? setDarkLogoPreview : setLogoPreview;
    setPreview(preview);
    setUploadingField(isDark ? "darkLogo" : "logo");
    void updateBrand({
      variables: {
        input: {
          employeePortalId: employeePortal.id,
          [field]: null,
        },
      },
      uploadables: {
        [`input.${field}`]: file,
      },
    }).then(
      () => {
        setPreview(null);
        setUploadingField(null);
      },
      () => {
        setPreview(null);
        setUploadingField(null);
      },
    );
  }

  function clearLogo(field: "logoFile" | "darkLogoFile") {
    const isDark = field === "darkLogoFile";
    const setPreview = isDark ? setDarkLogoPreview : setLogoPreview;
    setUploadingField(isDark ? "darkLogo" : "logo");
    void updateBrand({
      variables: {
        input: {
          employeePortalId: employeePortal.id,
          [field]: null,
        },
      },
    }).then(
      () => {
        setPreview(null);
        setUploadingField(null);
      },
      () => {
        setUploadingField(null);
      },
    );
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("branding.title")}
        </Heading>
        <Text size={2} color="neutral">
          {t("branding.description")}
        </Text>
      </div>
      <Card size={2} variant="soft">
        <div className={body()}>
          <div className={logos()}>
            <div className={logoCell()}>
              <Text size={2} weight="medium" highContrast>
                {t("branding.fields.logo")}
              </Text>
              <ImageDropzone
                ratio="wide"
                src={logoSrc}
                disabled={!employeePortal.canUploadBrand || busy}
                uploading={uploadingField === "logo"}
                placeholder={t("branding.fields.placeholder")}
                clearLabel={t("branding.actions.removeLogo")}
                onFile={file => uploadLogo("logoFile", file)}
                onClear={employeePortal.canDeleteBrand && !busy
                  ? () => clearLogo("logoFile")
                  : undefined}
                onReject={handleReject}
              />
            </div>
            <div className={logoCell()}>
              <Text size={2} weight="medium" highContrast>
                {t("branding.fields.darkLogo")}
              </Text>
              <div className={darkIsland()}>
                <ImageDropzone
                  ratio="wide"
                  src={darkLogoSrc}
                  disabled={!employeePortal.canUploadBrand || busy}
                  uploading={uploadingField === "darkLogo"}
                  placeholder={t("branding.fields.placeholder")}
                  clearLabel={t("branding.actions.removeDarkLogo")}
                  onFile={file => uploadLogo("darkLogoFile", file)}
                  onClear={employeePortal.canDeleteBrand && !busy
                    ? () => clearLogo("darkLogoFile")
                    : undefined}
                  onReject={handleReject}
                />
              </div>
            </div>
          </div>
          <Callout color="sky">
            {t("branding.fields.acceptedFiles")}
          </Callout>
        </div>
      </Card>
    </section>
  );
}
