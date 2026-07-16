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
import { graphql, useFragment } from "react-relay";

import { CommitmentCard } from "#/components/CommitmentCard/CommitmentCard";

import type { SecurityCommitmentListItem_commitment$key } from "./__generated__/SecurityCommitmentListItem_commitment.graphql";
import { CommitmentIcon } from "./commitmentIcons";

const fragment = graphql`
  fragment SecurityCommitmentListItem_commitment on CompliancePortalCommitment {
    icon
    eyebrow
    title
    description
  }
`;

interface SecurityCommitmentListItemProps {
  commitmentKey: SecurityCommitmentListItem_commitment$key;
}

export function SecurityCommitmentListItem({ commitmentKey }: SecurityCommitmentListItemProps) {
  const commitment = useFragment(fragment, commitmentKey);

  return (
    <CommitmentCard
      icon={<CommitmentIcon icon={commitment.icon} size={32} weight="light" />}
      eyebrow={<Text size={1} color="gold">{commitment.eyebrow}</Text>}
      title={(
        <Text size={4} weight="medium" color="neutral" highContrast>
          {commitment.title}
        </Text>
      )}
      description={<Text size={2} color="neutral">{commitment.description}</Text>}
    />
  );
}
