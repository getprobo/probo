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

import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useCallback, useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useRelayEnvironment } from "react-relay";
import {
  createOperationDescriptor,
  getRequest,
  graphql,
  type IEnvironment,
  type Subscription,
} from "relay-runtime";

/* eslint-disable relay/unused-fields */

// Count only: a refresh landing after a close must not restore the old status.
export const pendingCountQuery = graphql`
  query usePendingCountRefreshQuery($campaignId: ID!) {
    node(id: $campaignId) {
      ... on AccessReviewCampaign {
        id
        pendingEntries: entries(first: 0, filter: { decision: PENDING }) {
          totalCount
        }
      }
    }
  }
`;

// Serial queue: one in-flight refresh, one queued rerun; refreshNow fires once on creation.
export function createPendingCountRefresh(
  environment: IEnvironment,
  campaignId: string,
  refreshNow: boolean,
  onError: () => void,
): { refresh: () => void; dispose: () => void } {
  let state: "idle" | "running" | "queued" | "disposed" = "idle";
  let subscription: Subscription | undefined;
  const operation = createOperationDescriptor(getRequest(pendingCountQuery), { campaignId });

  const refresh = () => {
    if (state === "disposed") {
      return;
    }
    if (state !== "idle") {
      state = "queued";
      return;
    }
    state = "running";
    const settle = (failed: boolean) => {
      const again = state === "queued";
      state = "idle";
      if (again) {
        refresh();
      } else if (failed) {
        onError();
      }
    };
    environment.execute({ operation }).subscribe({
      start: (s) => {
        subscription = s;
      },
      complete: () => settle(false),
      error: () => settle(true),
    });
  };

  const dispose = () => {
    state = "disposed";
    subscription?.unsubscribe();
  };

  if (refreshNow) {
    refresh();
  }
  return { refresh, dispose };
}

export function usePendingCountRefresh(campaignId: string, refreshNow: boolean): () => void {
  const environment = useRelayEnvironment();
  const { t } = useTranslation();
  const { add: addToast } = useToast();
  const refreshRef = useRef(() => {});

  useEffect(() => {
    const { refresh, dispose } = createPendingCountRefresh(environment, campaignId, refreshNow, () => {
      addToast({
        title: t("campaignDetailPage.messages.error"),
        description: t("campaignDetailPage.errors.pendingCountRefresh"),
        type: "error",
      });
    });
    refreshRef.current = refresh;
    return dispose;
  }, [environment, campaignId, refreshNow, t, addToast]);

  return useCallback(() => refreshRef.current(), []);
}
