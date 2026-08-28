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

import { useEffect, useState } from "react";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";
import { useParams } from "react-router";

import type { useEnrollDeviceMutation } from "#/__generated__/core/useEnrollDeviceMutation.graphql";
import type { useEnrollDeviceStatusQuery } from "#/__generated__/core/useEnrollDeviceStatusQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";

const POLL_INTERVAL_MS = 3000;
const POLL_TIMEOUT_MS = 15 * 60 * 1000;

const enrollDeviceMutation = graphql`
  mutation useEnrollDeviceMutation($input: EnrollDeviceInput!) {
    enrollDevice(input: $input) {
      enrollmentUrl
      device {
        id
      }
    }
  }
`;

const enrollDeviceStatusQuery = graphql`
  query useEnrollDeviceStatusQuery($deviceId: ID!) @throwOnFieldError {
    viewer @required(action: THROW) {
      enrolledDevice(id: $deviceId) {
        id
        state
        hostname
      }
    }
  }
`;

export function useEnrollDevice() {
  const { organizationId } = useParams();
  const environment = useRelayEnvironment();
  const [enrollDevice, isCreating] = useMutation<useEnrollDeviceMutation>(
    enrollDeviceMutation,
    { errorToast: false },
  );
  const [deviceId, setDeviceId] = useState<string | null>(null);
  const [deepLink, setDeepLink] = useState<string | null>(null);
  const [isWaiting, setIsWaiting] = useState(false);
  const [isComplete, setIsComplete] = useState(false);
  const [hasTimedOut, setHasTimedOut] = useState(false);
  const [hostname, setHostname] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (!isWaiting || deviceId === null) {
      return;
    }

    let cancelled = false;
    let timeoutId: ReturnType<typeof setTimeout> | undefined;
    const deadline = Date.now() + POLL_TIMEOUT_MS;

    const scheduleNext = () => {
      if (!cancelled) {
        timeoutId = setTimeout(runPoll, POLL_INTERVAL_MS);
      }
    };

    const finishTimeout = () => {
      setIsWaiting(false);
      setHasTimedOut(true);
    };

    const runPoll = async () => {
      if (cancelled) {
        return;
      }

      if (document.hidden) {
        if (Date.now() > deadline) {
          finishTimeout();
          return;
        }
        scheduleNext();
        return;
      }

      if (Date.now() > deadline) {
        finishTimeout();
        return;
      }

      try {
        const data = await fetchQuery<useEnrollDeviceStatusQuery>(
          environment,
          enrollDeviceStatusQuery,
          { deviceId },
          { fetchPolicy: "network-only" },
        ).toPromise();

        if (cancelled) {
          return;
        }

        const device = data?.viewer.enrolledDevice;
        if (device === undefined || device === null) {
          scheduleNext();
          return;
        }

        setHostname(device.hostname ?? null);

        if (device.state === "ACTIVE") {
          setIsComplete(true);
          setIsWaiting(false);
          return;
        }
      } catch {
        if (Date.now() > deadline) {
          finishTimeout();
          return;
        }
      }

      if (Date.now() > deadline) {
        finishTimeout();
        return;
      }

      scheduleNext();
    };

    scheduleNext();

    return () => {
      cancelled = true;
      if (timeoutId !== undefined) {
        clearTimeout(timeoutId);
      }
    };
  }, [deviceId, environment, isWaiting]);

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  const enrolledOrganizationId = organizationId;

  async function openAgent() {
    setHasTimedOut(false);
    setFailed(false);

    try {
      if (deepLink !== null) {
        setIsWaiting(true);
        window.location.assign(deepLink);
        return;
      }

      const response = await enrollDevice({
        variables: {
          input: { organizationId: enrolledOrganizationId },
        },
      });
      const payload = response.enrollDevice;
      setDeviceId(payload.device.id);
      setDeepLink(payload.enrollmentUrl);
      setIsWaiting(true);
      window.location.assign(payload.enrollmentUrl);
    } catch {
      setFailed(true);
    }
  }

  return {
    openAgent,
    isCreating,
    isWaiting,
    isComplete,
    hasTimedOut,
    failed,
    hostname,
  };
}
