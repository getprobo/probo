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

import {
  type FindingKind,
  type FindingPriority,
  type FindingStatus,
  isFindingKind,
  isFindingPriority,
  isFindingStatus,
} from "./findingOptions";

export const findingsPageSize = 100;

export interface FindingsListVariables {
  first: number | null;
  after: string | null;
  last: number | null;
  before: string | null;
  kind: FindingKind | null;
  status: FindingStatus | null;
  priority: FindingPriority | null;
  ownerId: string | null;
  auditId: string | null;
}

function pageNumber(value: string | null): number {
  const page = Number(value);
  if (!Number.isInteger(page) || page < 1) {
    return 1;
  }
  return page;
}

function param(value: string | null) {
  return value === "" ? null : value;
}

export function findingsListVariables(input: {
  kind: FindingKind | null;
  status: FindingStatus | null;
  priority: FindingPriority | null;
  ownerId: string | null;
  auditId: string | null;
  after: string | null;
  before: string | null;
}): FindingsListVariables {
  const after = input.after || null;
  const before = after ? null : input.before || null;

  return {
    first: before ? null : findingsPageSize,
    after: before ? null : after,
    last: before ? findingsPageSize : null,
    before,
    kind: input.kind,
    status: input.status,
    priority: input.priority,
    ownerId: input.ownerId,
    auditId: input.auditId,
  };
}

export function findingsListVariablesKey(variables: Partial<FindingsListVariables>) {
  return JSON.stringify([
    variables.first ?? null,
    variables.after ?? null,
    variables.last ?? null,
    variables.before ?? null,
    variables.kind ?? null,
    variables.status ?? null,
    variables.priority ?? null,
    variables.ownerId ?? null,
    variables.auditId ?? null,
  ]);
}

export function useFindingsListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const kindValue = searchParams.get("kind") ?? "";
  const kind = isFindingKind(kindValue) ? kindValue : null;
  const statusValue = searchParams.get("status") ?? "";
  const status = isFindingStatus(statusValue) ? statusValue : null;
  const priorityValue = searchParams.get("priority") ?? "";
  const priority = isFindingPriority(priorityValue) ? priorityValue : null;
  const ownerId = param(searchParams.get("owner"));
  const auditId = param(searchParams.get("audit"));
  const after = searchParams.get("after") || null;
  const before = after ? null : searchParams.get("before") || null;
  const page = after == null && before == null
    ? 1
    : pageNumber(searchParams.get("page"));

  const write = useCallback((
    mutate: (next: URLSearchParams) => void,
    options: { resetPage: boolean; replace: boolean },
  ) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      mutate(next);
      if (options.resetPage) {
        next.delete("after");
        next.delete("before");
        next.delete("page");
      }
      return next;
    }, { replace: options.replace });
  }, [setSearchParams]);

  const setText = useCallback((key: string, value: string | null) => {
    write((next) => {
      if (value) {
        next.set(key, value);
      } else {
        next.delete(key);
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setKind = useCallback((value: FindingKind | null) => {
    setText("kind", value);
  }, [setText]);

  const setStatus = useCallback((value: FindingStatus | null) => {
    setText("status", value);
  }, [setText]);

  const setPriority = useCallback((value: FindingPriority | null) => {
    setText("priority", value);
  }, [setText]);

  const setOwnerId = useCallback((value: string | null) => {
    setText("owner", value);
  }, [setText]);

  const setAuditId = useCallback((value: string | null) => {
    setText("audit", value);
  }, [setText]);

  const setPageCursor = useCallback((cursor: {
    after?: string | null;
    before?: string | null;
    page: number;
  }) => {
    write((next) => {
      next.delete("after");
      next.delete("before");
      next.delete("page");
      if (cursor.after) {
        next.set("after", cursor.after);
      }
      if (cursor.before) {
        next.set("before", cursor.before);
      }
      if (cursor.page > 1) {
        next.set("page", String(cursor.page));
      }
    }, { resetPage: false, replace: false });
  }, [write]);

  const clear = useCallback(() => {
    write((next) => {
      next.delete("q");
      next.delete("kind");
      next.delete("status");
      next.delete("priority");
      next.delete("owner");
      next.delete("audit");
    }, { resetPage: true, replace: true });
  }, [write]);

  const hasActiveFilters = kind != null
    || status != null
    || priority != null
    || ownerId != null
    || auditId != null;

  return {
    kind,
    status,
    priority,
    ownerId,
    auditId,
    after,
    before,
    page,
    hasActiveFilters,
    variables: findingsListVariables({
      kind,
      status,
      priority,
      ownerId,
      auditId,
      after,
      before,
    }),
    setKind,
    setStatus,
    setPriority,
    setOwnerId,
    setAuditId,
    setPageCursor,
    clear,
  };
}
