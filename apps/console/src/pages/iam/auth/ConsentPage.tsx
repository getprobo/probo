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

import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconArrowsClockwise,
  IconEnvelope,
  IconLockOpen,
  IconUser,
  IconUserCircle,
  useToast,
} from "@probo/ui";
import { useCallback, useState } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentPageMutation } from "#/__generated__/iam/ConsentPageMutation.graphql";
import type { ConsentPageQuery } from "#/__generated__/iam/ConsentPageQuery.graphql";

export const consentPageQuery = graphql`
  query ConsentPageQuery($consentId: ID!) {
    node(id: $consentId) @required(action: THROW) {
      ... on Consent {
        id
        application {
          name
        }
        scopes
      }
    }
  }
`;

const approveConsentMutation = graphql`
  mutation ConsentPageMutation($input: ApproveConsentInput!) {
    approveConsent(input: $input) {
      redirectURL
      deviceAuthorized
    }
  }
`;

const scopeLabels: Record<string, string> = {
  openid: "Verify your identity",
  email: "View your email address",
  profile: "View your profile information",
  offline_access: "Stay signed in and access your data while you're away",
};

const scopeIcons: Record<string, React.ReactNode> = {
  openid: <IconUser size={18} className="shrink-0 text-txt-tertiary" />,
  email: <IconEnvelope size={18} className="shrink-0 text-txt-tertiary" />,
  profile: <IconUserCircle size={18} className="shrink-0 text-txt-tertiary" />,
  offline_access: <IconArrowsClockwise size={18} className="shrink-0 text-txt-tertiary" />,
};

export default function ConsentPage(props: {
  queryRef: PreloadedQuery<ConsentPageQuery>;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [deviceResult, setDeviceResult] = useState<"authorized" | "denied" | null>(null);

  const data = usePreloadedQuery(consentPageQuery, props.queryRef);
  usePageTitle(__("Authorize Application"));

  const { node: consent } = data;

  const [approveConsent, isInFlight]
    = useMutation<ConsentPageMutation>(approveConsentMutation);

  const handleAction = useCallback(
    (approved: boolean) => {
      if (!consent.id) return;

      approveConsent({
        variables: {
          input: {
            consentId: consent.id,
            approved,
          },
        },
        onCompleted: (response, errors) => {
          if (errors) {
            toast({
              title: __("Authorization failed"),
              description: formatError(
                __("Something went wrong. Please try again."),
                errors,
              ),
              variant: "error",
            });
            return;
          }

          if (!response.approveConsent) {
            toast({
              title: __("Authorization failed"),
              description: __("Something went wrong. Please try again."),
              variant: "error",
            });
            return;
          }

          if (response.approveConsent.deviceAuthorized != null) {
            setDeviceResult(response.approveConsent.deviceAuthorized ? "authorized" : "denied");
            return;
          }

          if (response.approveConsent.redirectURL) {
            window.location.href = response.approveConsent.redirectURL;
          }
        },
        onError: (err) => {
          toast({
            title: __("Error"),
            description:
              err.message || __("Something went wrong. Please try again."),
            variant: "error",
          });
        },
      });
    },
    [consent, approveConsent, __, toast],
  );

  if (!consent.application || !consent.scopes) {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{__("Invalid Request")}</h1>
        <p className="text-txt-tertiary">
          {__("This consent request is invalid or has expired.")}
        </p>
      </div>
    );
  }

  if (deviceResult === "authorized") {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{__("Device Authorized")}</h1>
        <p className="text-txt-tertiary">
          {__("Your device has been successfully authorized. You can close this window and return to your device.")}
        </p>
      </div>
    );
  }

  if (deviceResult === "denied") {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{__("Access Denied")}</h1>
        <p className="text-txt-tertiary">
          {__("You have denied the authorization request. You can close this window.")}
        </p>
      </div>
    );
  }

  return (
    <div className="w-full max-w-md mx-auto pt-8 space-y-6">
      <div className="space-y-2 text-center">
        <div className="flex justify-center mb-4">
          <div className="w-12 h-12 rounded-full flex items-center justify-center bg-level-1">
            <IconLockOpen size={24} />
          </div>
        </div>
        <h1 className="text-2xl font-bold">
          {__("Authorize")}
          {" "}
          <span className="font-bold">{consent.application.name}</span>
        </h1>
        <p className="text-txt-tertiary text-sm">
          {__(
            "This application is requesting access to your account with the following permissions:",
          )}
        </p>
      </div>

      <ul className="space-y-2">
        {consent.scopes.map((scope: string) => {
          const label = scopeLabels[scope];
          if (!label) return null;
          return (
            <li
              key={scope}
              className="flex items-center gap-2.5 px-3 py-2.5 text-sm text-txt-secondary border border-border-mid rounded-lg"
            >
              {scopeIcons[scope]}
              {__(label)}
            </li>
          );
        })}
      </ul>

      <div className="flex gap-3">
        <Button
          variant="secondary"
          className="flex-1 h-10"
          disabled={isInFlight}
          onClick={() => handleAction(false)}
        >
          {__("Deny")}
        </Button>
        <Button
          className="flex-1 h-10"
          disabled={isInFlight}
          onClick={() => handleAction(true)}
        >
          {isInFlight ? __("Authorizing...") : __("Allow")}
        </Button>
      </div>

      <p className="text-center text-xs text-txt-tertiary">
        {__("You can revoke access at any time from your account settings.")}
      </p>
    </div>
  );
}
