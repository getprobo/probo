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

import type { MentionItem, MentionSource } from "@probo/ui";
import { useCallback } from "react";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";

import type { useMentionMembersQuery } from "#/__generated__/core/useMentionMembersQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const mentionMembersPageSize = 8;

const mentionMembersQuery = graphql`
  query useMentionMembersQuery($organizationId: ID!, $first: Int!, $filter: ProfileFilter) {
    organization: node(id: $organizationId) @required(action: THROW) {
      ... on Organization {
        profiles(
          first: $first
          orderBy: { direction: ASC, field: FULL_NAME }
          filter: $filter
        ) {
          edges {
            node {
              id
              fullName
              emailAddress
            }
          }
        }
      }
    }
  }
`;

export function useMentionMembers(): MentionSource {
  const environment = useRelayEnvironment();
  const organizationId = useOrganizationId();

  return useCallback(
    (query: string, signal: AbortSignal) =>
      new Promise<MentionItem[]>((resolve, reject) => {
        const subscription = fetchQuery<useMentionMembersQuery>(
          environment,
          mentionMembersQuery,
          {
            organizationId,
            first: mentionMembersPageSize,
            filter: {
              states: ["ACTIVE", "PENDING"],
              contractEnded: false,
              query: query.trim() || null,
            },
          },
          { fetchPolicy: "network-only" },
        ).subscribe({
          next: (data) => {
            const edges = data.organization.profiles?.edges ?? [];
            resolve(
              edges.map(({ node }) => ({
                id: node.id,
                label: node.fullName,
                description: node.emailAddress,
              })),
            );
          },
          error: reject,
        });

        signal.addEventListener("abort", () => {
          subscription.unsubscribe();
          resolve([]);
        });
      }),
    [environment, organizationId],
  );
}
