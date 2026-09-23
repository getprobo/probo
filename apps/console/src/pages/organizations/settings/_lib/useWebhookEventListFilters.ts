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

export const webhookEventListStatusOptions = [
  "pending",
  "succeeded",
  "failed",
] as const;

export type WebhookEventListStatusOption = (typeof webhookEventListStatusOptions)[number];
export type WebhookEventListStatus = "all" | WebhookEventListStatusOption;

const graphqlStatuses = {
  pending: "PENDING",
  succeeded: "SUCCEEDED",
  failed: "FAILED",
} as const;

export function isWebhookEventListStatusOption(value: string): value is WebhookEventListStatusOption {
  return (webhookEventListStatusOptions as readonly string[]).includes(value);
}

export function webhookEventListGraphqlFilter(status: WebhookEventListStatus) {
  if (status === "all") {
    return null;
  }

  return { status: graphqlStatuses[status] };
}

export interface WebhookEventListFilters {
  status: WebhookEventListStatus;
  setStatus: (value: WebhookEventListStatus) => void;
  hasActiveFilters: boolean;
}

export function useWebhookEventListFilters(): WebhookEventListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const raw = searchParams.get("status") ?? "";
  const status: WebhookEventListStatus = isWebhookEventListStatusOption(raw) ? raw : "all";

  const setStatus = useCallback((value: WebhookEventListStatus) => {
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
