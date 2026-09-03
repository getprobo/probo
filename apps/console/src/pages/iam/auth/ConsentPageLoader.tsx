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

import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Component, type ReactNode, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useQueryLoader } from "react-relay";
import { useSearchParams } from "react-router";

import type { ConsentPageQuery } from "#/__generated__/iam/ConsentPageQuery.graphql";

import ConsentPage, { consentPageQuery } from "./ConsentPage";

function ConsentPageQueryLoader() {
  const [queryRef, loadQuery]
    = useQueryLoader<ConsentPageQuery>(consentPageQuery);
  const [searchParams] = useSearchParams();
  const consentId = searchParams.get("consent_id") ?? "";

  useEffect(() => {
    loadQuery({ consentId });
  }, [loadQuery, consentId]);

  if (!queryRef) return null;

  return <ConsentPage queryRef={queryRef} />;
}

class ConsentErrorBoundary extends Component<
  { fallback: ReactNode; children: ReactNode },
  { hasError: boolean }
> {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback;
    }
    return this.props.children;
  }
}

function ConsentErrorFallback() {
  const { t } = useTranslation();

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

export default function ConsentPageLoader() {
  const [searchParams] = useSearchParams();
  const consentId = searchParams.get("consent_id") ?? "";

  return (
    <ConsentErrorBoundary key={consentId} fallback={<ConsentErrorFallback />}>
      <ConsentPageQueryLoader />
    </ConsentErrorBoundary>
  );
}
