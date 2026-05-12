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

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { CookieBannerConsentRecordsPageQuery } from "#/__generated__/core/CookieBannerConsentRecordsPageQuery.graphql";

import CookieBannerConsentRecordsPage, { cookieBannerConsentRecordsPageQuery } from "./CookieBannerConsentRecordsPage";
import { CookieBannerConsentRecordsPageSkeleton } from "./CookieBannerConsentRecordsPageSkeleton";

export default function CookieBannerConsentRecordsPageLoader() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  if (typeof cookieBannerId !== "string") {
    throw new Error("Missing cookieBannerId parameter");
  }

  const [queryRef, loadQuery] = useQueryLoader<CookieBannerConsentRecordsPageQuery>(
    cookieBannerConsentRecordsPageQuery,
  );

  useEffect(() => {
    loadQuery({ cookieBannerId });
  }, [loadQuery, cookieBannerId]);

  if (!queryRef) {
    return <CookieBannerConsentRecordsPageSkeleton />;
  }

  return (
    <Suspense fallback={<CookieBannerConsentRecordsPageSkeleton />}>
      <CookieBannerConsentRecordsPage queryRef={queryRef} />
    </Suspense>
  );
}
