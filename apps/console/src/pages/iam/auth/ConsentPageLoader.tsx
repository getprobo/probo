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

import { useTranslate } from "@probo/i18n";
import { Component, type ReactNode, useEffect } from "react";
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
  const { __ } = useTranslate();

  return (
    <div className="w-full max-w-md mx-auto pt-8 space-y-6 text-center">
      <h1 className="text-2xl font-bold">{__("Invalid Request")}</h1>
      <p className="text-txt-tertiary">
        {__("This consent request is invalid or has expired.")}
      </p>
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
