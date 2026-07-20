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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type FormEvent, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { useSubscribeToMailingList } from "#/lib/mailingList/useSubscribeToMailingList";

interface SubscribeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  // Relay store id of the current trust center (for the subscribe updater).
  trustCenterId: string;
  // Verified viewer email, shown read-only (server attributes the subscription).
  viewerEmail: string;
  organizationName: string;
}

// Auth-gated mailing-list subscribe confirmation. The form only mounts while
// open so each open starts clean without a reset effect. Dismiss is blocked
// while the mutation is in flight so a Cancel/Escape cannot race a reopen.
export function SubscribeDialog({
  open,
  onOpenChange,
  trustCenterId,
  viewerEmail,
  organizationName,
}: SubscribeDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (!next && isSubmitting) {
          return;
        }
        onOpenChange(next);
      }}
    >
      <DialogPopup>
        {open && (
          <SubscribeForm
            onClose={() => onOpenChange(false)}
            onSubmittingChange={setIsSubmitting}
            trustCenterId={trustCenterId}
            viewerEmail={viewerEmail}
            organizationName={organizationName}
          />
        )}
      </DialogPopup>
    </Dialog>
  );
}

interface SubscribeFormProps {
  onClose: () => void;
  onSubmittingChange: (submitting: boolean) => void;
  trustCenterId: string;
  viewerEmail: string;
  organizationName: string;
}

function SubscribeForm({
  onClose,
  onSubmittingChange,
  trustCenterId,
  viewerEmail,
  organizationName,
}: SubscribeFormProps) {
  const { t } = useTranslation("updates");
  const [subscribe, isSubscribing] = useSubscribeToMailingList(trustCenterId);
  const aliveRef = useRef(true);

  useEffect(() => {
    aliveRef.current = true;
    return () => {
      aliveRef.current = false;
    };
  }, []);

  useEffect(() => {
    onSubmittingChange(isSubscribing);
    return () => {
      onSubmittingChange(false);
    };
  }, [isSubscribing, onSubmittingChange]);

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (isSubscribing) {
      return;
    }
    try {
      await subscribe();
      if (aliveRef.current) {
        onClose();
      }
    } catch {
      // Errors are surfaced by the mutation notifier; keep the form open.
    }
  };

  return (
    <form className="flex flex-col gap-4" onSubmit={(e) => { void onSubmit(e); }}>
      <DialogHeader>
        <DialogTitle>{t("dialog.title")}</DialogTitle>
        <DialogDescription>{t("dialog.description")}</DialogDescription>
      </DialogHeader>

      <DialogBody>
        <div className="flex flex-col gap-4">
          <Field label={t("dialog.email")}>
            <TextField value={viewerEmail} readOnly disabled />
          </Field>
          <Text size={1} color="faint">
            {t("dialog.consent", { name: organizationName })}
          </Text>
        </div>
      </DialogBody>

      <DialogFooter>
        <Button
          type="button"
          variant="soft"
          color="neutral"
          disabled={isSubscribing}
          onClick={onClose}
        >
          {t("dialog.cancel")}
        </Button>
        <Button type="submit" variant="solid" color="neutral" highContrast loading={isSubscribing}>
          {t("dialog.submit")}
        </Button>
      </DialogFooter>
    </form>
  );
}
