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
import { useParams } from "react-router";

import type { SAMLConfigurationPageQuery } from "#/__generated__/iam/SAMLConfigurationPageQuery.graphql";
import { IAMRelayProvider } from "#/providers/IAMRelayProvider";

import {
  SAMLConfigurationPage,
  samlConfigurationPageQuery,
} from "./SAMLConfigurationPage";
import { SAMLConfigurationPageSkeleton } from "./SAMLConfigurationPageSkeleton";

function SAMLConfigurationPageQueryLoader() {
  const { samlConfigurationId } = useParams<{ samlConfigurationId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<SAMLConfigurationPageQuery>(
    samlConfigurationPageQuery,
  );

  useEffect(() => {
    if (samlConfigurationId != null) {
      loadQuery({ samlConfigurationId });
    }
  }, [loadQuery, samlConfigurationId]);

  if (samlConfigurationId == null) {
    throw new Error(":samlConfigurationId missing in route params");
  }

  const currentQueryRef = queryRef != null
    && queryRef.variables.samlConfigurationId === samlConfigurationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <SAMLConfigurationPageSkeleton />;
  }

  return (
    <Suspense fallback={<SAMLConfigurationPageSkeleton />}>
      <SAMLConfigurationPage
        key={samlConfigurationId}
        queryRef={currentQueryRef}
      />
    </Suspense>
  );
}

export default function SAMLConfigurationPageLoader() {
  return (
    <IAMRelayProvider>
      <SAMLConfigurationPageQueryLoader />
    </IAMRelayProvider>
  );
}
