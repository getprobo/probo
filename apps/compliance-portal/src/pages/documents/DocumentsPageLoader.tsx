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

import { useEffect, useRef } from "react";
import { useQueryLoader } from "react-relay";

import type { DocumentsPageQuery } from "./__generated__/DocumentsPageQuery.graphql";
import { toQueryVariables } from "./_lib/toQueryVariables";
import { useDocumentTab } from "./_lib/useDocumentTab";
import { DocumentsPage, documentsPageQuery } from "./DocumentsPage";
import { DocumentsPageSkeleton } from "./DocumentsPageSkeleton";

export default function DocumentsPageLoader() {
  const { tab } = useDocumentTab();
  const [queryRef, loadQuery] = useQueryLoader<DocumentsPageQuery>(documentsPageQuery);

  // Seed the first fetch with the URL's tab; later changes are handled by the
  // page's refetch.
  const initialVariables = useRef(toQueryVariables(tab));

  useEffect(() => {
    loadQuery(initialVariables.current);
  }, [loadQuery]);

  if (!queryRef) {
    return <DocumentsPageSkeleton />;
  }

  return <DocumentsPage queryRef={queryRef} />;
}
