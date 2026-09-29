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

import { FileCsvIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Field } from "@probo/ui/src/v2/form/Field";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { ExportSCIMEventsDialog_exportMutation } from "#/__generated__/iam/ExportSCIMEventsDialog_exportMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { scimPage } from "../variants";

const exportMutation = graphql`
  mutation ExportSCIMEventsDialog_exportMutation(
    $input: RequestSCIMEventExportInput!
  ) {
    requestSCIMEventExport(input: $input) {
      exportJobId
    }
  }
`;

const DAY_MS = 24 * 60 * 60 * 1000;

export interface ExportSCIMEventsDialogProps {
  organizationId: string;
}

export function ExportSCIMEventsDialog({
  organizationId,
}: ExportSCIMEventsDialogProps) {
  const { t, i18n } = useTranslation();
  const [open, setOpen] = useState(false);
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const { exportFields } = scimPage();
  const [requestSCIMEventExport, isExporting]
    = useMutation<ExportSCIMEventsDialog_exportMutation>(
      exportMutation,
      {
        successMessage: t("scimPage.export.messages.success"),
        errorToast: t("scimPage.export.errors.request"),
      },
    );

  function handleExport() {
    if (fromDate === "" || toDate === "" || fromDate > toDate) {
      return;
    }

    void requestSCIMEventExport({
      variables: {
        input: {
          organizationId,
          fromTime: new Date(`${fromDate}T00:00:00Z`).toISOString(),
          toTime: new Date(Date.parse(`${toDate}T00:00:00Z`) + DAY_MS).toISOString(),
        },
      },
    }).then(
      () => {
        setOpen(false);
        setFromDate("");
        setToDate("");
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger
        render={(
          <Button
            variant="soft"
            color="gold"
            iconStart={<FileCsvIcon />}
          >
            {t("scimPage.export.actions.export")}
          </Button>
        )}
      />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("scimPage.export.title")}</DialogTitle>
          <DialogDescription>
            {t("scimPage.export.description")}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <div className={exportFields()}>
            <Field label={t("scimPage.export.fields.from")} required>
              <DateField
                size={2}
                required
                value={fromDate}
                locale={i18n.language}
                onValueChange={setFromDate}
              />
            </Field>
            <Field label={t("scimPage.export.fields.to")} required>
              <DateField
                size={2}
                required
                value={toDate}
                locale={i18n.language}
                min={fromDate === "" ? undefined : fromDate}
                onValueChange={setToDate}
              />
            </Field>
          </div>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("scimPage.export.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="neutral"
            highContrast
            loading={isExporting}
            disabled={fromDate === "" || toDate === "" || fromDate > toDate}
            iconStart={<FileCsvIcon />}
            onClick={handleExport}
          >
            {t("scimPage.export.actions.export")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
