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

import type { SubprocessorsPageQuery } from "./__generated__/SubprocessorsPageQuery.graphql";
import { toQueryVariables } from "./_lib/toQueryVariables";
import { useSubprocessorFilters } from "./_lib/useSubprocessorFilters";
import { SubprocessorsPage, subprocessorsPageQuery } from "./SubprocessorsPage";
import { SubprocessorsPageSkeleton } from "./SubprocessorsPageSkeleton";

export default function SubprocessorsPageLoader() {
  const filters = useSubprocessorFilters();
  const [queryRef, loadQuery] = useQueryLoader<SubprocessorsPageQuery>(subprocessorsPageQuery);

  // Seed the first fetch with the URL's filter values; later changes are handled
  // by the page's refetch.
  const initialVariables = useRef(toQueryVariables(filters));

  useEffect(() => {
    loadQuery(initialVariables.current);
  }, [loadQuery]);

  if (!queryRef) {
    return <SubprocessorsPageSkeleton />;
  }

  return <SubprocessorsPage queryRef={queryRef} />;
}
