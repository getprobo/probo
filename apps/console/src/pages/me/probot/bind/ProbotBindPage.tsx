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

import { usePageTitle } from "@probo/hooks";
import { Button, Card } from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { Link, useNavigate } from "react-router";
import { graphql } from "relay-runtime";

import type { ProbotBindPageConfirmMutation } from "#/__generated__/core/ProbotBindPageConfirmMutation.graphql";
import type { ProbotBindPageQuery } from "#/__generated__/core/ProbotBindPageQuery.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const probotBindPageQuery = graphql`
  query ProbotBindPageQuery($token: String!) {
    probotIdentityBindPreview(token: $token) {
      provider
      externalTenantId
      externalUserId
      externalTenantName
      externalUserName
    }
  }
`;

const confirmProbotIdentityBindingMutation = graphql`
  mutation ProbotBindPageConfirmMutation($input: ConfirmProbotIdentityBindingInput!) {
    confirmProbotIdentityBinding(input: $input) {
      probotIdentityBinding {
        id
        provider
        externalTenantId
        externalUserId
      }
    }
  }
`;

function bindPreviewValue(name: string, id: string): string {
  if (name !== "") {
    return name;
  }

  return id;
}

interface ProbotBindPageProps {
  queryRef: PreloadedQuery<ProbotBindPageQuery>;
  token: string;
}

export function ProbotBindPage({ queryRef, token }: ProbotBindPageProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [confirmed, setConfirmed] = useState(false);

  usePageTitle(t("probotBindPage.title"));

  const data = usePreloadedQuery<ProbotBindPageQuery>(
    probotBindPageQuery,
    queryRef,
  );
  const [confirmBinding, isConfirming]
    = useMutation<ProbotBindPageConfirmMutation>(
      confirmProbotIdentityBindingMutation,
      {
        successMessage: t("probotBindPage.messages.linked"),
        errorToast: t("probotBindPage.messages.linkFailed"),
      },
    );

  const preview = data.probotIdentityBindPreview;
  const tenant = bindPreviewValue(
    preview.externalTenantName,
    preview.externalTenantId,
  );
  const account = bindPreviewValue(
    preview.externalUserName,
    preview.externalUserId,
  );

  const handleConfirm = async () => {
    await confirmBinding({ variables: { input: { token } } });
    setConfirmed(true);
  };

  return (
    <div className="space-y-6 w-full max-w-md mx-auto pt-8">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-bold">{t("probotBindPage.title")}</h1>
        <p className="text-txt-tertiary">
          {t("probotBindPage.description")}
        </p>
      </div>

      <Card className="p-6 space-y-4">
        <dl className="space-y-3 text-sm">
          <div className="flex justify-between gap-4">
            <dt className="text-txt-tertiary">
              {t("probotBindPage.fields.provider")}
            </dt>
            <dd className="text-right break-all">
              {t(`probotBindPage.providers.${preview.provider}`, {
                defaultValue: preview.provider,
              })}
            </dd>
          </div>
          {tenant !== "" && (
            <div className="flex justify-between gap-4">
              <dt className="text-txt-tertiary">
                {t("probotBindPage.fields.tenant")}
              </dt>
              <dd
                className={
                  preview.externalTenantName === ""
                    ? "font-mono text-right break-all"
                    : "text-right break-all"
                }
              >
                {tenant}
              </dd>
            </div>
          )}
          <div className="flex justify-between gap-4">
            <dt className="text-txt-tertiary">
              {t("probotBindPage.fields.account")}
            </dt>
            <dd
              className={
                preview.externalUserName === ""
                  ? "font-mono text-right break-all"
                  : "text-right break-all"
              }
            >
              {account}
            </dd>
          </div>
        </dl>

        {confirmed
          ? (
              <p className="text-txt-secondary text-sm">
                {t("probotBindPage.confirmed")}
              </p>
            )
          : (
              <Button
                className="w-full"
                disabled={isConfirming}
                onClick={() => {
                  void handleConfirm();
                }}
              >
                {t("probotBindPage.actions.confirm")}
              </Button>
            )}
      </Card>

      <div className="text-center text-sm text-txt-secondary">
        <Link
          to="/"
          className="underline hover:text-txt-primary"
          onClick={(event) => {
            event.preventDefault();
            void navigate("/");
          }}
        >
          {t("probotBindPage.actions.back")}
        </Link>
      </div>
    </div>
  );
}

export function ProbotBindMissingToken() {
  const { t } = useTranslation();

  usePageTitle(t("probotBindPage.title"));

  return (
    <div className="space-y-4 w-full max-w-md mx-auto pt-8 text-center">
      <h1 className="text-3xl font-bold">{t("probotBindPage.title")}</h1>
      <p className="text-txt-tertiary">
        {t("probotBindPage.errors.missingToken")}
      </p>
    </div>
  );
}

export function ProbotBindPageFallback() {
  const { t } = useTranslation();

  return (
    <div className="space-y-4 w-full max-w-md mx-auto pt-8 text-center">
      <h1 className="text-3xl font-bold">{t("probotBindPage.title")}</h1>
      <p className="text-txt-tertiary">{t("probotBindPage.loading")}</p>
    </div>
  );
}

export function ProbotBindTokenErrorFallback() {
  const { t } = useTranslation();

  return (
    <div className="space-y-4 w-full max-w-md mx-auto pt-8 text-center">
      <h1 className="text-3xl font-bold">{t("probotBindPage.title")}</h1>
      <p className="text-txt-tertiary">
        {t("probotBindPage.errors.invalidOrExpiredToken")}
      </p>
    </div>
  );
}
