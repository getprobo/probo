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

import { Suspense, useCallback, useEffect, useState } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentLayoutQuery } from "#/__generated__/core/DocumentLayoutQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentLayout, documentLayoutQuery } from "./DocumentLayout";

function DocumentLayoutQueryLoader() {
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }
  const [queryRef, loadQuery] = useQueryLoader<DocumentLayoutQuery>(documentLayoutQuery);

  // Detect param changes (e.g. navigating from versioned to versionless URL)
  // and refetch without remounting the component tree.
  const paramsKey = `${documentId}-${versionId}`;
  const [prevParamsKey, setPrevParamsKey] = useState(paramsKey);
  if (queryRef && paramsKey !== prevParamsKey) {
    setPrevParamsKey(paramsKey);
    loadQuery(
      { documentId, versionId: versionId ?? "", versionSpecified: !!versionId },
      { fetchPolicy: "store-and-network" },
    );
  }

  useEffect(() => {
    if (!queryRef) {
      loadQuery(
        {
          documentId,
          versionId: versionId ?? "",
          versionSpecified: !!versionId,
        },
        { fetchPolicy: "store-and-network" },
      );
    }
  });

  const onRefetch = useCallback(() => {
    loadQuery(
      { documentId, versionId: versionId ?? "", versionSpecified: !!versionId },
      { fetchPolicy: "store-and-network" },
    );
  }, [documentId, versionId, loadQuery]);

  if (!queryRef) return <PageSkeleton />;

  return <DocumentLayout queryRef={queryRef} onRefetch={onRefetch} />;
}

export default function DocumentLayoutLoader() {
  const { documentId } = useParams();

  return (
    <CoreRelayProvider>
      <Suspense key={documentId} fallback={<PageSkeleton />}>
        <DocumentLayoutQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
