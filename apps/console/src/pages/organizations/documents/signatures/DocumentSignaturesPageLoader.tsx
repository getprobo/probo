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

import type { DocumentSignaturesPageQuery } from "#/__generated__/core/DocumentSignaturesPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DocumentSignaturesPage, documentSignaturesPageQuery } from "./DocumentSignaturesPage";

function DocumentSignaturesPageQueryLoader() {
  const organizationId = useOrganizationId();
  const { documentId, versionId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery] = useQueryLoader<DocumentSignaturesPageQuery>(documentSignaturesPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({
        organizationId,
        documentId: documentId,
        versionId: versionId ?? "",
        versionSpecified: !!versionId,
      });
    }
  });

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentSignaturesPage queryRef={queryRef} />;
}

export default function DocumentSignaturesPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<LinkCardSkeleton />}>
        <DocumentSignaturesPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
