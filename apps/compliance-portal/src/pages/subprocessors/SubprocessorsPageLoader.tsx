// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
