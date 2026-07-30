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

import { useConfirm } from "@probo/ui";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import type { useRevokeDeviceMutation } from "#/__generated__/core/useRevokeDeviceMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

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
  const confirm = useConfirm();
  const pendingLabel = t("devices.values.pending");

  const [revokeDevice, isRevoking] = useMutation<useRevokeDeviceMutation>(
    revokeDeviceMutation,
    {
      successMessage: t("devices.messages.revoked"),
      errorToast: t("devices.errors.revoke"),
    },
  );

  const confirmRevoke = useCallback(
    (device: RevokeDeviceInput) => {
      confirm(
        () =>
          revokeDevice({
            variables: { input: { deviceId: device.id } },
          }),
        {
          message: t("devices.confirmations.revoke", {
            hostname: displayValue(device.hostname, pendingLabel),
          }),
          variant: "danger",
          label: t("devices.actions.revoke"),
        },
      );
    },
    [t, confirm, pendingLabel, revokeDevice],
  );

  return [confirmRevoke, isRevoking] as const;
}
