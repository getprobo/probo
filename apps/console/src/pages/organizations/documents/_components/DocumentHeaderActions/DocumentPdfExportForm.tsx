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

import { DownloadSimpleIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent, use, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentPdfExportForm_version$key } from "#/__generated__/core/DocumentPdfExportForm_version.graphql";
import { CurrentUser } from "#/providers/CurrentUser";

import { useExportDocumentVersionPdf } from "../../_lib/useExportDocumentVersionPdf";
import { documentPdfExportForm } from "../../variants";

const documentPdfExportFormFragment = graphql`
  fragment DocumentPdfExportForm_version on DocumentVersion {
    id
    title
    major
    minor
  }
`;

interface DocumentPdfExportFormProps {
  versionKey: DocumentPdfExportForm_version$key;
  onDone: () => void;
}

// Exports the latest published version. The watermark defaults to the
// current user's email, which the server also uses when no text is sent.
export function DocumentPdfExportForm({ versionKey, onDone }: DocumentPdfExportFormProps) {
  const version = useFragment(documentPdfExportFormFragment, versionKey);
  const { t } = useTranslation("organizations/documents");
  const { email } = use(CurrentUser);
  const [exportPdf, isExporting] = useExportDocumentVersionPdf();
  const [withSignatures, setWithSignatures] = useState(false);
  const [withWatermark, setWithWatermark] = useState(false);
  const [watermarkText, setWatermarkText] = useState(email);
  const { form, body, option, optionText } = documentPdfExportForm();

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const text = watermarkText.trim();

    void exportPdf({
      variables: {
        input: {
          documentVersionId: version.id,
          withSignatures,
          withWatermark,
          watermarkText: withWatermark && text !== "" && text !== email ? text : null,
        },
      },
    }).then((response) => {
      const link = window.document.createElement("a");
      link.href = response.exportDocumentVersionPDF.data;
      link.download = `${version.title}-v${version.major}.${version.minor}.pdf`;
      window.document.body.appendChild(link);
      link.click();
      link.remove();
      onDone();
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <form className={form()} onSubmit={handleSubmit}>
      <DialogHeader>
        <DialogTitle>{t("details.actions.download.title")}</DialogTitle>
        <DialogDescription>
          {t("details.actions.download.description", {
            version: `${version.major}.${version.minor}`,
          })}
        </DialogDescription>
      </DialogHeader>
      <DialogBody>
        <div className={body()}>
          <label className={option()}>
            <Checkbox
              checked={withSignatures}
              disabled={isExporting}
              onCheckedChange={setWithSignatures}
            />
            <span className={optionText()}>
              <Text size={2} weight="medium">{t("details.actions.download.signatures")}</Text>
              <Text size={1} color="faint">{t("details.actions.download.signaturesDescription")}</Text>
            </span>
          </label>
          <label className={option()}>
            <Checkbox
              checked={withWatermark}
              disabled={isExporting}
              onCheckedChange={setWithWatermark}
            />
            <span className={optionText()}>
              <Text size={2} weight="medium">{t("details.actions.download.watermark")}</Text>
              <Text size={1} color="faint">{t("details.actions.download.watermarkDescription")}</Text>
            </span>
          </label>
          {withWatermark && (
            <Field label={t("details.actions.download.watermarkText")}>
              <TextField
                value={watermarkText}
                disabled={isExporting}
                placeholder={t("details.actions.download.watermarkPlaceholder")}
                onValueChange={setWatermarkText}
              />
            </Field>
          )}
        </div>
      </DialogBody>
      <DialogFooter>
        <DialogClose
          render={(
            <Button variant="soft" color="neutral" disabled={isExporting}>
              {t("details.actions.cancel")}
            </Button>
          )}
        />
        <Button
          type="submit"
          variant="solid"
          color="neutral"
          highContrast
          loading={isExporting}
          iconStart={<DownloadSimpleIcon />}
        >
          {t("details.actions.download.submit")}
        </Button>
      </DialogFooter>
    </form>
  );
}
