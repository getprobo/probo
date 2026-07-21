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
import { useTranslation } from "react-i18next";
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
  openid: "consentPage.scopes.openid",
  email: "consentPage.scopes.email",
  profile: "consentPage.scopes.profile",
  offline_access: "consentPage.scopes.offlineAccess",
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

function scopeLabel(name: string, translate: (key: string) => string): string {
  const key = scopeLabels[name];
  return key ? translate(key) : formatApiScopeLabel(name);
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
  const translated = scopeLabel(props.scope, props.translate);

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
  const { t } = useTranslation();
  const { toast } = useToast();
  const [deviceResult, setDeviceResult] = useState<"authorized" | "denied" | null>(null);
  const [pendingAction, setPendingAction] = useState<"allow" | "deny" | null>(null);
  const [redirectState, setRedirectState] = useState<{
    url: string;
    approved: boolean;
  } | null>(null);

  const data = usePreloadedQuery<ConsentPageQuery>(consentPageQuery, props.queryRef);
  usePageTitle(t("consentPage.pageTitle"));

  const { node: consent } = data;

  const [approveConsent] = useMutation<ConsentPageMutation>(approveConsentMutation);

  const { oidcScopes, apiScopes } = useMemo(
    () => partitionScopes(consent.scopes ?? []),
    [consent.scopes],
  );

  const apiScopesSummary = useMemo(
    () => t("consentPage.apiAccess", { count: apiScopes.length }),
    [t, apiScopes.length],
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
              title: t("consentPage.errors.authorizationFailed"),
              description: formatError(
                t("consentPage.errors.generic"),
                errors,
              ),
              variant: "error",
            });
            return;
          }

          if (!response.approveConsent) {
            setPendingAction(null);
            toast({
              title: t("consentPage.errors.authorizationFailed"),
              description: t("consentPage.errors.generic"),
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
            title: t("common.error"),
            description:
              err.message || t("consentPage.errors.generic"),
            variant: "error",
          });
        },
      });
    },
    [consent, approveConsent, t, toast, pendingAction],
  );

  if (!consent.application || !consent.scopes) {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{t("consentPage.invalidRequest.title")}</h1>
        <p className="text-txt-tertiary">
          {t("consentPage.invalidRequest.description")}
        </p>
      </div>
    );
  }

  if (deviceResult === "authorized") {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{t("consentPage.deviceAuthorized.title")}</h1>
        <p className="text-txt-tertiary">
          {t("consentPage.deviceAuthorized.description")}
        </p>
      </div>
    );
  }

  if (deviceResult === "denied") {
    return (
      <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
        <h1 className="text-2xl font-bold">{t("consentPage.accessDenied.title")}</h1>
        <p className="text-txt-tertiary">
          {t("consentPage.accessDenied.description")}
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
            {redirectState.approved ? t("consentPage.authorizationComplete") : t("consentPage.accessDenied.title")}
          </h1>
          <p className="text-txt-tertiary">
            {t("consentPage.redirectingTo")}
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
          {t("consentPage.authorize")}
          {" "}
          <span className="font-bold">{consent.application.name}</span>
        </h1>
        <p className="text-txt-tertiary text-sm">
          {t("consentPage.description")}
        </p>
      </div>

      <div className="space-y-2">
        {oidcScopes.length > 0 && (
          <ul className="space-y-2">
            {oidcScopes.map(scope => (
              <ConsentScopeRow
                key={scope}
                scope={scope}
                translate={t}
              />
            ))}
          </ul>
        )}

        <ConsentApiScopesAccordion
          scopes={apiScopes}
          translate={t}
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
          {t("consentPage.actions.deny")}
        </Button>
        <Button
          className="flex-1 h-10"
          disabled={pendingAction !== null}
          icon={pendingAction === "allow" ? Spinner : undefined}
          onClick={() => handleAction(true)}
        >
          {t("consentPage.actions.allow")}
        </Button>
      </div>

      <p className="text-center text-xs text-txt-tertiary">
        {t("consentPage.revokeNotice")}
      </p>
    </div>
  );
}
