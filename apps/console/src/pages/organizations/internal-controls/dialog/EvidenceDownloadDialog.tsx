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

import { downloadFile } from "@probo/helpers";
import { Breadcrumb, Dialog, DialogContent, Spinner } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";

import type { EvidenceGraphFileQuery } from "#/__generated__/core/EvidenceGraphFileQuery.graphql";
import { evidenceFileQuery } from "#/hooks/graph/EvidenceGraph";

type Props = {
  evidenceId: string;
  onClose: () => void;
};
export function EvidenceDownloadDialog({ evidenceId, onClose }: Props) {
  const { t } = useTranslation();

  return (
    <Dialog
      className="max-w-sm"
      onClose={onClose}
      defaultOpen
      title={(
        <Breadcrumb
          items={[{ label: t("evidenceDownloadDialog.breadcrumb.evidences") }, { label: t("evidenceDownloadDialog.breadcrumb.download") }]}
        />
      )}
    >
      <DialogContent padded>
        <Suspense
          fallback={(
            <div className="flex gap-2 justify-center">
              <Spinner />
              {t("evidenceDownloadDialog.generating")}
            </div>
          )}
        >
          <DownloadLink evidenceId={evidenceId} onClose={onClose} />
        </Suspense>
      </DialogContent>
    </Dialog>
  );
}

/**
 * Force the download of an evidence file
 */
function DownloadLink({ evidenceId, onClose }: Props) {
  const data = useLazyLoadQuery<EvidenceGraphFileQuery>(evidenceFileQuery, {
    evidenceId,
  });
  const evidence = data.node;

  useEffect(() => {
    downloadFile(
      evidence.file?.downloadUrl,
      evidence.file?.fileName ?? "evidence",
    );
    onClose();
  }, [evidence.file?.downloadUrl, evidence.file?.fileName, onClose]);

  return null;
}
