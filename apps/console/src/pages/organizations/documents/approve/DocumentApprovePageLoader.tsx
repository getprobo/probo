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

import { Spinner } from "@probo/ui";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { DocumentApprovePageQuery } from "#/__generated__/core/DocumentApprovePageQuery.graphql";

import {
  DocumentApprovePage,
  documentApprovePageQuery,
} from "./DocumentApprovePage";

function DocumentApprovePageQueryLoader() {
  const { documentId } = useParams();
  if (!documentId) {
    throw new Error(":documentId missing in route params");
  }

  const [queryRef, loadQuery]
    = useQueryLoader<DocumentApprovePageQuery>(documentApprovePageQuery);

  useEffect(() => {
    loadQuery({ documentId });
  }, [loadQuery, documentId]);

  if (!queryRef) {
    return <Spinner />;
  }

  return (
    <Suspense fallback={<Spinner />}>
      <DocumentApprovePage queryRef={queryRef} />
    </Suspense>
  );
}

export default function DocumentApprovePageLoader() {
  return (
    <Suspense fallback={<Spinner />}>
      <DocumentApprovePageQueryLoader />
    </Suspense>
  );
}
