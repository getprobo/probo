// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import { useParams } from "react-router";

import type { DevicePosturesPageQuery } from "#/__generated__/core/DevicePosturesPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import { DevicePosturesPage, devicePosturesPageQuery } from "./DevicePosturesPage";

function DevicePosturesPageQueryLoader() {
  const { deviceId } = useParams();
  if (!deviceId) {
    throw new Error(":deviceId missing in route params");
  }

  const [queryRef, loadQuery] = useQueryLoader<DevicePosturesPageQuery>(
    devicePosturesPageQuery,
  );

  useEffect(() => {
    if (!queryRef) {
      loadQuery({ deviceId });
    }
  });

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return (
    <Suspense fallback={<LinkCardSkeleton />}>
      <DevicePosturesPage queryRef={queryRef} />
    </Suspense>
  );
}

export default function DevicePosturesPageLoader() {
  return (
    <CoreRelayProvider>
      <DevicePosturesPageQueryLoader />
    </CoreRelayProvider>
  );
}
