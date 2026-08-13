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
import { Button, Card, Slack } from "@probo/ui";
import type { PropsWithChildren } from "react";
import { useTranslation } from "react-i18next";
import {
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { EmployeeBindPageConfirmMutation } from "#/__generated__/core/EmployeeBindPageConfirmMutation.graphql";
import type { EmployeeBindPageQuery } from "#/__generated__/core/EmployeeBindPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

export const employeeBindPageQuery = graphql`
  query EmployeeBindPageQuery($token: String!) {
    probotIdentityBindPreview(token: $token) {
      externalTenantId
      externalUserId
      externalTenantName
      externalUserName
    }
  }
`;

const confirmProbotIdentityBindingMutation = graphql`
  mutation EmployeeBindPageConfirmMutation(
    $input: ConfirmProbotIdentityBindingInput!
  ) {
    confirmProbotIdentityBinding(input: $input) {
      probotIdentityBinding {
        id
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

function EmployeeBindShell({ children }: PropsWithChildren) {
  return (
    <div className="min-h-screen bg-level-0 text-txt-primary grid place-items-center px-4 py-12">
      <div className="w-full max-w-md space-y-6">{children}</div>
    </div>
  );
}

function EmployeeBindBrand() {
  return (
    <div className="flex justify-center">
      <div className="h-14 w-14 flex items-center justify-center rounded-2xl bg-subtle">
        <Slack className="h-8 w-8" />
      </div>
    </div>
  );
}

interface EmployeeBindPageProps {
  queryRef: PreloadedQuery<EmployeeBindPageQuery>;
  token: string;
}

export function EmployeeBindPage({ queryRef, token }: EmployeeBindPageProps) {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const organizationId = useOrganizationId();

  usePageTitle(t("employeeBindPage.title"));

  const data = usePreloadedQuery<EmployeeBindPageQuery>(
    employeeBindPageQuery,
    queryRef,
  );
  const [confirmBinding, isConfirming]
    = useMutation<EmployeeBindPageConfirmMutation>(
      confirmProbotIdentityBindingMutation,
      {
        successMessage: t("employeeBindPage.messages.linked"),
        errorToast: t("employeeBindPage.errors.linkFailed"),
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
  const tenantIsId = preview.externalTenantName === "";
  const accountIsId = preview.externalUserName === "";

  const handleConfirm = async () => {
    await confirmBinding({ variables: { input: { token } } });
    void navigate(
      `/organizations/${encodeURIComponent(organizationId)}/employee/bindings`,
      { replace: true },
    );
  };

  return (
    <EmployeeBindShell>
      <EmployeeBindBrand />
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold">
          {t("employeeBindPage.title")}
        </h1>
        <p className="text-sm text-txt-tertiary">
          {t("employeeBindPage.description")}
        </p>
      </div>

      <Card className="p-6 space-y-6">
        <dl className="space-y-4">
          {tenant !== "" && (
            <div className="space-y-1">
              <dt className="text-xs font-medium uppercase tracking-wide text-txt-tertiary">
                {t("employeeBindPage.fields.tenant")}
              </dt>
              <dd
                className={
                  tenantIsId
                    ? "text-base font-mono break-all"
                    : "text-base font-medium break-all"
                }
              >
                {tenant}
              </dd>
            </div>
          )}
          <div className="space-y-1">
            <dt className="text-xs font-medium uppercase tracking-wide text-txt-tertiary">
              {t("employeeBindPage.fields.account")}
            </dt>
            <dd
              className={
                accountIsId
                  ? "text-base font-mono break-all"
                  : "text-base font-medium break-all"
              }
            >
              {account}
            </dd>
          </div>
        </dl>

        <Button
          className="w-full"
          disabled={isConfirming}
          onClick={() => {
            void handleConfirm();
          }}
        >
          {isConfirming
            ? t("employeeBindPage.actions.linking")
            : t("employeeBindPage.actions.confirm")}
        </Button>
      </Card>
    </EmployeeBindShell>
  );
}

export function EmployeeBindMissingToken() {
  const { t } = useTranslation();

  usePageTitle(t("employeeBindPage.title"));

  return (
    <EmployeeBindShell>
      <EmployeeBindBrand />
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold">
          {t("employeeBindPage.missingToken.title")}
        </h1>
        <p className="text-sm text-txt-tertiary">
          {t("employeeBindPage.missingToken.description")}
        </p>
      </div>
    </EmployeeBindShell>
  );
}

export function EmployeeBindInvalidToken() {
  const { t } = useTranslation();

  usePageTitle(t("employeeBindPage.title"));

  return (
    <EmployeeBindShell>
      <EmployeeBindBrand />
      <div className="space-y-2 text-center">
        <h1 className="text-2xl font-semibold">
          {t("employeeBindPage.errors.invalidOrExpiredTokenTitle")}
        </h1>
        <p className="text-sm text-txt-tertiary">
          {t("employeeBindPage.errors.invalidOrExpiredToken")}
        </p>
      </div>
    </EmployeeBindShell>
  );
}

export function EmployeeBindLoading() {
  const { t } = useTranslation();

  usePageTitle(t("employeeBindPage.title"));

  return (
    <EmployeeBindShell>
      <EmployeeBindBrand />
      <p className="text-sm text-center text-txt-tertiary">
        {t("employeeBindPage.loading")}
      </p>
    </EmployeeBindShell>
  );
}
