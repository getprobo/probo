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

import { Layout, Skeleton } from "@probo/ui";
import { Suspense } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Outlet } from "react-router";

import type { ViewerLayoutQuery } from "#/__generated__/iam/ViewerLayoutQuery.graphql";

import { ViewerDropdown } from "./_components/ViewerDropdown";

export const viewerLayoutQuery = graphql`
  query ViewerLayoutQuery {
    viewer @required(action: THROW) {
      ...ViewerDropdownFragment
    }
  }
`;

export function ViewerLayout(props: {
  hideSidebar?: boolean;
  queryRef: PreloadedQuery<ViewerLayoutQuery>;
}) {
  const { queryRef } = props;

  const { viewer } = usePreloadedQuery<ViewerLayoutQuery>(
    viewerLayoutQuery,
    queryRef,
  );

  return (
    <Layout
      headerTrailing={(
        <div className="ml-auto">
          <Suspense fallback={<Skeleton className="w-32 h-8" />}>
            <ViewerDropdown fKey={viewer} />
          </Suspense>
        </div>
      )}
    >
      <Outlet />
    </Layout>
  );
}
