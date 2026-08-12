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

import { useMemo } from "react";
import { matchPath, useLocation } from "react-router";

import { type NavGroup, navItemPath } from "./navigation";

/**
 * URL prefixes that put a pathname inside a product.
 *
 * A group with a segment owns everything below it, which is one pattern and
 * covers routes the nav table does not enumerate: `settings/webhooks` has no
 * entry of its own but still belongs to Settings. The two groups that carry no
 * segment are identified by their entries' paths instead.
 */
function navGroupPatterns(group: NavGroup): string[] {
  const prefixes = group.segment != null
    ? [group.segment]
    : group.items.map(item => navItemPath(group, item));

  return prefixes.map(prefix => `/organizations/:organizationId/${prefix}`);
}

/**
 * The product the current URL belongs to, or null outside the nav (a path the
 * viewer has no permission for, or one under no group at all).
 *
 * Every group owns a distinct first segment, so at most one can match and the
 * first hit is the answer — no ranking needed. `end: false` makes each pattern
 * a prefix match that still respects segment boundaries, so `data` does not
 * swallow a hypothetical `database`.
 */
export function useActiveNavGroup(groups: NavGroup[]): NavGroup | null {
  const { pathname } = useLocation();

  return useMemo(
    () =>
      groups.find(group =>
        navGroupPatterns(group).some(path => matchPath({ path, end: false }, pathname) != null),
      ) ?? null,
    [groups, pathname],
  );
}
