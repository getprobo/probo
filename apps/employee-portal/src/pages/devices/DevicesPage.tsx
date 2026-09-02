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

import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { DevicesPageQuery } from "#/__generated__/core/DevicesPageQuery.graphql";
import { NotFoundError } from "#/lib/relay/errors";

import { DevicesList } from "./_components/DevicesList";
import { devicesPage } from "./_components/variants";

export const devicesPageQuery = graphql`
  query DevicesPageQuery($organizationId: ID!, $first: Int) @throwOnFieldError {
    viewer @required(action: THROW) {
      ...DevicesList_viewer @arguments(organizationId: $organizationId, first: $first)
    }
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...DevicesList_organization
      }
    }
  }
`;

interface DevicesPageProps {
  queryRef: PreloadedQuery<DevicesPageQuery>;
}

export function DevicesPage({ queryRef }: DevicesPageProps) {
  const slots = devicesPage();
  const { viewer, organization } = usePreloadedQuery<DevicesPageQuery>(
    devicesPageQuery,
    queryRef,
  );

  if (organization?.__typename !== "Organization") {
    throw new NotFoundError("invalid type for organization node");
  }

  return (
    <main className={slots.main()}>
      <DevicesList viewerKey={viewer} organizationKey={organization} />
    </main>
  );
}
