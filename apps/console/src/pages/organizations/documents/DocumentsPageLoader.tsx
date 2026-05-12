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

import type { DocumentsPageQuery } from "#/__generated__/core/DocumentsPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import DocumentsPage, { documentsPageQuery } from "./DocumentsPage";

function DocumentsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<DocumentsPageQuery>(documentsPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ organizationId });
    }
  });

  if (!queryRef) return <PageSkeleton />;

  return <DocumentsPage queryRef={queryRef} />;
}

export default function DocumentsPageLoader() {
  return (
    <CoreRelayProvider>
      <Suspense fallback={<PageSkeleton />}>
        <DocumentsPageQueryLoader />
      </Suspense>
    </CoreRelayProvider>
  );
}
