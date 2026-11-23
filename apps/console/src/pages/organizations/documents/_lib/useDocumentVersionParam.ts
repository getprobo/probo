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

// The document page shows a previous version while this search param holds
// its id, and the live document otherwise.
export const documentVersionParam = "version";

export function useDocumentVersionParam() {
  const [searchParams, setSearchParams] = useSearchParams();
  const versionId = searchParams.get(documentVersionParam);

  const selectVersion = useCallback((next: string | null) => {
    setSearchParams((prev) => {
      const params = new URLSearchParams(prev);
      if (next == null) {
        params.delete(documentVersionParam);
      } else {
        params.set(documentVersionParam, next);
      }
      return params;
    });
  }, [setSearchParams]);

  return { versionId, selectVersion };
}

// The document list keeps its filters in the search the details page was
// opened with. The version belongs to the details page only.
export function documentListSearch(search: string): string {
  const params = new URLSearchParams(search);
  params.delete(documentVersionParam);
  const next = params.toString();
  return next === "" ? "" : `?${next}`;
}
