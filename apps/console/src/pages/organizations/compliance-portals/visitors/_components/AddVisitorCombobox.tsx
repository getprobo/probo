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

import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type { AddVisitorComboboxQuery } from "#/__generated__/core/AddVisitorComboboxQuery.graphql";

import { visitorDisplayName } from "../_lib/visitorIdentity";
import { addVisitorDialog } from "../variants";

export const addVisitorComboboxQuery = graphql`
  query AddVisitorComboboxQuery($compliancePortalId: ID!, $query: String!) {
    node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        memberCandidates(query: $query) {
          id
          fullName
          emailAddress
        }
      }
    }
  }
`;

export interface AddVisitorCandidate {
  id: string;
  fullName: string;
  emailAddress: string;
}

interface AddVisitorComboboxProps {
  queryRef: PreloadedQuery<AddVisitorComboboxQuery>;
  onSelect: (candidate: AddVisitorCandidate) => void;
}

export function AddVisitorCombobox({
  queryRef,
  onSelect,
}: AddVisitorComboboxProps) {
  const data = usePreloadedQuery<AddVisitorComboboxQuery>(
    addVisitorComboboxQuery,
    queryRef,
  );
  if (data.node?.__typename !== "CompliancePortal") {
    return null;
  }

  const candidates = data.node.memberCandidates;
  if (candidates.length === 0) {
    return null;
  }

  const { item, hit, row, avatar, name, email } = addVisitorDialog();

  return (
    <List>
      {candidates.map((candidate) => {
        const displayName = visitorDisplayName(
          candidate.fullName,
          candidate.emailAddress,
        );

        return (
          <ListItem key={candidate.id} className={item()}>
            <button
              type="button"
              className={hit()}
              aria-label={displayName}
              onClick={() => {
                onSelect({
                  id: candidate.id,
                  fullName: candidate.fullName,
                  emailAddress: candidate.emailAddress,
                });
              }}
            />
            <div className={row()}>
              <Avatar
                size={3}
                variant="soft"
                color="gold"
                className={avatar()}
                fallback={displayName.charAt(0).toUpperCase() || "?"}
              />
              <ListItemContent>
                <Text size={2} weight="medium" color="neutral" highContrast className={name()}>
                  {displayName}
                </Text>
                {candidate.fullName.trim() !== "" && (
                  <Text size={1} color="gold" className={email()}>
                    {candidate.emailAddress}
                  </Text>
                )}
              </ListItemContent>
            </div>
          </ListItem>
        );
      })}
    </List>
  );
}
