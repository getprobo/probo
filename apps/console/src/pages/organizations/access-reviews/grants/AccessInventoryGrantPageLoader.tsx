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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { AccessInventoryGrantPageQuery } from "#/__generated__/core/AccessInventoryGrantPageQuery.graphql";
import { PageError } from "#/components/PageError";
import { isNotFoundError } from "#/lib/relay/errors";

import { ACCESS_INVENTORY_GRANT_CHANGE_EVENTS_PAGE_SIZE } from "./_lib/pageSize";
import {
  AccessInventoryGrantPage,
  accessInventoryGrantPageQuery,
} from "./AccessInventoryGrantPage";
import { AccessInventoryGrantPageError } from "./AccessInventoryGrantPageError";
import { AccessInventoryGrantPageSkeleton } from "./AccessInventoryGrantPageSkeleton";

function AccessInventoryGrantPageQueryLoader() {
  const { grantId } = useParams();
  if (grantId == null || grantId === "") {
    throw new Error(":grantId missing in route params");
  }

  const [queryRef, loadQuery]
    = useQueryLoader<AccessInventoryGrantPageQuery>(
      accessInventoryGrantPageQuery,
    );

  useEffect(() => {
    loadQuery({
      grantId,
      first: ACCESS_INVENTORY_GRANT_CHANGE_EVENTS_PAGE_SIZE,
    });
  }, [grantId, loadQuery]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.grantId === grantId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <AccessInventoryGrantPageSkeleton />;
  }

  return (
    <Suspense fallback={<AccessInventoryGrantPageSkeleton />}>
      <AccessInventoryGrantPage
        key={grantId}
        queryRef={currentQueryRef}
      />
    </Suspense>
  );
}

export default function AccessInventoryGrantPageLoader() {
  const { grantId } = useParams();

  return (
    <ErrorBoundary
      key={grantId}
      fallback={(error, reset) => {
        if (isNotFoundError(error)) {
          return (
            <PageError
              error={error instanceof Error ? error : new Error("not found")}
              resetErrorBoundary={reset}
            />
          );
        }

        return <AccessInventoryGrantPageError onRetry={reset} />;
      }}
    >
      <AccessInventoryGrantPageQueryLoader />
    </ErrorBoundary>
  );
}
