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

import { DownloadSimpleIcon, FileDashedIcon, FilePdfIcon, TrashIcon } from "@phosphor-icons/react";
import { safeOpenUrl } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalNDACard_compliancePortal$key } from "#/__generated__/core/CompliancePortalNDACard_compliancePortal.graphql";
import { CompliancePortalTonedCard } from "#/pages/organizations/compliance-portals/_components/CompliancePortalTonedCard";

import { type NdaUploadError, useNdaDropzone } from "../_lib/useNdaDropzone";
import { ndaCard } from "../variants";

import { CompliancePortalNDADeleteDialog } from "./CompliancePortalNDADeleteDialog";

const fragment = graphql`
  fragment CompliancePortalNDACard_compliancePortal on CompliancePortal {
    id
    nda {
      fileName
      downloadUrl
    }
    canUploadNDA: permission(action: "compliance-portal:portal:upload-nda")
    canDeleteNDA: permission(action: "compliance-portal:portal:delete-nda")
  }
`;

export interface CompliancePortalNDACardProps {
  compliancePortalKey: CompliancePortalNDACard_compliancePortal$key;
  isUploading: boolean;
  onFile: (file: File) => void;
  onReject: (error: NdaUploadError) => void;
}

export function CompliancePortalNDACard({
  compliancePortalKey,
  isUploading,
  onFile,
  onReject,
}: CompliancePortalNDACardProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const compliancePortal = useFragment(fragment, compliancePortalKey);
  const fileName = compliancePortal.nda?.fileName;
  const hasFile = fileName != null;
  const canUpload = compliancePortal.canUploadNDA;
  const { getRootProps, getInputProps, isDragActive, open } = useNdaDropzone({
    disabled: !canUpload || isUploading,
    onFile,
    onReject,
  });
  const { frame, controls, actions } = ndaCard({
    dashed: isDragActive || !hasFile,
  });

  const downloadUrl = compliancePortal.nda?.downloadUrl;
  const showDelete = hasFile && compliancePortal.canDeleteNDA;
  const tone = isDragActive ? "sky" : hasFile ? "green" : "sand";
  const dropHint = isDragActive
    ? t("ndaSection.dropActive")
    : hasFile
      ? t("ndaSection.dropReplace")
      : t("ndaSection.dropHint");

  return (
    <CompliancePortalTonedCard
      {...getRootProps({ className: frame() })}
      tone={tone}
      icon={hasFile
        ? <FilePdfIcon size={24} weight="duotone" />
        : <FileDashedIcon size={24} weight="duotone" />}
      lead={canUpload
        ? (
            <Text size={2} weight="medium" color={tone === "sand" ? "neutral" : tone} className="truncate">
              {dropHint}
            </Text>
          )
        : undefined}
      control={downloadUrl != null || showDelete
        ? (
            <div className={controls()}>
              {downloadUrl != null && (
                <IconButton
                  size={1}
                  variant="ghost"
                  color="neutral"
                  aria-label={t("ndaSection.actions.download")}
                  onClick={() => {
                    safeOpenUrl(downloadUrl);
                  }}
                >
                  <DownloadSimpleIcon />
                </IconButton>
              )}
              {showDelete && (
                <CompliancePortalNDADeleteDialog compliancePortalId={compliancePortal.id}>
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="red"
                    disabled={isUploading}
                    aria-label={t("ndaSection.actions.delete")}
                  >
                    <TrashIcon />
                  </IconButton>
                </CompliancePortalNDADeleteDialog>
              )}
            </div>
          )
        : undefined}
    >
      <input {...getInputProps()} />
      <Text size={3} weight="medium" highContrast className="w-full min-w-0 truncate">
        {hasFile ? fileName : t("ndaSection.empty")}
      </Text>
      <Text size={2} color="neutral">
        {t("ndaSection.description")}
      </Text>
      {canUpload && (
        <div className={actions()}>
          <Button
            type="button"
            variant="solid"
            color="neutral"
            highContrast
            loading={isUploading}
            onClick={() => open()}
          >
            {isUploading
              ? t("ndaSection.actions.uploading")
              : hasFile
                ? t("ndaSection.actions.replace")
                : t("ndaSection.actions.upload")}
          </Button>
        </div>
      )}
    </CompliancePortalTonedCard>
  );
}
