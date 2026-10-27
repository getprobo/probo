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

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { NotificationSettingsPageQuery } from "#/__generated__/iam/NotificationSettingsPageQuery.graphql";
import { IAMRelayProvider } from "#/lib/relay/IAMRelayProvider";
import { useEmployeePortalRoute } from "#/pages/iam/_lib/EmployeePortalRouteContext";

import {
  NotificationSettingsPage,
  notificationSettingsPageQuery,
} from "./NotificationSettingsPage";
import { NotificationSettingsPageSkeleton } from "./NotificationSettingsPageSkeleton";

function NotificationSettingsPageLoaderInner() {
  const { organizationId } = useEmployeePortalRoute();
  const [queryRef, loadQuery] = useQueryLoader<NotificationSettingsPageQuery>(
    notificationSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({ organizationId }, { fetchPolicy: "store-and-network" });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <NotificationSettingsPageSkeleton />;
  }

  return (
    <Suspense fallback={<NotificationSettingsPageSkeleton />}>
      <NotificationSettingsPage organizationId={organizationId} queryRef={currentQueryRef} />
    </Suspense>
  );
}

export default function NotificationSettingsPageLoader() {
  return (
    <IAMRelayProvider>
      <NotificationSettingsPageLoaderInner />
    </IAMRelayProvider>
  );
}
