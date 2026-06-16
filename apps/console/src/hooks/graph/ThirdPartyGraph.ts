// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { useMemo } from "react";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { ThirdPartyGraphCreateMutation } from "#/__generated__/core/ThirdPartyGraphCreateMutation.graphql";
import type { ThirdPartyGraphSelectQuery } from "#/__generated__/core/ThirdPartyGraphSelectQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const createThirdPartyMutation = graphql`
  mutation ThirdPartyGraphCreateMutation(
    $input: CreateThirdPartyInput!
    $connections: [ID!]!
  ) {
    createThirdParty(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
          canUpdate: permission(action: "core:thirdParty:update")
          canDelete: permission(action: "core:thirdParty:delete")
        }
      }
    }
  }
`;

export function useCreateThirdPartyMutation() {
  const { __ } = useTranslate();
  return useMutationWithToasts<ThirdPartyGraphCreateMutation>(createThirdPartyMutation, {
    successMessage: __("Third party created successfully"),
    errorMessage: __("Failed to create third party"),
  });
}

export const thirdPartiesSelectQuery = graphql`
  query ThirdPartyGraphSelectQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        thirdParties(
          first: 100
          orderBy: { direction: ASC, field: NAME }
        ) {
          edges {
            node {
              id
              name
              websiteUrl
              level
            }
          }
        }
      }
    }
  }
`;

export function useThirdParties(organizationId: string) {
  const data = useLazyLoadQuery<ThirdPartyGraphSelectQuery>(
    thirdPartiesSelectQuery,
    {
      organizationId: organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  return useMemo(() => {
    return data.organization?.thirdParties?.edges.map(edge => edge.node) ?? [];
  }, [data]);
}
