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

import { CheckIcon, WarningIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { SegmentedControl } from "@probo/ui/src/v2/SegmentedControl/SegmentedControl";
import { SegmentedControlItem } from "@probo/ui/src/v2/SegmentedControl/SegmentedControlItem";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent, useId, useState } from "react";
import { useTranslation } from "react-i18next";

import {
  rightsRequestFormConfig,
  type SubmittableRightsRequestType,
  submittableRightsRequestTypes,
} from "../_lib/rightsRequest";
import { useCreateRightsRequest } from "../_lib/useCreateRightsRequest";

import { newRequestForm } from "./variants";

interface NewRequestDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // Relay connection id to prepend the created request into.
  connectionId: string;
  // Verified viewer identity, used to prefill the (read-only) email and name.
  viewerEmail: string;
  viewerName: string;
}

// The "New Request" modal. The form lives in a child that only mounts while the
// dialog is open, so each open starts from a clean slate without a reset effect.
export function NewRequestDialog({
  open,
  onOpenChange,
  connectionId,
  viewerEmail,
  viewerName,
}: NewRequestDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <NewRequestForm
          onClose={() => onOpenChange(false)}
          connectionId={connectionId}
          viewerEmail={viewerEmail}
          viewerName={viewerName}
        />
      </DialogPopup>
    </Dialog>
  );
}

interface NewRequestFormProps {
  onClose: () => void;
  connectionId: string;
  viewerEmail: string;
  viewerName: string;
}

function NewRequestForm({ onClose, connectionId, viewerEmail, viewerName }: NewRequestFormProps) {
  const { t } = useTranslation("requests");
  const [submit, isSubmitting] = useCreateRightsRequest();
  const typeLabelId = useId();

  const [type, setType] = useState<SubmittableRightsRequestType>("ACCESS");
  const [name, setName] = useState(viewerName);
  const [details, setDetails] = useState("");
  const [submitted, setSubmitted] = useState(false);

  const config = rightsRequestFormConfig[type];
  const { root, label, success, successIcon } = newRequestForm();

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    try {
      await submit({
        variables: {
          input: {
            requestType: type,
            dataSubject: name.trim() === "" ? null : name.trim(),
            details: details.trim() === "" ? null : details.trim(),
          },
          connections: [connectionId],
        },
      });
      setSubmitted(true);
    } catch {
      // Errors are surfaced by the mutation notifier; keep the form open.
    }
  };

  if (submitted) {
    return (
      <div className={success()}>
        <span className={successIcon()}>
          <CheckIcon weight="bold" />
        </span>
        <div className="flex flex-col gap-1">
          <Text size={3} weight="medium" color="neutral" highContrast>
            {t("dialog.success.title")}
          </Text>
          <Text size={2} color="faint">
            {t("dialog.success.description")}
          </Text>
        </div>
        <Button variant="soft" color="neutral" onClick={onClose}>
          {t("dialog.success.close")}
        </Button>
      </div>
    );
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={(e) => { void onSubmit(e); }}>
      <DialogHeader>
        <DialogTitle>{t("dialog.title")}</DialogTitle>
        <DialogDescription>{t("dialog.description")}</DialogDescription>
      </DialogHeader>

      <DialogBody>
        <div className={root()}>
          <div className={label()}>
            <Text id={typeLabelId} size={2} weight="medium" color="neutral" highContrast>
              {t("dialog.typeLabel")}
            </Text>
            <SegmentedControl
              aria-labelledby={typeLabelId}
              value={type}
              onValueChange={value => setType(value as SubmittableRightsRequestType)}
            >
              {submittableRightsRequestTypes.map(option => (
                <SegmentedControlItem key={option} value={option}>
                  {t(`typeOption.${option}`)}
                </SegmentedControlItem>
              ))}
            </SegmentedControl>
          </div>

          {config.showDeletionWarning && (
            <Callout color="amber" variant="soft" icon={<WarningIcon weight="fill" />}>
              {t("form.deletionWarning")}
            </Callout>
          )}

          <Field label={config.nameOptional ? t("form.nameOptional") : t("form.name")}>
            <TextField
              value={name}
              placeholder={t("form.namePlaceholder")}
              required={!config.nameOptional}
              onChange={e => setName(e.target.value)}
            />
          </Field>

          <Field label={t("form.email")}>
            <TextField value={viewerEmail} readOnly disabled />
          </Field>

          <Field label={t(`form.details.${type}.label`)}>
            <Textarea
              value={details}
              placeholder={t(`form.details.${type}.placeholder`)}
              onChange={e => setDetails(e.target.value)}
            />
          </Field>
        </div>
      </DialogBody>

      <DialogFooter>
        <Button type="button" variant="soft" color="neutral" onClick={onClose}>
          {t("dialog.cancel")}
        </Button>
        <Button type="submit" variant="solid" color="neutral" highContrast loading={isSubmitting}>
          {t("dialog.submit")}
        </Button>
      </DialogFooter>
    </form>
  );
}
