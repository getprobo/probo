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

export type ConnectorConnectionStatus
  = "CONNECTED"
    | "DISCONNECTED"
    | "NOT_AUTHORIZED"
    | "RECONNECT_REQUIRED";

export const connectionStatuses = [
  "CONNECTED",
  "DISCONNECTED",
  "NOT_AUTHORIZED",
  "RECONNECT_REQUIRED",
] as const satisfies readonly ConnectorConnectionStatus[];

export type IntegrationsListFilter = {
  query: string;
} | null;

export function isConnectionStatus(value: string): value is ConnectorConnectionStatus {
  return (connectionStatuses as readonly string[]).includes(value);
}

export function integrationsListFilter(query: string): IntegrationsListFilter {
  const trimmed = query.trim();
  if (trimmed === "") {
    return null;
  }

  return { query: trimmed };
}

export function useIntegrationsListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawStatus = searchParams.get("status") ?? "";
  const status = isConnectionStatus(rawStatus) ? rawStatus : null;

  const setParam = useCallback((key: string, value: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set(key, value);
      } else {
        next.delete(key);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setQuery = useCallback((value: string) => {
    setParam("q", value);
  }, [setParam]);

  const setStatus = useCallback((value: ConnectorConnectionStatus | null) => {
    setParam("status", value ?? "");
  }, [setParam]);

  return { query, setQuery, status, setStatus };
}
