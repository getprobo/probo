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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { ViewerLayoutQuery } from "#/__generated__/iam/ViewerLayoutQuery.graphql";

import { TopBarUserMenuSkeleton } from "./_components/TopBar/TopBarUserMenuSkeleton";
import { topBar } from "./_components/TopBar/variants";
import { ViewerDropdown } from "./_components/ViewerDropdown";

export const viewerLayoutQuery = graphql`
  query ViewerLayoutQuery @throwOnFieldError {
    viewer @required(action: THROW) {
      ...ViewerDropdownFragment
    }
  }
`;

interface ViewerLayoutProps {
  queryRef: PreloadedQuery<ViewerLayoutQuery>;
}

export function ViewerLayout({ queryRef }: ViewerLayoutProps) {
  const { t } = useTranslation();
  const { viewer } = usePreloadedQuery<ViewerLayoutQuery>(
    viewerLayoutQuery,
    queryRef,
  );
  const slots = topBar();
  const tagline = t("topBar.tagline");

  return (
    <div className="flex min-h-dvh flex-col bg-sand-2">
      <header className={slots.bar()}>
        <div className={slots.inner()}>
          <span className={slots.brand()}>
            <span className={slots.brandText()}>
              <Text
                size={2}
                weight="medium"
                color="neutral"
                highContrast
                className={slots.brandName()}
              >
                {tagline}
              </Text>
            </span>
          </span>
          <Suspense fallback={<TopBarUserMenuSkeleton />}>
            <ViewerDropdown identityKey={viewer} />
          </Suspense>
        </div>
      </header>
      <Outlet />
    </div>
  );
}
