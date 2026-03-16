// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
import { useParams } from "react-router";

import { RelayProvider } from "#/providers/RelayProviders";

import type { DocumentPageQuery } from "./__generated__/DocumentPageQuery.graphql";
import { DocumentPage, documentPageQuery } from "./DocumentPage";

function DocumentPageQueryLoader() {
  const { documentId } = useParams<{ documentId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<DocumentPageQuery>(documentPageQuery);

  useEffect(() => {
    if (documentId) {
      loadQuery({ id: documentId });
    }
  }, [documentId, loadQuery]);

  if (!queryRef) return null;

  return <DocumentPage queryRef={queryRef} />;
}

export default function DocumentPageLoader() {
  return (
    <RelayProvider>
      <DocumentPageQueryLoader />
    </RelayProvider>
  );
}
