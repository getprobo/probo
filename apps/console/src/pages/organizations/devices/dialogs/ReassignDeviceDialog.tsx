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

import { formatError, type GraphQLError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { ReassignDeviceDialog_device$key } from "#/__generated__/core/ReassignDeviceDialog_device.graphql";
import type { ReassignDeviceDialogMutation } from "#/__generated__/core/ReassignDeviceDialogMutation.graphql";

import { DeviceOwnerSelect } from "../_components/DeviceOwnerSelect";

const reassignDeviceDialogFragment = graphql`
  fragment ReassignDeviceDialog_device on Device {
    id
    owner {
      id
      ...DeviceOwnerSelect_profile
    }
  }
`;

const reassignDeviceMutation = graphql`
  mutation ReassignDeviceDialogMutation($input: SetDeviceOwnerInput!) {
    setDeviceOwner(input: $input) {
      device {
        id
        owner {
          id
          fullName
        }
      }
    }
  }
`;

interface ReassignDeviceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  deviceKey: ReassignDeviceDialog_device$key;
  organizationId: string;
}

export function ReassignDeviceDialog({
  open,
  onOpenChange,
  deviceKey,
  organizationId,
}: ReassignDeviceDialogProps) {
  const { t } = useTranslation("organizations/devices");
  const toast = useToast();
  const device = useFragment(reassignDeviceDialogFragment, deviceKey);
  const serverOwnerId = device.owner?.id ?? null;
  const [ownerId, setOwnerId] = useState<string | null>(serverOwnerId);
  const [prevServerOwnerId, setPrevServerOwnerId] = useState(serverOwnerId);
  if (serverOwnerId !== prevServerOwnerId) {
    setPrevServerOwnerId(serverOwnerId);
    setOwnerId(serverOwnerId);
  }
  const [setDeviceOwner, isUpdating] = useMutation<ReassignDeviceDialogMutation>(reassignDeviceMutation);

  function handleOpenChange(next: boolean) {
    if (!next) {
      setOwnerId(serverOwnerId);
    }
    onOpenChange(next);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setDeviceOwner({
      variables: {
        input: {
          deviceId: device.id,
          ownerId: ownerId ?? undefined,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast.add({
            title: t("common.error"),
            description: errors[0].message,
            type: "error",
          });
          return;
        }
        toast.add({
          title: t("common.success"),
          description: t("devices.messages.ownerUpdated"),
          type: "success",
        });
        onOpenChange(false);
      },
      onError(error) {
        toast.add({
          title: t("common.error"),
          description: formatError(t("devices.errors.reassign"), error as GraphQLError),
          type: "error",
        });
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("devices.actions.reassign")}</DialogTitle>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
          <DialogBody>
            {open && (
              <DeviceOwnerSelect
                organizationId={organizationId}
                profileKey={device.owner ?? null}
                value={ownerId}
                onValueChange={setOwnerId}
              />
            )}
          </DialogBody>
          <DialogFooter>
            <Button type="submit" loading={isUpdating}>
              {t("devices.actions.reassign")}
            </Button>
          </DialogFooter>
        </form>
      </DialogPopup>
    </Dialog>
  );
}
