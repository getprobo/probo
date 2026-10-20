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

import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { connectMethodFromSlug } from "../../_lib/connectMethods";
import { providerFromSlug } from "../../_lib/integrationPath";

import { ConnectVendorPage, connectVendorPageQuery } from "./ConnectVendorPage";

export default function ConnectVendorPageLoader() {
  const organizationId = useOrganizationId();
  const { provider: providerSlug, method: methodSlug } = useParams<{
    provider?: string;
    method?: string;
  }>();
  const { includeAWS, includeAzure, includeGCP } = workloadIdentityIncludes(
    providerSlug,
    methodSlug,
  );
  const [queryRef, loadQuery]
    = useQueryLoader<ConnectVendorPageQuery>(connectVendorPageQuery);

  useEffect(() => {
    loadQuery({
      organizationId,
      includeAWS,
      includeAzure,
      includeGCP,
    });
  }, [loadQuery, organizationId, includeAWS, includeAzure, includeGCP]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    && queryRef.variables.includeAWS === includeAWS
    && queryRef.variables.includeAzure === includeAzure
    && queryRef.variables.includeGCP === includeGCP
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <ConnectVendorPage queryRef={currentQueryRef} />
    </Suspense>
  );
}

function workloadIdentityIncludes(
  providerSlug: string | undefined,
  methodSlug: string | undefined,
) {
  const provider = providerSlug == null ? null : providerFromSlug(providerSlug);
  const method = methodSlug == null ? null : connectMethodFromSlug(methodSlug);
  const workloadIdentity = method === "WORKLOAD_IDENTITY";

  return {
    includeAWS: workloadIdentity && provider === "AWS",
    includeAzure: workloadIdentity && provider === "AZURE",
    includeGCP: workloadIdentity && provider === "GCP",
  };
}
