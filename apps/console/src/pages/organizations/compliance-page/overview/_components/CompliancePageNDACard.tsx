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

import { safeOpenUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card } from "@probo/ui";
import { type ChangeEventHandler, useRef } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNDACard_compliancePage$key } from "#/__generated__/core/CompliancePageNDACard_compliancePage.graphql";

const fragment = graphql`
  fragment CompliancePageNDACard_compliancePage on CompliancePortal {
    nda {
      fileName
      downloadUrl
    }
    canUploadNDA: permission(action: "compliance-portal:portal:upload-nda")
    canDeleteNDA: permission(action: "compliance-portal:portal:delete-nda")
  }
`;

export interface CompliancePageNDACardProps {
  compliancePageKey: CompliancePageNDACard_compliancePage$key;
  isBusy: boolean;
  isUploading: boolean;
  onFileChange: ChangeEventHandler<HTMLInputElement>;
  onDelete: () => void;
}

export function CompliancePageNDACard(props: CompliancePageNDACardProps) {
  const { compliancePageKey, isBusy, isUploading, onFileChange, onDelete } = props;

  const { __ } = useTranslate();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const compliancePage = useFragment(fragment, compliancePageKey);
  const fileName = compliancePage.nda?.fileName;

  if (!fileName) {
    return null;
  }

  return (
    <Card padded>
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <span className="font-medium">{fileName}</span>
          <p className="text-sm text-txt-secondary">
            {__(
              "Visitors must accept this agreement before accessing your compliance page.",
            )}
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <Button
            type="button"
            variant="secondary"
            onClick={() => {
              if (compliancePage.nda?.downloadUrl) {
                safeOpenUrl(compliancePage.nda.downloadUrl);
              }
            }}
          >
            {__("Download")}
          </Button>

          {compliancePage.canUploadNDA && (
            <>
              <Button
                type="button"
                variant="secondary"
                disabled={isBusy}
                onClick={() => fileInputRef.current?.click()}
              >
                {isUploading ? __("Uploading...") : __("Replace")}
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

          {compliancePage.canDeleteNDA && (
            <Button
              type="button"
              variant="danger"
              disabled={isBusy}
              onClick={onDelete}
            >
              {__("Delete")}
            </Button>
          )}
        </div>
      </div>
    </Card>
  );
}
