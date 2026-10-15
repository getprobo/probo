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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { ConnectionHandler, type DataID, graphql } from "relay-runtime";

import type { useDeleteDevicesMutation } from "#/__generated__/core/useDeleteDevicesMutation.graphql";
import { type MutationFeedback, useMutation } from "#/lib/relay/useMutation";

export const DevicesConnectionKey = "DevicesPage_devices";

const deleteDevicesMutation = graphql`
  mutation useDeleteDevicesMutation($input: DeleteDevicesInput!) {
    deleteDevices(input: $input) {
      # @deleteEdge accepts one id. Unlink every returned id in the updater.
      # Do not use @deleteRecord; DeviceLayout reads the node with
      # @required(action: THROW).
      deletedDeviceIds
    }
  }
`;

interface DeleteDevicesOptions {
  organizationId: string;
  connectionId?: DataID;
}

export function useDeleteDevices(options: DeleteDevicesOptions) {
  const { t } = useTranslation("organizations/devices");
  const organizationId = options.organizationId;
  const connectionId = options.connectionId;

  const [deleteDevices, isDeleting] = useMutation<useDeleteDevicesMutation>(
    deleteDevicesMutation,
    {
      successMessage: t("devices.messages.deleted"),
      errorToast: t("devices.errors.delete"),
    },
  );

  const remove = useCallback(
    (deviceIds: readonly string[], feedback?: MutationFeedback) => {
      const connection = connectionId
        ?? ConnectionHandler.getConnectionID(organizationId, DevicesConnectionKey);

      return deleteDevices({
        variables: {
          input: { deviceIds },
        },
        updater(store) {
          const payload = store.getRootField("deleteDevices");
          const deleted: unknown = payload?.getValue("deletedDeviceIds");
          const record = store.get(connection);
          if (!record || !Array.isArray(deleted)) {
            return;
          }

          for (const id of deleted) {
            if (typeof id === "string") {
              ConnectionHandler.deleteNode(record, id);
            }
          }

          const totalCount: unknown = record.getValue("totalCount");
          if (typeof totalCount === "number") {
            record.setValue(Math.max(0, totalCount - deleted.length), "totalCount");
          }
        },
      }, feedback);
    },
    [deleteDevices, organizationId, connectionId],
  );

  return [remove, isDeleting] as const;
}
