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
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import { useMutation } from "#/lib/relay/useMutation";

import type { useSubscribeToMailingListMutation } from "./__generated__/useSubscribeToMailingListMutation.graphql";

const subscribeToMailingListMutation = graphql`
  mutation useSubscribeToMailingListMutation {
    subscribeToMailingList {
      subscription {
        id
      }
    }
  }
`;

// Subscribes the authenticated viewer to the trust center mailing list and
// links the new subscriber onto currentTrustCenter.viewerSubscription.
export function useSubscribeToMailingList(trustCenterId: string) {
  const { t } = useTranslation("updates");
  const [commit, isSubscribing] = useMutation<useSubscribeToMailingListMutation>(
    subscribeToMailingListMutation,
    { successMessage: t("dialog.successToast") },
  );

  const subscribe = useCallback(async () => {
    await commit({
      variables: {},
      updater: (store, data) => {
        const subscription = data?.subscribeToMailingList?.subscription;
        if (!subscription?.id) {
          return;
        }
        const trustCenterRecord = store.get(trustCenterId);
        const subscriptionRecord = store.get(subscription.id);
        if (trustCenterRecord == null || subscriptionRecord == null) {
          return;
        }
        trustCenterRecord.setLinkedRecord(subscriptionRecord, "viewerSubscription");
      },
    });
  }, [commit, trustCenterId]);

  return [subscribe, isSubscribing] as const;
}
