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

import { safeOpenUrl } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ChangeEventHandler, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalNDACard_compliancePortal$key } from "#/__generated__/core/CompliancePortalNDACard_compliancePortal.graphql";

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
  onFileChange: ChangeEventHandler<HTMLInputElement>;
}

export function CompliancePortalNDACard({
  compliancePortalKey,
  isUploading,
  onFileChange,
}: CompliancePortalNDACardProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, copy, actions } = ndaCard();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const compliancePortal = useFragment(fragment, compliancePortalKey);
  const fileName = compliancePortal.nda?.fileName;

  if (!fileName) {
    return null;
  }

  return (
    <Card size={2} variant="soft">
      <div className={root()}>
        <div className={copy()}>
          <Text size={3} weight="medium" highContrast>
            {fileName}
          </Text>
          <Text size={2} color="neutral">
            {t("ndaSection.acceptanceDescription")}
          </Text>
        </div>

        <div className={actions()}>
          <Button
            type="button"
            variant="soft"
            color="neutral"
            onClick={() => {
              if (compliancePortal.nda?.downloadUrl) {
                safeOpenUrl(compliancePortal.nda.downloadUrl);
              }
            }}
          >
            {t("ndaSection.actions.download")}
          </Button>

          {compliancePortal.canUploadNDA && (
            <>
              <Button
                type="button"
                variant="soft"
                color="neutral"
                loading={isUploading}
                onClick={() => fileInputRef.current?.click()}
              >
                {isUploading
                  ? t("ndaSection.actions.uploading")
                  : t("ndaSection.actions.replace")}
              </Button>
              <input
                ref={fileInputRef}
                type="file"
                hidden
                accept="application/pdf,.pdf"
                onChange={onFileChange}
              />
            </>
          )}

          {compliancePortal.canDeleteNDA && (
            <CompliancePortalNDADeleteDialog compliancePortalId={compliancePortal.id}>
              <Button
                type="button"
                variant="solid"
                color="red"
                disabled={isUploading}
              >
                {t("ndaSection.delete.actions.delete")}
              </Button>
            </CompliancePortalNDADeleteDialog>
          )}
        </div>
      </div>
    </Card>
  );
}
