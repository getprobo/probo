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
import { Button, useToast } from "@probo/ui";
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

function ScopeIcon({ scope }: { scope: string }) {
  switch (scope) {
    case "openid":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth="1.5"
          stroke="currentColor"
          className="w-[18px] h-[18px] shrink-0 text-txt-tertiary"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
          />
        </svg>
      );
    case "email":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth="1.5"
          stroke="currentColor"
          className="w-[18px] h-[18px] shrink-0 text-txt-tertiary"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75"
          />
        </svg>
      );
    case "profile":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth="1.5"
          stroke="currentColor"
          className="w-[18px] h-[18px] shrink-0 text-txt-tertiary"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M17.982 18.725A7.488 7.488 0 0 0 12 15.75a7.488 7.488 0 0 0-5.982 2.975m11.963 0a9 9 0 1 0-11.963 0m11.963 0A8.966 8.966 0 0 1 12 21a8.966 8.966 0 0 1-5.982-2.275M15 9.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
          />
        </svg>
      );
    case "offline_access":
      return (
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth="1.5"
          stroke="currentColor"
          className="w-[18px] h-[18px] shrink-0 text-txt-tertiary"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182M21.015 4.356v4.992"
          />
        </svg>
      );
    default:
      return null;
  }
}

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
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth="1.5"
              stroke="currentColor"
              className="w-6 h-6"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M13.5 10.5V6.75a4.5 4.5 0 1 1 9 0v3.75M3.75 21.75h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H3.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z"
              />
            </svg>
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
              <ScopeIcon scope={scope} />
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
