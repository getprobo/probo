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

import type { EnrollDevicePageQuery } from "#/__generated__/iam/EnrollDevicePageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  EnrollDevicePage,
  enrollDevicePageQuery,
} from "./EnrollDevicePage";

function EnrollDevicePageQueryLoader() {
  const [queryRef, loadQuery] = useQueryLoader<EnrollDevicePageQuery>(
    enrollDevicePageQuery,
  );

  useEffect(() => {
    loadQuery({});
  }, [loadQuery]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return <EnrollDevicePage queryRef={queryRef} />;
}

export default function EnrollDevicePageLoader() {
  return (
    <IAMRelayProvider>
      <Suspense fallback={<PageSkeleton />}>
        <EnrollDevicePageQueryLoader />
      </Suspense>
    </IAMRelayProvider>
  );
}
