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

import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { useNavigate, useParams } from "react-router";

import type { BindPageQuery } from "#/__generated__/core/BindPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";
import { PageHeader } from "#/pages/_components/PageHeader";

import { BindConfirmCard } from "./_components/BindConfirmCard";
import { bindingsPage } from "./_components/variants";
import { useConfirmIdentityBinding } from "./_lib/useConfirmIdentityBinding";

export const bindPageQuery = graphql`
  query BindPageQuery($token: String!) @throwOnFieldError {
    preview: probotIdentityBindPreview(token: $token) {
      externalTenantId
      externalUserId
      externalTenantName
      externalUserName
    }
  }
`;

interface BindPageProps {
  queryRef: PreloadedQuery<BindPageQuery>;
  token: string;
}

export function BindPage({ queryRef, token }: BindPageProps) {
  const { t } = useTranslation("bindings");
  const { t: tApp } = useTranslation();
  const navigate = useNavigate();
  const { organizationId } = useParams();
  const slots = bindingsPage();
  const { preview } = usePreloadedQuery<BindPageQuery>(bindPageQuery, queryRef);
  const [confirmBinding, isConfirming] = useConfirmIdentityBinding();

  if (organizationId === undefined) {
    throw new NotFoundError("organizationId is required");
  }

  async function handleConfirm() {
    await confirmBinding({ variables: { input: { token } } });
    void navigate(`/${organizationId}/bindings`, { replace: true });
  }

  return (
    <main className={slots.main()}>
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        parent={{
          label: t("list.breadcrumb"),
          to: `/${organizationId}/bindings`,
        }}
        currentLabel={t("bind.breadcrumb")}
        title={t("bind.title")}
      />
      <BindConfirmCard
        externalTenantId={preview.externalTenantId}
        externalUserId={preview.externalUserId}
        externalTenantName={preview.externalTenantName}
        externalUserName={preview.externalUserName}
        isConfirming={isConfirming}
        onConfirm={() => {
          void handleConfirm();
        }}
      />
    </main>
  );
}
