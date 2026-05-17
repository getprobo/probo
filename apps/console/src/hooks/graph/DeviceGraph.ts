// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { promisifyMutation, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { useConfirm } from "@probo/ui";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DeviceGraphAssignDeviceMutation } from "#/__generated__/core/DeviceGraphAssignDeviceMutation.graphql";
import type { DeviceGraphCreateEnrollmentTokenMutation } from "#/__generated__/core/DeviceGraphCreateEnrollmentTokenMutation.graphql";
import type { DeviceGraphRevokeDeviceMutation } from "#/__generated__/core/DeviceGraphRevokeDeviceMutation.graphql";
import type { DeviceGraphRevokeEnrollmentTokenMutation } from "#/__generated__/core/DeviceGraphRevokeEnrollmentTokenMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

/* eslint-disable relay/unused-fields, relay/must-colocate-fragment-spreads */

export const deviceConnectionKey = "DevicesPage_devices";
export const deviceEnrollmentTokenConnectionKey
  = "DevicesEnrollPage_deviceEnrollmentTokens";

export const devicesQuery = graphql`
  query DeviceGraphListQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        id
        canCreateEnrollmentToken: permission(
          action: "core:device-enrollment-token:create"
        )
        canRevokeDevice: permission(action: "core:device:revoke")
        canAssignDevice: permission(action: "core:device:assign")
        ...DeviceGraphPaginatedFragment
      }
    }
  }
`;

export const paginatedDevicesFragment = graphql`
  fragment DeviceGraphPaginatedFragment on Organization
  @refetchable(queryName: "DevicesListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "DeviceOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    devices(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "DevicesListQuery_devices") {
      __id
      edges {
        node {
          id
          hostname
          platform
          osVersion
          agentVersion
          enrolledAt
          lastSeenAt
          revokedAt
          latestPostures {
            id
            checkKey
            status
            observedAt
          }
        }
      }
    }
  }
`;

export const deviceNodeQuery = graphql`
  query DeviceGraphNodeQuery($deviceId: ID!) {
    node(id: $deviceId) {
      id
      ... on Device {
        hostname
        hardwareUuid
        serialNumber
        platform
        osVersion
        agentVersion
        enrolledAt
        lastSeenAt
        revokedAt
        latestPostures {
          id
          checkKey
          status
          observedAt
        }
      }
    }
  }
`;

export const deviceEnrollmentTokensQuery = graphql`
  query DeviceGraphEnrollmentTokensQuery($organizationId: ID!) {
    node(id: $organizationId) {
      ... on Organization {
        id
        canCreateEnrollmentToken: permission(
          action: "core:device-enrollment-token:create"
        )
        deviceEnrollmentTokens(first: 100) {
          edges {
            node {
              id
              name
              createdAt
              expiresAt
              revokedAt
              maxUses
              usedCount
            }
          }
        }
      }
    }
  }
`;

const createEnrollmentTokenMutation = graphql`
  mutation DeviceGraphCreateEnrollmentTokenMutation(
    $input: CreateDeviceEnrollmentTokenInput!
  ) {
    createDeviceEnrollmentToken(input: $input) {
      secret
      enrollmentToken {
        id
        name
        createdAt
        expiresAt
        maxUses
        usedCount
      }
    }
  }
`;

export function useCreateDeviceEnrollmentTokenMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts<DeviceGraphCreateEnrollmentTokenMutation>(
    createEnrollmentTokenMutation,
    {
      successMessage: __(
        "Enrollment token created. Copy it now — it will not be shown again.",
      ),
      errorMessage: __("Failed to create enrollment token"),
    },
  );
}

const revokeEnrollmentTokenMutation = graphql`
  mutation DeviceGraphRevokeEnrollmentTokenMutation(
    $input: RevokeDeviceEnrollmentTokenInput!
  ) {
    revokeDeviceEnrollmentToken(input: $input) {
      enrollmentToken {
        id
        revokedAt
      }
    }
  }
`;

export function useRevokeDeviceEnrollmentToken(
  token: { id?: string },
) {
  const [mutate] = useMutation<DeviceGraphRevokeEnrollmentTokenMutation>(
    revokeEnrollmentTokenMutation,
  );
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!token.id) return;
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: {
            input: { enrollmentTokenId: token.id! },
          },
        }),
      {
        message: __(
          "Revoke this enrollment token? Devices using it can finish their current enrollment but no new devices will be able to use it.",
        ),
      },
    );
  };
}

const revokeDeviceMutation = graphql`
  mutation DeviceGraphRevokeDeviceMutation($input: RevokeDeviceInput!) {
    revokeDevice(input: $input) {
      device {
        id
        revokedAt
      }
    }
  }
`;

export function useRevokeDevice(device: { id?: string; hostname?: string }) {
  const [mutate] = useMutation<DeviceGraphRevokeDeviceMutation>(
    revokeDeviceMutation,
  );
  const confirm = useConfirm();
  const { __ } = useTranslate();

  return () => {
    if (!device.id) return;
    confirm(
      () =>
        promisifyMutation(mutate)({
          variables: { input: { deviceId: device.id! } },
        }),
      {
        message: sprintf(
          __(
            "Revoke device \"%s\"? The agent on the device will stop reporting and must be re-enrolled.",
          ),
          device.hostname ?? device.id,
        ),
      },
    );
  };
}

const assignDeviceMutation = graphql`
  mutation DeviceGraphAssignDeviceMutation($input: AssignDeviceToUserInput!) {
    assignDeviceToUser(input: $input) {
      device {
        id
      }
    }
  }
`;

export function useAssignDeviceToUser() {
  const { __ } = useTranslate();
  return useMutationWithToasts<DeviceGraphAssignDeviceMutation>(
    assignDeviceMutation,
    {
      successMessage: __("Device assigned."),
      errorMessage: __("Failed to assign device"),
    },
  );
}
