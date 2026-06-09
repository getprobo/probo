// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError, type GraphQLError, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconArrowsClockwise,
  IconCircleCheck,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import { fetchQuery, useMutation, useRelayEnvironment } from "react-relay";
import { graphql } from "relay-runtime";

import type { EnrollDeviceButtonMutation } from "#/__generated__/core/EnrollDeviceButtonMutation.graphql";
import type { EnrollDeviceButtonStatusQuery } from "#/__generated__/core/EnrollDeviceButtonStatusQuery.graphql";

const TOKEN_VALIDITY_SECONDS = 60 * 60 * 24 * 7;
const TOKEN_DEFAULT_MAX_USES = 25;
const POLL_INTERVAL_MS = 3000;
const POLL_TIMEOUT_MS = 15 * 60 * 1000;

const enrollDeviceButtonMutation = graphql`
  mutation EnrollDeviceButtonMutation(
    $input: CreateDeviceEnrollmentTokenInput!
  ) {
    createDeviceEnrollmentToken(input: $input) {
      secret
      enrollmentToken {
        id
      }
    }
  }
`;

const enrollDeviceButtonStatusQuery = graphql`
  query EnrollDeviceButtonStatusQuery($enrollmentTokenId: ID!) {
    deviceEnrollmentStatus(enrollmentTokenId: $enrollmentTokenId) {
      state
      device {
        id
        hostname
      }
    }
  }
`;

interface EnrollDeviceButtonProps {
  organizationId: string | null;
  onComplete?: () => void;
}

export function EnrollDeviceButton(
  { organizationId, onComplete }: EnrollDeviceButtonProps,
) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const environment = useRelayEnvironment();
  const [deepLink, setDeepLink] = useState<string | null>(null);
  const [enrollmentTokenId, setEnrollmentTokenId] = useState<string | null>(null);
  const [isWaitingForActivity, setIsWaitingForActivity] = useState(false);
  const [isEnrollmentComplete, setIsEnrollmentComplete] = useState(false);
  const [hasTimedOut, setHasTimedOut] = useState(false);
  const [deviceHostname, setDeviceHostname] = useState<string | null>(null);

  const [createEnrollmentToken, isCreating]
    = useMutation<EnrollDeviceButtonMutation>(enrollDeviceButtonMutation);

  useEffect(() => {
    let isCurrentRequest = true;

    if (!organizationId) {
      return () => {
        isCurrentRequest = false;
      };
    }

    createEnrollmentToken({
      variables: {
        input: {
          organizationId,
          name: __("Device enrollment"),
          validitySeconds: TOKEN_VALIDITY_SECONDS,
          maxUses: TOKEN_DEFAULT_MAX_USES,
        },
      },
      onCompleted(response, errors) {
        if (!isCurrentRequest) {
          return;
        }

        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }

        const payload = response.createDeviceEnrollmentToken;
        const url = new URL("probo://enroll");
        url.searchParams.set("server", window.location.origin);
        url.searchParams.set("token", payload.secret);
        setDeepLink(url.toString());
        setEnrollmentTokenId(payload.enrollmentToken.id);
      },
      onError(error) {
        if (!isCurrentRequest) {
          return;
        }

        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create enrollment token"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });

    return () => {
      isCurrentRequest = false;
    };
  }, [__, createEnrollmentToken, organizationId, toast]);

  useEffect(() => {
    if (!isWaitingForActivity || !enrollmentTokenId) {
      return;
    }

    const deadline = Date.now() + POLL_TIMEOUT_MS;

    const interval = setInterval(() => {
      if (Date.now() > deadline) {
        setIsWaitingForActivity(false);
        setHasTimedOut(true);
        return;
      }

      if (document.hidden) {
        return;
      }

      fetchQuery<EnrollDeviceButtonStatusQuery>(
        environment,
        enrollDeviceButtonStatusQuery,
        { enrollmentTokenId },
        { fetchPolicy: "network-only" },
      ).subscribe({
        next(data: EnrollDeviceButtonStatusQuery["response"]) {
          const status = data.deviceEnrollmentStatus;
          setDeviceHostname(status.device?.hostname ?? null);

          if (status.state === "FIRST_ACTIVITY_RECEIVED") {
            setIsEnrollmentComplete(true);
            setIsWaitingForActivity(false);
            onComplete?.();
          }
        },
      });
    }, POLL_INTERVAL_MS);

    return () => clearInterval(interval);
  }, [enrollmentTokenId, environment, isWaitingForActivity, onComplete]);

  function handleOpenAgent() {
    if (!deepLink) {
      return;
    }

    setHasTimedOut(false);
    setIsWaitingForActivity(true);
    window.location.assign(deepLink);
  }

  if (isEnrollmentComplete) {
    return (
      <div className="space-y-1">
        <div className="flex items-center gap-2 text-sm font-medium text-txt-success">
          <IconCircleCheck size={18} />
          {deviceHostname
            ? sprintf(__("%s is enrolled."), deviceHostname)
            : __("This device is enrolled.")}
        </div>
        <p className="text-sm text-txt-secondary">
          {__("You can close this window.")}
        </p>
      </div>
    );
  }

  if (isWaitingForActivity) {
    return (
      <div className="flex items-center gap-2 text-sm text-txt-secondary">
        <IconArrowsClockwise size={18} className="animate-spin" />
        {__("Waiting for the agent's first check-in…")}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {hasTimedOut && (
        <p className="text-sm text-txt-secondary">
          {__(
            "We haven't heard from the agent yet. Make sure the desktop agent is installed and running, then try again.",
          )}
        </p>
      )}
      <Button onClick={handleOpenAgent} disabled={!deepLink}>
        {isCreating
          ? __("Preparing…")
          : hasTimedOut
            ? __("Try again")
            : __("Open Probo agent")}
      </Button>
    </div>
  );
}
