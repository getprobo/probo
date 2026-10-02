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

import type { ConnectorDetailsPageQuery } from "#/__generated__/core/ConnectorDetailsPageQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { providerFromSlug } from "./_lib/integrationPath";
import {
  ConnectorDetailsPage,
  connectorDetailsPageQuery,
} from "./ConnectorDetailsPage";
import { ConnectorDetailsPageSkeleton } from "./ConnectorDetailsPageSkeleton";

export default function ConnectorDetailsPageLoader() {
  const organizationId = useOrganizationId();
  const { provider: providerSlug } = useParams<{ provider: string }>();
  const provider = providerSlug == null ? null : providerFromSlug(providerSlug);
  const [queryRef, loadQuery] = useQueryLoader<ConnectorDetailsPageQuery>(
    connectorDetailsPageQuery,
  );

  useEffect(() => {
    if (provider == null) {
      return;
    }
    loadQuery({
      organizationId,
      provider: provider as ConnectorDetailsPageQuery["variables"]["provider"],
    }, { fetchPolicy: "store-and-network" });
  }, [loadQuery, organizationId, provider]);

  if (provider == null) {
    throw new Error(":provider missing in route params");
  }

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.provider === provider
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <ConnectorDetailsPageSkeleton />;
  }

  return (
    <Suspense fallback={<ConnectorDetailsPageSkeleton />}>
      <ConnectorDetailsPage key={provider} queryRef={currentQueryRef} />
    </Suspense>
  );
}
