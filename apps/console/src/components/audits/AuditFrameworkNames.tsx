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
import { graphql, usePaginationFragment } from "react-relay";

import type { AuditFrameworkNames_audit$key } from "#/__generated__/core/AuditFrameworkNames_audit.graphql";
import type { AuditFrameworkNamesPaginationQuery } from "#/__generated__/core/AuditFrameworkNamesPaginationQuery.graphql";

const PAGE_SIZE = 100;

const auditFrameworkNamesFragment = graphql`
  fragment AuditFrameworkNames_audit on Audit
  @refetchable(queryName: "AuditFrameworkNamesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    frameworks(first: $first, after: $after)
      @connection(key: "AuditGraph_frameworks") {
      edges {
        cursor
        node {
          id
          name
        }
      }
    }
  }
`;

export function useAuditFrameworkNames(
  auditRef: AuditFrameworkNames_audit$key,
): { label: string; names: string[] } {
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    AuditFrameworkNamesPaginationQuery,
    AuditFrameworkNames_audit$key
  >(auditFrameworkNamesFragment, auditRef);
  const failedRef = useRef(false);

  useEffect(() => {
    if (failedRef.current || !hasNext || isLoadingNext) {
      return;
    }

    loadNext(PAGE_SIZE, {
      onComplete: (error) => {
        if (error != null) {
          failedRef.current = true;
        }
      },
    });
  }, [hasNext, isLoadingNext, loadNext]);

  const names = data.frameworks?.edges?.flatMap((edge) => {
    const name = edge?.node?.name;
    return name ? [name] : [];
  }) ?? [];

  return { label: names.join(", "), names };
}
