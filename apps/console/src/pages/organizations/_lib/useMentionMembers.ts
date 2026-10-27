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
import { useMemo } from "react";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";
import type { IEnvironment } from "relay-runtime";

import type { useMentionMembersProfileQuery } from "#/__generated__/core/useMentionMembersProfileQuery.graphql";
import type { useMentionMembersQuery } from "#/__generated__/core/useMentionMembersQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const mentionMembersPageSize = 8;

const mentionedMembers = new Map<string, MentionItem>();
const mentionedMemberLoads = new Map<string, Promise<MentionItem | null>>();

function mentionCacheKey(organizationId: string, id: string): string {
  return `${organizationId}:${id}`;
}

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
              avatar {
                downloadUrl
              }
            }
          }
        }
      }
    }
  }
`;

const mentionMemberQuery = graphql`
  query useMentionMembersProfileQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on Profile {
        id
        fullName
        emailAddress
        avatar {
          downloadUrl
        }
      }
    }
  }
`;

function mentionItemFromProfile(node: {
  id: string;
  fullName: string;
  emailAddress: string;
  avatar?: { downloadUrl?: string | null } | null;
}): MentionItem {
  return {
    id: node.id,
    label: node.fullName,
    description: node.emailAddress,
    avatarUrl: node.avatar?.downloadUrl,
  };
}

function rememberMentionedMembers(organizationId: string, items: MentionItem[]) {
  for (const item of items) {
    mentionedMembers.set(mentionCacheKey(organizationId, item.id), item);
  }
}

function mentionItemFromNode(
  node: useMentionMembersProfileQuery["response"]["node"],
): MentionItem | null {
  if (node == null || node.__typename !== "Profile") {
    return null;
  }

  return mentionItemFromProfile(node);
}

function loadMentionedMember(environment: IEnvironment, organizationId: string, id: string) {
  const cacheKey = mentionCacheKey(organizationId, id);
  const cached = mentionedMembers.get(cacheKey);
  if (cached) {
    return Promise.resolve(cached);
  }

  const pending = mentionedMemberLoads.get(cacheKey);
  if (pending) {
    return pending;
  }

  const load = new Promise<MentionItem | null>((resolve) => {
    fetchQuery<useMentionMembersProfileQuery>(
      environment,
      mentionMemberQuery,
      { id },
      // Always re-authorize against the current organization/session rather
      // than reusing a Relay store entry a different context may have
      // populated for this same profile id.
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        const item = mentionItemFromNode(data.node);
        if (item) {
          mentionedMembers.set(cacheKey, item);
        }
        resolve(item);
      },
      error: (error: unknown) => {
        console.error("Cannot load mentioned member", error);
        resolve(null);
      },
    });
  }).finally(() => {
    mentionedMemberLoads.delete(cacheKey);
  });

  mentionedMemberLoads.set(cacheKey, load);
  return load;
}

export function useMentionMembers(): MentionSource {
  const environment = useRelayEnvironment();
  const organizationId = useOrganizationId();

  return useMemo(() => ({
    search: (query, signal) =>
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
            const items = edges.map(({ node }) => mentionItemFromProfile(node));
            rememberMentionedMembers(organizationId, items);
            resolve(items);
          },
          error: reject,
        });

        signal.addEventListener("abort", () => {
          subscription.unsubscribe();
          resolve([]);
        });
      }),
    resolve: (id, signal) => {
      if (signal.aborted) {
        return Promise.resolve(mentionedMembers.get(mentionCacheKey(organizationId, id)) ?? null);
      }

      return loadMentionedMember(environment, organizationId, id);
    },
  }), [environment, organizationId]);
}
