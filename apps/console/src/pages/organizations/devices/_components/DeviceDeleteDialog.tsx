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

import { TrashIcon, WarningIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { ConfirmDialog } from "@probo/ui/src/v2/Dialog/ConfirmDialog";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import type { DataID } from "relay-runtime";

import { confirmEachDevice } from "../_lib/confirmEachDevice";
import { displayValue } from "../_lib/deviceDisplay";
import { useDeleteDevice } from "../_lib/useDeleteDevice";

interface DeviceDeleteTarget {
  id: string;
  hostname?: string | null;
}

interface DeviceDeleteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
  connectionId?: DataID;
  devices: readonly DeviceDeleteTarget[];
  onDeleted?: () => void;
}

export function DeviceDeleteDialog({
  open,
  onOpenChange,
  organizationId,
  connectionId,
  devices,
  onDeleted,
}: DeviceDeleteDialogProps) {
  const { t } = useTranslation("organizations/devices");
  const toast = useToast();
  const [remove] = useDeleteDevice({ organizationId, connectionId });
  const [isDeleting, setIsDeleting] = useState(false);
  const hostnameLabel = displayValue(devices[0]?.hostname, t("devices.values.pending"));
  const title = devices.length === 1
    ? t("devices.confirmations.delete", { hostname: hostnameLabel })
    : t("devices.confirmations.deleteMany", { count: devices.length });
  const warning = devices.length === 1
    ? t("devices.confirmations.deleteWarning")
    : t("devices.confirmations.deleteWarningMany");
  const successMessage = devices.length === 1
    ? t("devices.messages.deleted")
    : t("devices.messages.deletedMany", { count: devices.length });

  function handleDelete() {
    setIsDeleting(true);
    void (async () => {
      try {
        const confirmed = await confirmEachDevice({
          devices,
          successMessage,
          confirmDevice: (deviceId, feedback) => remove(deviceId, feedback),
          onPartial: (count, total) => {
            toast.add({
              title: t("devices.messages.deletedPartial", { count, total }),
              type: "warning",
            });
          },
        });
        if (confirmed) {
          onOpenChange(false);
          onDeleted?.();
        }
      } finally {
        setIsDeleting(false);
      }
    })();
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={title}
      description={(
        <Callout color="amber" icon={<WarningIcon weight="fill" />}>
          {warning}
        </Callout>
      )}
      cancelLabel={t("common.actions.close")}
      confirmLabel={t("devices.actions.delete")}
      confirmIcon={<TrashIcon />}
      loading={isDeleting}
      confirmDisabled={devices.length === 0}
      onConfirm={handleDelete}
    />
  );
}
