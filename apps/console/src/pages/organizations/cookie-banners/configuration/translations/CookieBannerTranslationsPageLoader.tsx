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

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { CookieBannerTranslationsPageQuery } from "#/__generated__/core/CookieBannerTranslationsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import CookieBannerTranslationsPage, { cookieBannerTranslationsPageQuery } from "./CookieBannerTranslationsPage";

export default function CookieBannerTranslationsPageLoader() {
  const { cookieBannerId } = useParams<{ cookieBannerId: string }>();
  if (!cookieBannerId) {
    throw new Error("missing cookieBannerId param");
  }

  const [queryRef, loadQuery] = useQueryLoader<CookieBannerTranslationsPageQuery>(cookieBannerTranslationsPageQuery);

  useEffect(() => {
    loadQuery({ cookieBannerId });
  }, [loadQuery, cookieBannerId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <CookieBannerTranslationsPage queryRef={queryRef} />
    </Suspense>
  );
}
