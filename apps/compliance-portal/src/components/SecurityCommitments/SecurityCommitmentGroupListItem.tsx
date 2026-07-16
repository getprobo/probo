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

import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { SecurityCommitmentGroupListItem_group$key } from "./__generated__/SecurityCommitmentGroupListItem_group.graphql";
import { SecurityCommitmentListItem } from "./SecurityCommitmentListItem";
import { securityCommitments } from "./variants";

const fragment = graphql`
  fragment SecurityCommitmentGroupListItem_group on CompliancePortalCommitmentGroup {
    title
    description
    commitments(first: 100) {
      edges {
        node {
          id
          ...SecurityCommitmentListItem_commitment
        }
      }
    }
  }
`;

interface SecurityCommitmentGroupListItemProps {
  groupKey: SecurityCommitmentGroupListItem_group$key;
  // Only the first group carries the section eyebrow, matching the design.
  showEyebrow: boolean;
}

export function SecurityCommitmentGroupListItem({ groupKey, showEyebrow }: SecurityCommitmentGroupListItemProps) {
  const { t } = useTranslation();
  const group = useFragment(fragment, groupKey);
  const slots = securityCommitments();

  const commitments = group.commitments.edges.map(edge => edge.node);

  if (commitments.length === 0) {
    return null;
  }

  return (
    <div className={slots.group()}>
      <div className={slots.groupHeader()}>
        {showEyebrow && (
          <Text size={1} color="gold">
            {t("home.sections.securityCommitments")}
          </Text>
        )}
        <Text size={2} weight="medium" color="neutral" highContrast>
          {group.title}
        </Text>
        <Text size={2} color="neutral">
          {group.description}
        </Text>
      </div>
      <div className={slots.grid()}>
        {commitments.map(commitment => (
          <SecurityCommitmentListItem key={commitment.id} commitmentKey={commitment} />
        ))}
      </div>
    </div>
  );
}
