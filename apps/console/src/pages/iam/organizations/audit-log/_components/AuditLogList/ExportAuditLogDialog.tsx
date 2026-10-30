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
import { DateRangeField } from "@probo/ui/src/v2/form/DateRangeField";
import { Field } from "@probo/ui/src/v2/form/Field";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { ExportAuditLogDialog_exportMutation } from "#/__generated__/iam/ExportAuditLogDialog_exportMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { exportAuditLogDialog } from "../variants";

const AUDIT_LOG_NS = "iam/organizations/audit-log";
const DAY_MS = 24 * 60 * 60 * 1000;

const exportMutation = graphql`
  mutation ExportAuditLogDialog_exportMutation(
    $input: RequestAuditLogExportInput!
  ) {
    requestAuditLogExport(input: $input) {
      exportJobId
    }
  }
`;

export function ExportAuditLogDialog() {
  const organizationId = useOrganizationId();
  const { t, i18n } = useTranslation(AUDIT_LOG_NS);
  const [open, setOpen] = useState(false);
  const [fromDate, setFromDate] = useState("");
  const [toDate, setToDate] = useState("");
  const { fields } = exportAuditLogDialog();
  const [requestAuditLogExport, isExporting]
    = useMutation<ExportAuditLogDialog_exportMutation>(
      exportMutation,
      {
        successMessage: t("auditLogPage.export.messages.success"),
        errorToast: t("auditLogPage.export.errors.request"),
      },
    );
  function handleExport() {
    if (fromDate === "" || toDate === "") {
      return;
    }

    void requestAuditLogExport({
      variables: {
        input: {
          organizationId,
          fromTime: new Date(`${fromDate}T00:00:00Z`).toISOString(),
          toTime: new Date(Date.parse(`${toDate}T00:00:00Z`) + DAY_MS).toISOString(),
        },
      },
    }).then(() => {
      setOpen(false);
      setFromDate("");
      setToDate("");
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
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
            {t("auditLogPage.export.actions.export")}
          </Button>
        )}
      />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("auditLogPage.export.title")}</DialogTitle>
          <DialogDescription>
            {t("auditLogPage.export.description")}
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <div className={fields()}>
            <Field label={t("auditLogPage.export.fields.range")} required>
              <DateRangeField
                required
                nullable
                value={{ from: fromDate, to: toDate }}
                locale={i18n.language}
                placeholder={t("auditLogPage.filters.datesPlaceholder")}
                hint={t("auditLogPage.filters.datesHint")}
                clearLabel={t("auditLogPage.filters.clearDates")}
                aria-label={t("auditLogPage.export.fields.range")}
                onValueChange={(next) => {
                  setFromDate(next.from);
                  setToDate(next.to);
                }}
              />
            </Field>
          </div>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("auditLogPage.export.actions.cancel")}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color="neutral"
            highContrast
            loading={isExporting}
            disabled={fromDate === "" || toDate === ""}
            iconStart={<FileCsvIcon />}
            onClick={handleExport}
          >
            {t("auditLogPage.export.actions.export")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
