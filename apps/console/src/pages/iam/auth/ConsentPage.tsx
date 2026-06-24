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
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  IconArrowsClockwise,
  IconChevronDown,
  IconEnvelope,
  IconKey,
  IconLockOpen,
  IconUser,
  IconUserCircle,
  Spinner,
  useToast,
} from "@probo/ui";
import { useCallback, useEffect, useMemo, useState } from "react";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentPageMutation } from "#/__generated__/iam/ConsentPageMutation.graphql";
import type { ConsentPageQuery } from "#/__generated__/iam/ConsentPageQuery.graphql";
import { formatApiScopeLabel } from "#/pages/iam/oauthTokens/_components/scopeLabels";

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

function scopeIcon(name: string): React.ReactNode {
  return scopeIcons[name] ?? <IconKey size={18} className="shrink-0 text-txt-tertiary" />;
}

function scopeLabel(name: string): string {
  return scopeLabels[name] ?? formatApiScopeLabel(name);
}

function isApiScope(scope: string): boolean {
  return scope.startsWith("v1:");
}

function partitionScopes(scopes: readonly string[]) {
  const oidcScopes: string[] = [];
  const apiScopes: string[] = [];

  for (const scope of scopes) {
    if (isApiScope(scope)) {
      apiScopes.push(scope);
    } else {
      oidcScopes.push(scope);
    }
  }

  return { oidcScopes, apiScopes };
}

function ConsentScopeRow(props: {
  scope: string;
  translate: (label: string) => string;
  nested?: boolean;
}) {
  const label = scopeLabel(props.scope);
  const translated = label !== props.scope ? props.translate(label) : label;

  return (
    <li
      className={
        props.nested
          ? "flex items-center gap-2.5 py-1.5 text-sm text-txt-secondary"
          : "flex items-center gap-2.5 px-3 py-2.5 text-sm text-txt-secondary border border-border-mid rounded-lg"
      }
    >
      {scopeIcon(props.scope)}
      {translated}
    </li>
  );
}

function ConsentApiScopesAccordion(props: {
  scopes: readonly string[];
  translate: (label: string) => string;
  summaryLabel: string;
}) {
  if (props.scopes.length === 0) {
    return null;
  }

  return (
    <details className="group border border-border-mid rounded-lg">
      <summary className="flex cursor-pointer list-none items-center gap-2.5 px-3 py-2.5 text-sm text-txt-secondary select-none [&::-webkit-details-marker]:hidden">
        <IconKey size={18} className="shrink-0 text-txt-tertiary" />
        <span className="min-w-0 flex-1 text-start">{props.summaryLabel}</span>
        <IconChevronDown
          size={16}
          className="shrink-0 text-txt-tertiary transition-transform group-open:rotate-180"
        />
      </summary>
      <ul className="space-y-1 border-t border-border-mid px-3 py-2.5">
        {props.scopes.map(scope => (
          <ConsentScopeRow
            key={scope}
            scope={scope}
            translate={props.translate}
            nested
          />
        ))}
      </ul>
    </details>
  );
}

export default function ConsentPage(props: {
  queryRef: PreloadedQuery<ConsentPageQuery>;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [deviceResult, setDeviceResult] = useState<"authorized" | "denied" | null>(null);
  const [pendingAction, setPendingAction] = useState<"allow" | "deny" | null>(null);
  const [redirectState, setRedirectState] = useState<{
    url: string;
    approved: boolean;
  } | null>(null);

  const data = usePreloadedQuery<ConsentPageQuery>(consentPageQuery, props.queryRef);
  usePageTitle(__("Authorize Application"));

  const { node: consent } = data;

  const [approveConsent] = useMutation<ConsentPageMutation>(approveConsentMutation);

  const { oidcScopes, apiScopes } = useMemo(
    () => partitionScopes(consent.scopes ?? []),
    [consent.scopes],
  );

  const apiScopesSummary = useMemo(
    () => `${__("API access")} (${apiScopes.length})`,
    [__, apiScopes.length],
  );

  useEffect(() => {
    if (!redirectState) return;

    window.location.href = redirectState.url;
  }, [redirectState]);

  const handleAction = useCallback(
    (approved: boolean) => {
      if (!consent.id || pendingAction !== null) return;

      setPendingAction(approved ? "allow" : "deny");

      approveConsent({
        variables: {
          input: {
            consentId: consent.id,
            approved,
          },
        },
        onCompleted: (response, errors) => {
          if (errors) {
            setPendingAction(null);
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
            setPendingAction(null);
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
            setRedirectState({
              url: response.approveConsent.redirectURL,
              approved,
            });
          }
        },
        onError: (err) => {
          setPendingAction(null);
          toast({
            title: __("Error"),
            description:
              err.message || __("Something went wrong. Please try again."),
            variant: "error",
          });
        },
      });
    },
    [consent, approveConsent, __, toast, pendingAction],
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

  if (redirectState) {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <Spinner size={24} centered className="text-txt-tertiary" />
        <div className="space-y-2">
          <h1 className="text-2xl font-bold">
            {redirectState.approved ? __("Authorization Complete") : __("Access Denied")}
          </h1>
          <p className="text-txt-tertiary">
            {__("You will be redirected to")}
            {" "}
            <span className="font-medium text-txt-secondary">
              {consent.application.name}
            </span>
            …
          </p>
        </div>
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

      <div className="space-y-2">
        {oidcScopes.length > 0 && (
          <ul className="space-y-2">
            {oidcScopes.map(scope => (
              <ConsentScopeRow
                key={scope}
                scope={scope}
                translate={__}
              />
            ))}
          </ul>
        )}

        <ConsentApiScopesAccordion
          scopes={apiScopes}
          translate={__}
          summaryLabel={apiScopesSummary}
        />
      </div>

      <div className="flex gap-3">
        <Button
          variant="secondary"
          className="flex-1 h-10"
          disabled={pendingAction !== null}
          icon={pendingAction === "deny" ? Spinner : undefined}
          onClick={() => handleAction(false)}
        >
          {__("Deny")}
        </Button>
        <Button
          className="flex-1 h-10"
          disabled={pendingAction !== null}
          icon={pendingAction === "allow" ? Spinner : undefined}
          onClick={() => handleAction(true)}
        >
          {__("Allow")}
        </Button>
      </div>

      <p className="text-center text-xs text-txt-tertiary">
        {__("You can revoke access at any time from your account settings.")}
      </p>
    </div>
  );
}
