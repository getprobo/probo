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
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  IconArrowInbox,
  IconWarning,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Suspense, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useLazyLoadQuery } from "react-relay";

import type { EvidenceGraphFileQuery } from "#/__generated__/core/EvidenceGraphFileQuery.graphql";
import { CSVPreview } from "#/components/documents/CSVPreview";
import { evidencePreviewFormat } from "#/components/documents/evidencePreviewFormat";
import { PDFPreview } from "#/components/documents/PDFPreview";
import { TextPreview } from "#/components/documents/TextPreview";
import { evidenceFileQuery } from "#/hooks/graph/EvidenceGraph";

type Props = {
  evidenceId: string;
  filename: string;
  onClose: () => void;
};

export function EvidencePreviewDialog({
  evidenceId,
  filename,
  onClose,
}: Props) {
  const { t } = useTranslation();
  const ref = useDialogRef();
  return (
    <Dialog
      ref={ref}
      defaultOpen
      title={
        <Breadcrumb items={[{ label: t("evidencePreviewDialog.breadcrumb.evidences") }, { label: filename }]} />
      }
      onClose={onClose}
    >
      <DialogContent padded>
        <Suspense fallback={<Spinner />}>
          <EvidencePreviewContent
            evidenceId={evidenceId}
            onClose={() => ref.current?.close()}
          />
        </Suspense>
      </DialogContent>
    </Dialog>
  );
}

const fetchUrlFromUriFile = async (
  fileUrl: string,
  options?: { signal?: AbortSignal },
): Promise<string> => {
  const response = await fetch(fileUrl, options);
  const text = await response.text();
  // URI files typically have the URL on the first line
  const firstLine = text.trim().split("\n")[0];
  if (!firstLine) {
    throw new Error("No URL found in URI file");
  }
  return firstLine;
};

function EvidencePreviewContent({
  evidenceId,
  onClose,
}: Omit<Props, "filename">) {
  const evidence = useLazyLoadQuery<EvidenceGraphFileQuery>(
    evidenceFileQuery,
    { evidenceId: evidenceId },
    { fetchPolicy: "network-only" },
  ).node;
  const { t } = useTranslation();
  const toast = useToast();
  const file = evidence.file;
  const previewFormat = evidencePreviewFormat(file?.mimeType, file?.fileName);
  useEffect(() => {
    if (previewFormat !== "uri") {
      return;
    }
    const abortController = new AbortController();
    fetchUrlFromUriFile(file?.downloadUrl ?? "", {
      signal: abortController.signal,
    })
      .then((url) => {
        window.open(url, "_blank");
      })
      .catch((e) => {
        if (e instanceof Error) {
          if (e.name === "AbortError") {
            return;
          }
          toast.add({
            title: t("evidencePreviewDialog.messages.error"),
            description: t("evidencePreviewDialog.errors.extractUrl"),
            type: "error",
          });
        } else {
          toast.add({
            title: t("evidencePreviewDialog.messages.error"),
            description: t("evidencePreviewDialog.errors.extractUrl"),
            type: "error",
          });
        }
      })
      .finally(onClose);
    return () => {
      abortController.abort();
    };
  }, [file?.downloadUrl, onClose, previewFormat, t, toast]);

  if (file == null || !file.downloadUrl) {
    return null;
  }

  if (previewFormat === "uri") {
    return (
      <div className="flex flex-col items-center gap-2 justify-center">
        <Spinner size={20} />
      </div>
    );
  }

  let preview;

  switch (previewFormat) {
    case "image":
      preview = (
        <img
          src={file.downloadUrl}
          alt={file.fileName}
          className="max-h-[70vh] object-contain"
        />
      );
      break;
    case "pdf":
      preview = (
        <div className="h-[70vh] overflow-hidden">
          <PDFPreview
            src={file.downloadUrl}
            name={file.fileName}
          />
        </div>
      );
      break;
    case "csv":
      preview = (
        <CSVPreview
          src={file.downloadUrl}
          columnLabel={t("evidencePreviewDialog.csv.column")}
          emptyMessage={t("evidencePreviewDialog.csv.empty")}
          errorMessage={t("evidencePreviewDialog.csv.error")}
          retryLabel={t("evidencePreviewDialog.csv.retry")}
          truncatedMessage={t("evidencePreviewDialog.csv.truncated")}
        />
      );
      break;
    case "json":
      preview = (
        <TextPreview
          src={file.downloadUrl}
          emptyMessage={t("evidencePreviewDialog.json.empty")}
          errorMessage={t("evidencePreviewDialog.json.error")}
          format="json"
          retryLabel={t("evidencePreviewDialog.json.retry")}
          truncatedMessage={t("evidencePreviewDialog.json.truncated")}
        />
      );
      break;
    case "markdown":
    case "text":
      preview = (
        <TextPreview
          src={file.downloadUrl}
          emptyMessage={t("evidencePreviewDialog.text.empty")}
          errorMessage={t("evidencePreviewDialog.text.error")}
          format={previewFormat}
          retryLabel={t("evidencePreviewDialog.text.retry")}
          truncatedMessage={t("evidencePreviewDialog.text.truncated")}
        />
      );
      break;
    case "unavailable":
      preview = (
        <div className="flex flex-col items-center gap-2 justify-center">
          <IconWarning size={20} />
          <p className="text-txt-secondary text-center">
            {t("evidencePreviewDialog.previewUnavailable", {
              mimeType: file.mimeType,
            })}
          </p>
          <Button asChild variant="secondary" icon={IconArrowInbox}>
            <a href={file.downloadUrl} target="_blank" rel="noreferrer">
              {t("evidencePreviewDialog.actions.download")}
            </a>
          </Button>
        </div>
      );
      break;
  }

  return (
    <div className="space-y-4">
      {preview}
      {evidence.description && (
        <p className="text-txt-secondary text-sm">{evidence.description}</p>
      )}
    </div>
  );
}
