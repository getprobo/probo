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

import type { AWSConnectFormQuery } from "#/__generated__/core/AWSConnectFormQuery.graphql";
import type { AzureConnectFormQuery } from "#/__generated__/core/AzureConnectFormQuery.graphql";
import type { ConnectVendorPageQuery } from "#/__generated__/core/ConnectVendorPageQuery.graphql";
import type { GCPConnectFormQuery } from "#/__generated__/core/GCPConnectFormQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { awsConnectFormQuery } from "./_components/AWSConnectForm";
import { azureConnectFormQuery } from "./_components/AzureConnectForm";
import { gcpConnectFormQuery } from "./_components/GCPConnectForm";
import { connectMethodFromSlug, providerFromSlug } from "./_lib/integrationPath";
import { ConnectVendorPage, connectVendorPageQuery } from "./ConnectVendorPage";

export default function ConnectVendorPageLoader() {
  const organizationId = useOrganizationId();
  const { provider: providerSlug = "", method: methodSlug } = useParams();
  const provider = providerFromSlug(providerSlug);
  const method = methodSlug == null ? null : connectMethodFromSlug(methodSlug);
  const [queryRef, loadQuery]
    = useQueryLoader<ConnectVendorPageQuery>(connectVendorPageQuery);
  const [awsQueryRef, loadAwsQuery]
    = useQueryLoader<AWSConnectFormQuery>(awsConnectFormQuery);
  const [gcpQueryRef, loadGcpQuery]
    = useQueryLoader<GCPConnectFormQuery>(gcpConnectFormQuery);
  const [azureQueryRef, loadAzureQuery]
    = useQueryLoader<AzureConnectFormQuery>(azureConnectFormQuery);

  useEffect(() => {
    loadQuery({ organizationId });
    if (method !== "WORKLOAD_IDENTITY") {
      return;
    }
    if (provider === "AWS") {
      loadAwsQuery({ organizationId });
    } else if (provider === "GCP") {
      loadGcpQuery({ organizationId });
    } else if (provider === "AZURE") {
      loadAzureQuery({ organizationId });
    }
  }, [
    loadAwsQuery,
    loadAzureQuery,
    loadGcpQuery,
    loadQuery,
    method,
    organizationId,
    provider,
  ]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;
  const setupReady = method !== "WORKLOAD_IDENTITY"
    || (provider !== "AWS" && provider !== "GCP" && provider !== "AZURE")
    || (provider === "AWS"
      && awsQueryRef != null
      && awsQueryRef.variables.organizationId === organizationId)
    || (provider === "GCP"
      && gcpQueryRef != null
      && gcpQueryRef.variables.organizationId === organizationId)
    || (provider === "AZURE"
      && azureQueryRef != null
      && azureQueryRef.variables.organizationId === organizationId);

  if (currentQueryRef == null || !setupReady) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <ConnectVendorPage
        queryRef={currentQueryRef}
        awsQueryRef={provider === "AWS" ? awsQueryRef ?? null : null}
        gcpQueryRef={provider === "GCP" ? gcpQueryRef ?? null : null}
        azureQueryRef={provider === "AZURE" ? azureQueryRef ?? null : null}
      />
    </Suspense>
  );
}
