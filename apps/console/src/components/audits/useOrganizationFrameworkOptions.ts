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

import { useCallback, useEffect, useState } from "react";
import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { useOrganizationFrameworkOptions_organization$key } from "#/__generated__/core/useOrganizationFrameworkOptions_organization.graphql";
import type { useOrganizationFrameworkOptionsPaginationQuery } from "#/__generated__/core/useOrganizationFrameworkOptionsPaginationQuery.graphql";
import type { useOrganizationFrameworkOptionsQuery } from "#/__generated__/core/useOrganizationFrameworkOptionsQuery.graphql";

const PAGE_SIZE = 100;

const organizationFrameworkOptionsQuery = graphql`
  query useOrganizationFrameworkOptionsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...useOrganizationFrameworkOptions_organization
      }
    }
  }
`;

const organizationFrameworkOptionsFragment = graphql`
  fragment useOrganizationFrameworkOptions_organization on Organization
  @refetchable(queryName: "useOrganizationFrameworkOptionsPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 100 }
    after: { type: "CursorKey", defaultValue: null }
  ) {
    frameworks(first: $first, after: $after)
      @connection(key: "useOrganizationFrameworkOptions_frameworks") {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

export type OrganizationFrameworkOption = {
  id: string;
  name: string;
};

export function useOrganizationFrameworkOptions(
  organizationId: string,
  fetchPolicy?: "store-or-network" | "network-only",
): {
  frameworks: OrganizationFrameworkOption[];
  hasNext: boolean;
  failed: boolean;
  retry: () => void;
} {
  const queryData = useLazyLoadQuery<useOrganizationFrameworkOptionsQuery>(
    organizationFrameworkOptionsQuery,
    { organizationId },
    { fetchPolicy },
  );
  const { data, hasNext, isLoadingNext, loadNext } = usePaginationFragment<
    useOrganizationFrameworkOptionsPaginationQuery,
    useOrganizationFrameworkOptions_organization$key
  >(
    organizationFrameworkOptionsFragment,
    queryData.organization as useOrganizationFrameworkOptions_organization$key,
  );
  const [failed, setFailed] = useState(false);
  const retry = useCallback(() => {
    setFailed(false);
  }, []);

  useEffect(() => {
    if (failed || !hasNext || isLoadingNext) {
      return;
    }

    loadNext(PAGE_SIZE, {
      onComplete: (error) => {
        if (error != null) {
          setFailed(true);
        }
      },
    });
  }, [failed, hasNext, isLoadingNext, loadNext]);

  const frameworks = data.frameworks?.edges?.flatMap((edge) => {
    const framework = edge?.node;
    if (framework == null) {
      return [];
    }

    return [{ id: framework.id, name: framework.name }];
  }) ?? [];

  return { frameworks, hasNext, failed, retry };
}
