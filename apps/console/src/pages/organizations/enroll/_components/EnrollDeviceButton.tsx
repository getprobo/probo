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

import { formatError } from "@probo/helpers";
import {
  Button,
  IconArrowsClockwise,
  IconCircleCheck,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, useMutation, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EnrollDeviceButtonMutation } from "#/__generated__/core/EnrollDeviceButtonMutation.graphql";
import type { EnrollDeviceButtonStatusQuery } from "#/__generated__/core/EnrollDeviceButtonStatusQuery.graphql";

const POLL_INTERVAL_MS = 3000;
const POLL_TIMEOUT_MS = 15 * 60 * 1000;

const enrollDeviceButtonMutation = graphql`
  mutation EnrollDeviceButtonMutation($input: EnrollDeviceInput!) {
    enrollDevice(input: $input) {
      enrollmentUrl
      device {
        id
      }
    }
  }
`;

const enrollDeviceButtonStatusQuery = graphql`
  query EnrollDeviceButtonStatusQuery($deviceId: ID!) {
    viewer @required(action: THROW) {
      enrolledDevice(id: $deviceId) {
        id
        state
        hostname
      }
    }
  }
`;

export interface EnrollmentSession {
  organizationId: string;
  deviceId: string;
  deepLink: string;
}

interface EnrollDeviceButtonProps {
  organizationId: string | null;
  session?: EnrollmentSession | null;
  onSessionCreated?: (session: EnrollmentSession) => void;
  onComplete?: () => void;
}

function EnrollDeviceButtonContent(
  {
    organizationId,
    session,
    onSessionCreated,
    onComplete,
  }: EnrollDeviceButtonProps,
) {
  const { t } = useTranslation();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const matchingSession
    = session && organizationId && session.organizationId === organizationId
      ? session
      : null;
  const [createdSession, setCreatedSession] = useState<
    Pick<EnrollmentSession, "deviceId" | "deepLink"> | null
  >(null);
  const deepLink = matchingSession?.deepLink ?? createdSession?.deepLink ?? null;
  const deviceId = matchingSession?.deviceId ?? createdSession?.deviceId ?? null;
  const [isWaitingForActivity, setIsWaitingForActivity] = useState(false);
  const [isEnrollmentComplete, setIsEnrollmentComplete] = useState(false);
  const [hasTimedOut, setHasTimedOut] = useState(false);
  const [deviceHostname, setDeviceHostname] = useState<string | null>(null);

  const [enrollDevice, isCreating]
    = useMutation<EnrollDeviceButtonMutation>(enrollDeviceButtonMutation);

  useEffect(() => {
    if (!isWaitingForActivity || !deviceId) {
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

    const runPoll = async () => {
      if (cancelled) {
        return;
      }

      if (document.hidden) {
        scheduleNext();
        return;
      }

      try {
        const data = await fetchQuery<EnrollDeviceButtonStatusQuery>(
          environment,
          enrollDeviceButtonStatusQuery,
          { deviceId },
          { fetchPolicy: "network-only" },
        ).toPromise();

        if (cancelled) {
          return;
        }

        const device = data?.viewer.enrolledDevice;
        if (device == null) {
          if (Date.now() > deadline) {
            setIsWaitingForActivity(false);
            setHasTimedOut(true);
            return;
          }

          scheduleNext();
          return;
        }

        setDeviceHostname(device.hostname ?? null);

        if (device.state === "ACTIVE") {
          setIsEnrollmentComplete(true);
          setIsWaitingForActivity(false);
          onComplete?.();
          return;
        }
      } catch {
        if (Date.now() > deadline) {
          setIsWaitingForActivity(false);
          setHasTimedOut(true);
          return;
        }

        scheduleNext();
        return;
      }

      if (Date.now() > deadline) {
        setIsWaitingForActivity(false);
        setHasTimedOut(true);
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
  }, [deviceId, environment, isWaitingForActivity, onComplete]);

  function openAgent(nextDeepLink: string) {
    setHasTimedOut(false);
    setIsWaitingForActivity(true);
    window.location.assign(nextDeepLink);
  }

  function handleOpenAgent() {
    if (deepLink) {
      openAgent(deepLink);
      return;
    }

    if (!organizationId) {
      return;
    }

    enrollDevice({
      variables: {
        input: {
          organizationId,
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast({
            title: t("common.error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }

        const payload = response.enrollDevice;
        const nextDeepLink = payload.enrollmentUrl;
        const nextDeviceId = payload.device.id;
        setCreatedSession({ deviceId: nextDeviceId, deepLink: nextDeepLink });
        onSessionCreated?.({
          organizationId,
          deviceId: nextDeviceId,
          deepLink: nextDeepLink,
        });
        openAgent(nextDeepLink);
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: formatError(
            t("devices.errors.create"),
            error,
          ),
          variant: "error",
        });
      },
    });
  }

  if (isEnrollmentComplete) {
    return (
      <div className="space-y-1">
        <div className="flex items-center gap-2 text-sm font-medium text-txt-success">
          <IconCircleCheck size={18} />
          {deviceHostname
            ? t("deviceEnrollment.status.enrolledWithHostname", { hostname: deviceHostname })
            : t("deviceEnrollment.status.enrolled")}
        </div>
        <p className="text-sm text-txt-secondary">
          {t("deviceEnrollment.status.closeWindow")}
        </p>
      </div>
    );
  }

  if (isWaitingForActivity) {
    return (
      <div className="flex items-center gap-2 text-sm text-txt-secondary">
        <IconArrowsClockwise size={18} className="animate-spin" />
        {t("deviceEnrollment.status.waitingForCheckIn")}
      </div>
    );
  }

  return (
    <>
      {hasTimedOut && (
        <p className="w-full text-sm text-txt-secondary">
          {t("deviceEnrollment.status.timedOut")}
        </p>
      )}
      <Button
        onClick={handleOpenAgent}
        disabled={isCreating || organizationId == null}
      >
        {isCreating
          ? t("deviceEnrollment.actions.preparing")
          : hasTimedOut
            ? t("common.actions.tryAgain")
            : t("deviceEnrollment.actions.openAgent")}
      </Button>
    </>
  );
}

export function EnrollDeviceButton(props: EnrollDeviceButtonProps) {
  return (
    <EnrollDeviceButtonContent
      key={props.organizationId ?? "none"}
      {...props}
    />
  );
}
