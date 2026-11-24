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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

// The document page shows a previous revision while this search param holds
// its id, and the live document otherwise.
export const documentRevisionParam = "revision";

// Links made before versions became tags of the history carry the version id
// here. The page resolves it to the revision the version tags.
export const legacyDocumentVersionParam = "version";

export function useDocumentRevisionParam() {
  const [searchParams, setSearchParams] = useSearchParams();
  const revisionId = searchParams.get(documentRevisionParam);

  const selectRevision = useCallback((next: string | null) => {
    setSearchParams((prev) => {
      const params = new URLSearchParams(prev);
      params.delete(legacyDocumentVersionParam);
      if (next == null) {
        params.delete(documentRevisionParam);
      } else {
        params.set(documentRevisionParam, next);
      }
      return params;
    });
  }, [setSearchParams]);

  return { revisionId, selectRevision };
}

// The document list keeps its filters in the search the details page was
// opened with. The revision belongs to the details page only.
export function documentListSearch(search: string): string {
  const params = new URLSearchParams(search);
  params.delete(documentRevisionParam);
  params.delete(legacyDocumentVersionParam);
  const next = params.toString();
  return next === "" ? "" : `?${next}`;
}
