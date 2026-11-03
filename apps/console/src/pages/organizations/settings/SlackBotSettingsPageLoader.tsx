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

import { CardSkeleton } from "@probo/ui/src/v2/Card/CardSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";

import type { SlackBotSettingsPageQuery } from "#/__generated__/core/SlackBotSettingsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import {
  SlackBotSettingsPage,
  slackBotSettingsPageQuery,
} from "./SlackBotSettingsPage";
import { slackBotSettingsPageSkeleton } from "./variants";

function SlackBotSettingsPageSkeleton() {
  const { root, grid } = slackBotSettingsPageSkeleton();

  return (
    <div className={root()}>
      <HeadingSkeleton size={4} className="w-20" />
      <div className={grid()}>
        <CardSkeleton size={5} />
      </div>
    </div>
  );
}

function SlackBotSettingsPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<SlackBotSettingsPageQuery>(
    slackBotSettingsPageQuery,
  );

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <SlackBotSettingsPageSkeleton />;
  }

  return (
    <Suspense fallback={<SlackBotSettingsPageSkeleton />}>
      <SlackBotSettingsPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function SlackBotSettingsPageLoader() {
  return (
    <CoreRelayProvider>
      <SlackBotSettingsPageQueryLoader />
    </CoreRelayProvider>
  );
}
