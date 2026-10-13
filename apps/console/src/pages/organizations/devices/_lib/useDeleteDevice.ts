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

import type { useDeleteDeviceMutation } from "#/__generated__/core/useDeleteDeviceMutation.graphql";
import { type MutationFeedback, useMutation } from "#/lib/relay/useMutation";

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

interface DeleteDeviceOptions {
  organizationId: string;
  connectionId?: DataID;
}

export function useDeleteDevice(options: DeleteDeviceOptions) {
  const { t } = useTranslation("organizations/devices");
  const organizationId = options.organizationId;
  const connectionId = options.connectionId;

  const [deleteDevice, isDeleting] = useMutation<useDeleteDeviceMutation>(
    deleteDeviceMutation,
    {
      successMessage: t("devices.messages.deleted"),
      errorToast: t("devices.errors.delete"),
    },
  );

  const remove = useCallback(
    (deviceId: string, feedback?: MutationFeedback) => {
      const connections = [
        connectionId
        ?? ConnectionHandler.getConnectionID(organizationId, DevicesConnectionKey),
      ];
      return deleteDevice({
        variables: {
          input: { deviceId },
          connections,
        },
      }, feedback);
    },
    [deleteDevice, organizationId, connectionId],
  );

  return [remove, isDeleting] as const;
}
