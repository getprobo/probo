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

  const { draftDeletedAt } = useOutletContext<{ draftDeletedAt: number }>();
  const [queryRef, loadQuery] = useQueryLoader<DocumentDescriptionPageQuery>(documentDescriptionPageQuery);

  useEffect(() => {
    if (!queryRef) {
      loadQuery({
        documentId: documentId,
        versionId: versionId ?? "",
        versionSpecified: !!versionId,
      });
    }
  });

  useEffect(() => {
    if (draftDeletedAt > 0) {
      loadQuery(
        { documentId, versionId: versionId ?? "", versionSpecified: !!versionId },
        { fetchPolicy: "network-only" },
      );
    }
  }, [draftDeletedAt, loadQuery, documentId, versionId]);

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return <DocumentDescriptionPage queryRef={queryRef} />;
}

export default function DocumentDescriptionPageLoader() {
  return (
    <CoreRelayProvider>
      <DocumentDescriptionPageQueryLoader />
    </CoreRelayProvider>
  );
}
