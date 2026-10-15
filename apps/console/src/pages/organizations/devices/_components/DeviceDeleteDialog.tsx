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
import { useTranslation } from "react-i18next";
import type { DataID } from "relay-runtime";

import { ownerLabel } from "../_lib/deviceDisplay";
import { useDeleteDevices } from "../_lib/useDeleteDevices";

interface DeviceDeleteTarget {
  id: string;
  ownerName?: string | null;
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
  const [remove, isDeleting] = useDeleteDevices({ organizationId, connectionId });
  const name = ownerLabel(devices[0]?.ownerName, t("devices.values.unassigned"));
  const title = devices.length === 1
    ? t("devices.confirmations.delete", { owner: name })
    : t("devices.confirmations.deleteMany", { count: devices.length });
  const warning = devices.length === 1
    ? t("devices.confirmations.deleteWarning")
    : t("devices.confirmations.deleteWarningMany");
  const successMessage = devices.length === 1
    ? t("devices.messages.deleted")
    : t("devices.messages.deletedMany", { count: devices.length });

  function handleDelete() {
    void (async () => {
      try {
        await remove(
          devices.map(device => device.id),
          { successMessage },
        );
        onOpenChange(false);
        onDeleted?.();
      } catch {
        // useMutation already reported the failure.
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
