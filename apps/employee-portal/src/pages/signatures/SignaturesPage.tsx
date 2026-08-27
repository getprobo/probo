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

import type { SignaturesPageQuery } from "#/__generated__/core/SignaturesPageQuery.graphql";
import { PageHeader } from "#/pages/_components/PageHeader";

import { SignaturesHistoryList } from "./_components/SignaturesHistoryList";
import { SignaturesPendingList } from "./_components/SignaturesPendingList";

export const signaturesPageQuery = graphql`
  query SignaturesPageQuery($organizationId: ID!, $first: Int) @throwOnFieldError {
    viewer @required(action: THROW) {
      ...SignaturesPendingList_viewer @arguments(organizationId: $organizationId, first: $first)
      ...SignaturesHistoryList_viewer @arguments(organizationId: $organizationId, first: $first)
    }
  }
`;

interface SignaturesPageProps {
  queryRef: PreloadedQuery<SignaturesPageQuery>;
}

export function SignaturesPage({ queryRef }: SignaturesPageProps) {
  const { t } = useTranslation("signatures");
  const { t: tApp } = useTranslation();
  const { viewer } = usePreloadedQuery<SignaturesPageQuery>(
    signaturesPageQuery,
    queryRef,
  );

  return (
    <main className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-8 pt-8 pb-32">
      <PageHeader
        homeLabel={tApp("homePage.breadcrumb")}
        currentLabel={t("breadcrumb")}
        title={t("title")}
      />
      <SignaturesPendingList viewerKey={viewer} />
      <SignaturesHistoryList viewerKey={viewer} />
    </main>
  );
}
