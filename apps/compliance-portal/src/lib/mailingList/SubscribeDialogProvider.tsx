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

import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { graphql, useFragment } from "react-relay";
import { useSearchParams } from "react-router";

import { SubscribeDialog } from "#/components/SubscribeDialog/SubscribeDialog";
import { buildSubscribeContinueUrl, redirectToInitiate, SUBSCRIBE_PARAM } from "#/lib/auth/continueUrl";
import {
  SubscribeDialogContextProvider,
} from "#/lib/mailingList/subscribeDialogContext";
import { useUnsubscribeFromMailingList } from "#/lib/mailingList/useUnsubscribeFromMailingList";

import type { SubscribeDialogProvider_query$key } from "./__generated__/SubscribeDialogProvider_query.graphql";

export const subscribeDialogProviderFragment = graphql`
  fragment SubscribeDialogProvider_query on Query {
    viewer {
      email
    }
    currentCompliancePortal @required(action: THROW) {
      id
      title
      viewerSubscription {
        id
      }
    }
  }
`;

interface SubscribeDialogProviderProps {
  queryKey: SubscribeDialogProvider_query$key;
  children: ReactNode;
}

// Owns the subscribe dialog and exposes openSubscribe / unsubscribe / isSubscribed
// to Updates CTAs and the TopBar user menu. Guests are sent through Sign-in with
// a continue URL that re-opens the dialog after authentication.
export function SubscribeDialogProvider({
  queryKey,
  children,
}: SubscribeDialogProviderProps) {
  const data = useFragment(subscribeDialogProviderFragment, queryKey);
  const [searchParams, setSearchParams] = useSearchParams();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [unsubscribeFromMailingList, isUnsubscribing] = useUnsubscribeFromMailingList();

  const viewer = data.viewer;
  const { id: trustCenterId, title, viewerSubscription } = data.currentCompliancePortal;
  const isSubscribed = viewerSubscription != null;

  const openSubscribe = useCallback(() => {
    if (viewer == null) {
      redirectToInitiate(buildSubscribeContinueUrl());
      return;
    }
    setDialogOpen(true);
  }, [viewer]);

  const unsubscribe = useCallback(async () => {
    try {
      await unsubscribeFromMailingList({ variables: {} });
    } catch {
      // Errors are surfaced by the mutation notifier.
    }
  }, [unsubscribeFromMailingList]);

  // After a guest signs in to subscribe, they land back with the subscribe
  // marker; open the dialog once and drop the marker so a reload can't re-open.
  // Derive the next params from the updater's previous snapshot so a concurrent
  // effect that already cleared an access-request marker is not undone.
  const resumed = useRef(false);
  useEffect(() => {
    if (resumed.current || viewer == null || searchParams.get(SUBSCRIBE_PARAM) == null) {
      return;
    }
    resumed.current = true;
    setDialogOpen(true);
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.delete(SUBSCRIBE_PARAM);
      return next;
    }, { replace: true });
  }, [viewer, searchParams, setSearchParams]);

  const value = useMemo(
    () => ({
      openSubscribe,
      isSubscribed,
      unsubscribe,
      isUnsubscribing,
    }),
    [openSubscribe, isSubscribed, unsubscribe, isUnsubscribing],
  );

  return (
    <SubscribeDialogContextProvider value={value}>
      {children}
      {viewer != null && (
        <SubscribeDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          trustCenterId={trustCenterId}
          viewerEmail={viewer.email}
          organizationName={title}
        />
      )}
    </SubscribeDialogContextProvider>
  );
}
