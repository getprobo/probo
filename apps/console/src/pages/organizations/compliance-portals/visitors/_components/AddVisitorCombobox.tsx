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
import { useTranslation } from "react-i18next";
import type { PreloadedQuery } from "react-relay";
import { graphql, usePreloadedQuery } from "react-relay";

import type { AddVisitorComboboxQuery } from "#/__generated__/core/AddVisitorComboboxQuery.graphql";

import { visitorDisplayName } from "../_lib/visitorIdentity";
import { addVisitorPopover } from "../variants";

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
  showEmpty?: boolean;
  disabled?: boolean;
}

export function AddVisitorCombobox({
  queryRef,
  onSelect,
  showEmpty = true,
  disabled = false,
}: AddVisitorComboboxProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const data = usePreloadedQuery<AddVisitorComboboxQuery>(
    addVisitorComboboxQuery,
    queryRef,
  );
  if (data.node?.__typename !== "CompliancePortal") {
    return null;
  }

  const { item, empty, avatar, identity, name, email } = addVisitorPopover();
  const candidates = data.node.memberCandidates;
  if (candidates.length === 0) {
    if (!showEmpty) {
      return null;
    }

    return <p className={empty()}>{t("addVisitorDialog.empty")}</p>;
  }

  return (
    <>
      {candidates.map((candidate) => {
        const displayName = visitorDisplayName(
          candidate.fullName,
          candidate.emailAddress,
        );

        return (
          <button
            key={candidate.id}
            type="button"
            className={item()}
            disabled={disabled}
            onClick={() => {
              onSelect({
                id: candidate.id,
                fullName: candidate.fullName,
                emailAddress: candidate.emailAddress,
              });
            }}
          >
            <Avatar
              size={2}
              variant="soft"
              color="gold"
              className={avatar()}
              fallback={displayName.charAt(0).toUpperCase() || "?"}
            />
            <span className={identity()}>
              <span className={name()}>{displayName}</span>
              {candidate.fullName.trim() !== "" && (
                <span className={email()}>{candidate.emailAddress}</span>
              )}
            </span>
          </button>
        );
      })}
    </>
  );
}
