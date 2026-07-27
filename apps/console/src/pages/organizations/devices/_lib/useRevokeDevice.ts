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

import { formatError } from "@probo/helpers";
import { useConfirm, useToast } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { useRevokeDeviceMutation } from "#/__generated__/core/useRevokeDeviceMutation.graphql";

import { displayValue } from "./deviceDisplay";

const revokeDeviceMutation = graphql`
  mutation useRevokeDeviceMutation($input: RevokeDeviceInput!) {
    revokeDevice(input: $input) {
      device {
        id
        revokedAt
        state
        ...DeviceDetailsCard_deviceFragment
      }
    }
  }
`;

interface RevokeDeviceInput {
  id: string;
  hostname: string | null | undefined;
}

export function useRevokeDevice() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const confirm = useConfirm();
  const pendingLabel = t("devices.values.pending");

  const [revokeDevice, isRevoking] = useMutation<useRevokeDeviceMutation>(
    revokeDeviceMutation,
  );

  const confirmRevoke = useCallback(
    (device: RevokeDeviceInput) => {
      confirm(
        () =>
          new Promise<void>((resolve) => {
            revokeDevice({
              variables: { input: { deviceId: device.id } },
              onCompleted(_, errors) {
                if (errors?.length) {
                  toast({
                    title: t("common.error"),
                    description: errors[0].message,
                    variant: "error",
                  });
                } else {
                  toast({
                    title: t("common.success"),
                    description: t("devices.messages.revoked"),
                    variant: "success",
                  });
                }
                resolve();
              },
              onError(error) {
                toast({
                  title: t("common.error"),
                  description: formatError(
                    t("devices.errors.revoke"),
                    error,
                  ),
                  variant: "error",
                });
                resolve();
              },
            });
          }),
        {
          message: t("devices.confirmations.revoke", { hostname: displayValue(device.hostname, pendingLabel) }),
          variant: "danger",
          label: t("devices.actions.revoke"),
        },
      );
    },
    [t, confirm, pendingLabel, revokeDevice, toast],
  );

  return [confirmRevoke, isRevoking] as const;
}
