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

import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useOutletContext, useParams } from "react-router";

import type { DocumentDescriptionPageQuery } from "#/__generated__/core/DocumentDescriptionPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentDescriptionPage, documentDescriptionPageQuery } from "./DocumentDescriptionPage";

function DocumentDescriptionPageQueryLoader() {
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const { versionChangedAt } = useOutletContext<{ versionChangedAt: number }>();
  const [queryRef, loadQuery] = useQueryLoader<DocumentDescriptionPageQuery>(documentDescriptionPageQuery);

  useEffect(() => {
    loadQuery(
      { documentId, versionId: versionId ?? "", versionSpecified: !!versionId },
      { fetchPolicy: "network-only" },
    );
  }, [documentId, versionId, versionChangedAt, loadQuery]);

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return (
    <DocumentDescriptionPage
      queryRef={queryRef}
      versionChangedAt={versionChangedAt}
    />
  );
}

export default function DocumentDescriptionPageLoader() {
  return (
    <CoreRelayProvider>
      <DocumentDescriptionPageQueryLoader />
    </CoreRelayProvider>
  );
}
