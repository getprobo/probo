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

import { useEffect } from "react";
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { ContextPageLoaderQuery } from "#/__generated__/core/ContextPageLoaderQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import ContextPage from "./ContextPage";

const contextPageQuery = graphql`
  query ContextPageLoaderQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...ContextPageFragment
      }
    }
  }
`;

function ContextPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ContextPageLoaderQuery>(contextPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [organizationId, loadQuery]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <ContextPageInner queryRef={queryRef} />;
}

function ContextPageInner({ queryRef }: { queryRef: PreloadedQuery<ContextPageLoaderQuery> }) {
  const data = usePreloadedQuery<ContextPageLoaderQuery>(contextPageQuery, queryRef);

  return <ContextPage organization={data.organization} />;
}

export default function ContextPageLoader() {
  return (
    <CoreRelayProvider>
      <ContextPageQueryLoader />
    </CoreRelayProvider>
  );
}
