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

export const documentAccessListStatusOptions = [
  "requested",
  "granted",
  "none",
  "revoked",
  "rejected",
] as const;

export type DocumentAccessListStatusOption = (typeof documentAccessListStatusOptions)[number];
export type DocumentAccessListStatus = "all" | DocumentAccessListStatusOption;

const graphqlStatuses = {
  requested: "REQUESTED",
  granted: "GRANTED",
  none: "NONE",
  revoked: "REVOKED",
  rejected: "REJECTED",
} as const;

export function isDocumentAccessListStatusOption(value: string): value is DocumentAccessListStatusOption {
  return (documentAccessListStatusOptions as readonly string[]).includes(value);
}

export function documentAccessListGraphqlFilter(status: DocumentAccessListStatus) {
  if (status === "all") {
    return null;
  }

  return { status: graphqlStatuses[status] };
}

export interface DocumentAccessListFilters {
  status: DocumentAccessListStatus;
  setStatus: (value: DocumentAccessListStatus) => void;
  hasActiveFilters: boolean;
}

export function useDocumentAccessListFilters(): DocumentAccessListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const raw = searchParams.get("status") ?? "";
  const status: DocumentAccessListStatus = isDocumentAccessListStatusOption(raw) ? raw : "all";

  const setStatus = useCallback((value: DocumentAccessListStatus) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value === "all") {
        next.delete("status");
      } else {
        next.set("status", value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    status,
    setStatus,
    hasActiveFilters: status !== "all",
  };
}
