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

import { Button, Dialog, DialogContent, IconArrowInbox, IconWarning } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { ThirdPartyAgreementPreviewDialog_agreement$key } from "#/__generated__/core/ThirdPartyAgreementPreviewDialog_agreement.graphql";
import { CSVPreview } from "#/components/documents/CSVPreview";
import { PDFPreview } from "#/components/documents/PDFPreview";
import { TextPreview } from "#/components/documents/TextPreview";

const agreementFragment = graphql`
  fragment ThirdPartyAgreementPreviewDialog_agreement on ThirdPartyAgreement {
    type
    file {
      fileName
      mimeType
      downloadUrl
    }
  }
`;

type Props = {
  agreementKey: ThirdPartyAgreementPreviewDialog_agreement$key;
  onClose: () => void;
};

export function ThirdPartyAgreementPreviewDialog({ agreementKey, onClose }: Props) {
  const { t } = useTranslation();
  const agreement = useFragment(agreementFragment, agreementKey);
  const file = agreement.file;
  const mimeType = file.mimeType?.split(";")[0]?.trim().toLowerCase() ?? "";
  const fileName = file.fileName.toLowerCase();
  const isCSV
    = mimeType === "text/csv"
      || mimeType === "application/csv"
      || fileName.endsWith(".csv");
  const isMarkdown
    = mimeType === "text/markdown"
      || fileName.endsWith(".md")
      || fileName.endsWith(".markdown");
  const isText
    = mimeType.startsWith("text/")
      || fileName.endsWith(".txt");

  let preview;
  if (mimeType.startsWith("image/")) {
    preview = (
      <img
        src={file.downloadUrl}
        alt={file.fileName}
        className="max-h-[70vh] object-contain"
      />
    );
  } else if (mimeType.includes("pdf") || fileName.endsWith(".pdf")) {
    preview = (
      <div className="h-[70vh] overflow-hidden">
        <PDFPreview src={file.downloadUrl} name={file.fileName} />
      </div>
    );
  } else if (isCSV) {
    preview = (
      <CSVPreview
        src={file.downloadUrl}
        columnLabel={t("thirdPartyAgreementPreviewDialog.csv.column")}
        emptyMessage={t("thirdPartyAgreementPreviewDialog.csv.empty")}
        errorMessage={t("thirdPartyAgreementPreviewDialog.csv.error")}
        retryLabel={t("thirdPartyAgreementPreviewDialog.csv.retry")}
        truncatedMessage={t("thirdPartyAgreementPreviewDialog.csv.truncated")}
      />
    );
  } else if (isMarkdown || isText) {
    preview = (
      <TextPreview
        src={file.downloadUrl}
        emptyMessage={t("thirdPartyAgreementPreviewDialog.text.empty")}
        errorMessage={t("thirdPartyAgreementPreviewDialog.text.error")}
        format={isMarkdown ? "markdown" : "text"}
        retryLabel={t("thirdPartyAgreementPreviewDialog.text.retry")}
        truncatedMessage={t("thirdPartyAgreementPreviewDialog.text.truncated")}
      />
    );
  } else {
    preview = (
      <div className="flex flex-col items-center gap-2 justify-center">
        <IconWarning size={20} />
        <p className="text-txt-secondary text-center">
          {t("thirdPartyAgreementPreviewDialog.previewUnavailable", {
            mimeType: file.mimeType,
          })}
        </p>
        <Button asChild variant="secondary" icon={IconArrowInbox}>
          <a href={file.downloadUrl} target="_blank" rel="noreferrer">
            {t("thirdPartyAgreementPreviewDialog.actions.download")}
          </a>
        </Button>
      </div>
    );
  }

  return (
    <Dialog
      defaultOpen
      title={t(`thirdPartyAgreementsPage.types.${agreement.type}`)}
      onClose={onClose}
      className="max-w-4xl"
    >
      <DialogContent padded>{preview}</DialogContent>
    </Dialog>
  );
}
