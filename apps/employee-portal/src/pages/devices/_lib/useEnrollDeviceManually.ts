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

import { useCallback, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";
import { useParams } from "react-router";

import type { useEnrollDeviceManuallyMutation } from "#/__generated__/core/useEnrollDeviceManuallyMutation.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

const enrollDeviceMutation = graphql`
  mutation useEnrollDeviceManuallyMutation($input: EnrollDeviceInput!) {
    enrollDevice(input: $input) {
      enrollmentToken
      serverUrl
      device {
        id
      }
    }
  }
`;

export type ManualEnrollment = {
  enrollmentToken: string;
  serverUrl: string;
};

export function useEnrollDeviceManually() {
  const { t } = useTranslation("devices");
  const { organizationId } = useParams();
  const [enrollDevice, isCreating] = useMutation<useEnrollDeviceManuallyMutation>(
    enrollDeviceMutation,
    { successMessage: t("addManually.created") },
  );
  const [enrollment, setEnrollment] = useState<ManualEnrollment | null>(null);
  const [failed, setFailed] = useState(false);
  const startedRef = useRef(false);

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  const enrolledOrganizationId = organizationId;

  const create = useCallback(async () => {
    setFailed(false);
    try {
      const response = await enrollDevice({
        variables: {
          input: { organizationId: enrolledOrganizationId },
        },
      });
      setEnrollment({
        enrollmentToken: response.enrollDevice.enrollmentToken,
        serverUrl: response.enrollDevice.serverUrl,
      });
    } catch {
      setFailed(true);
    }
  }, [enrollDevice, enrolledOrganizationId]);

  const start = useCallback(() => {
    if (startedRef.current) {
      return;
    }
    startedRef.current = true;
    void create();
  }, [create]);

  const retry = useCallback(() => {
    void create();
  }, [create]);

  return {
    start,
    retry,
    isCreating,
    enrollment,
    failed,
  };
}
