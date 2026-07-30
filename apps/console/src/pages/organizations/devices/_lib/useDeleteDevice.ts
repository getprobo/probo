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

import { useConfirm } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type DataID,
  graphql,
} from "relay-runtime";

import type { useDeleteDeviceMutation } from "#/__generated__/core/useDeleteDeviceMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { displayValue } from "./deviceDisplay";

export const DevicesConnectionKey = "DevicesPage_devices";

const deleteDeviceMutation = graphql`
  mutation useDeleteDeviceMutation(
    $input: DeleteDeviceInput!
    $connections: [ID!]!
  ) {
    deleteDevice(input: $input) {
      # @deleteEdge only unlinks from list connections — do not use
      # @deleteRecord; DeviceLayout reads the node with @required(action: THROW).
      deletedDeviceId @deleteEdge(connections: $connections)
    }
  }
`;

interface DeleteDeviceInput {
  id: string;
  hostname: string | null | undefined;
}

interface DeleteDeviceOptions {
  organizationId: string;
  connectionId?: DataID;
  onDeleted?: () => void;
}

export function useDeleteDevice(options: DeleteDeviceOptions) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const pendingLabel = t("devices.values.pending");
  const organizationId = options.organizationId;
  const connectionId = options.connectionId;
  const onDeleted = options.onDeleted;

  const [deleteDevice, isDeleting] = useMutation<useDeleteDeviceMutation>(
    deleteDeviceMutation,
    {
      successMessage: t("devices.messages.deleted"),
      errorToast: t("devices.errors.delete"),
    },
  );

  const confirmDelete = useCallback(
    (device: DeleteDeviceInput) => {
      const connections = [
        connectionId
        ?? ConnectionHandler.getConnectionID(organizationId, DevicesConnectionKey),
      ];

      confirm(
        async () => {
          await deleteDevice({
            variables: {
              input: { deviceId: device.id },
              connections,
            },
          });
          onDeleted?.();
        },
        {
          message: t("devices.confirmations.delete", {
            hostname: displayValue(device.hostname, pendingLabel),
          }),
          variant: "danger",
          label: t("devices.actions.delete"),
        },
      );
    },
    [
      t,
      confirm,
      pendingLabel,
      deleteDevice,
      organizationId,
      connectionId,
      onDeleted,
    ],
  );

  return [confirmDelete, isDeleting] as const;
}
