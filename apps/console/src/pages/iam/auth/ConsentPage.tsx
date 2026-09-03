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

import {
  ArrowsClockwiseIcon,
  CaretDownIcon,
  EnvelopeSimpleIcon,
  KeyIcon,
  LockOpenIcon,
  UserCircleIcon,
  UserIcon,
} from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useToast } from "@probo/ui";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Spinner } from "@probo/ui/src/v2/Spinner/Spinner";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useMutation, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { ConsentPageMutation } from "#/__generated__/iam/ConsentPageMutation.graphql";
import type { ConsentPageQuery } from "#/__generated__/iam/ConsentPageQuery.graphql";
import { formatAPIScopeLabel } from "#/pages/iam/oauthTokens/_components/scopeLabels";

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
  openid: <UserIcon className="size-[18px] shrink-0 text-sand-11" />,
  email: <EnvelopeSimpleIcon className="size-[18px] shrink-0 text-sand-11" />,
  profile: <UserCircleIcon className="size-[18px] shrink-0 text-sand-11" />,
  offline_access: <ArrowsClockwiseIcon className="size-[18px] shrink-0 text-sand-11" />,
};

function scopeIcon(name: string): React.ReactNode {
  return scopeIcons[name] ?? <KeyIcon className="size-[18px] shrink-0 text-sand-11" />;
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

function ConsentScopeRow({
  scope,
  nested,
}: {
  scope: string;
  nested?: boolean;
}) {
  const { t } = useTranslation();
  const key = scopeLabels[scope];
  const translated = key ? t(key) : formatAPIScopeLabel(scope, t);

  return (
    <li
      className={
        nested
          ? "flex items-center gap-2.5 py-1.5 text-2 text-sand-11"
          : "flex items-center gap-2.5 rounded-3 border border-sand-6 px-3 py-2.5 text-2 text-sand-11"
      }
    >
      {scopeIcon(scope)}
      {translated}
    </li>
  );
}

function ConsentApiScopesAccordion({
  scopes,
  summaryLabel,
}: {
  scopes: readonly string[];
  summaryLabel: string;
}) {
  if (scopes.length === 0) {
    return null;
  }

  return (
    <details className="group rounded-3 border border-sand-6">
      <summary className="flex cursor-pointer list-none items-center gap-2.5 px-3 py-2.5 text-2 text-sand-11 select-none [&::-webkit-details-marker]:hidden">
        <KeyIcon className="size-[18px] shrink-0 text-sand-11" />
        <span className="min-w-0 flex-1 text-start">{summaryLabel}</span>
        <CaretDownIcon
          className="size-4 shrink-0 text-sand-11 transition-transform group-open:rotate-180"
        />
      </summary>
      <ul className="space-y-1 border-t border-sand-6 px-3 py-2.5">
        {scopes.map(scope => (
          <ConsentScopeRow
            key={scope}
            scope={scope}
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
      <div className="flex w-full flex-col gap-4">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("consentPage.invalidRequest.title")}
        </Heading>
        <Callout color="red">
          {t("consentPage.invalidRequest.description")}
        </Callout>
      </div>
    );
  }

  if (deviceResult === "authorized") {
    return (
      <div className="flex w-full flex-col gap-4">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("consentPage.deviceAuthorized.title")}
        </Heading>
        <Text size={2} align="center" className="block">
          {t("consentPage.deviceAuthorized.description")}
        </Text>
      </div>
    );
  }

  if (deviceResult === "denied") {
    return (
      <div className="flex w-full flex-col gap-4">
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("consentPage.accessDenied.title")}
        </Heading>
        <Callout color="amber">
          {t("consentPage.accessDenied.description")}
        </Callout>
      </div>
    );
  }

  if (redirectState) {
    return (
      <div className="flex w-full flex-col items-center gap-4">
        <Spinner size={3} aria-label={t("consentPage.redirectingTo")} />
        <div className="flex flex-col gap-1">
          <Heading level={1} size={4} weight="medium" align="center" highContrast>
            {redirectState.approved ? t("consentPage.authorizationComplete") : t("consentPage.accessDenied.title")}
          </Heading>
          <Text size={2} align="center" className="block">
            {t("consentPage.redirectingTo")}
            {" "}
            <Text weight="medium" highContrast>
              {consent.application.name}
            </Text>
            …
          </Text>
        </div>
      </div>
    );
  }

  return (
    <div className="flex w-full flex-col gap-8">
      <div className="flex flex-col gap-1">
        <div className="mb-2 flex justify-center">
          <div className="flex size-12 items-center justify-center rounded-full bg-sand-3 text-sand-12">
            <LockOpenIcon className="size-6" />
          </div>
        </div>
        <Heading level={1} size={4} weight="medium" align="center" highContrast>
          {t("consentPage.authorize")}
          {" "}
          {consent.application.name}
        </Heading>
        <Text align="center" size={2} className="block">
          {t("consentPage.description")}
        </Text>
      </div>

      <div className="flex flex-col gap-2">
        {oidcScopes.length > 0 && (
          <ul className="flex flex-col gap-2">
            {oidcScopes.map(scope => (
              <ConsentScopeRow
                key={scope}
                scope={scope}
              />
            ))}
          </ul>
        )}

        <ConsentApiScopesAccordion
          scopes={apiScopes}
          summaryLabel={apiScopesSummary}
        />
      </div>

      <div className="flex gap-3">
        <Button
          variant="soft"
          color="neutral"
          size={3}
          className="flex-1"
          loading={pendingAction === "deny"}
          disabled={pendingAction !== null}
          onClick={() => handleAction(false)}
        >
          {t("consentPage.actions.deny")}
        </Button>
        <Button
          variant="solid"
          color="neutral"
          highContrast
          size={3}
          className="flex-1"
          loading={pendingAction === "allow"}
          disabled={pendingAction !== null}
          onClick={() => handleAction(true)}
        >
          {t("consentPage.actions.allow")}
        </Button>
      </div>

      <Text align="center" size={1} className="block">
        {t("consentPage.revokeNotice")}
      </Text>
    </div>
  );
}
