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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { graphql, useFragment } from "react-relay";

import { InlineErrorCard } from "#/components/errors/InlineErrorCard";

import type { SecurityCommitmentsSection_compliancePortal$key } from "./__generated__/SecurityCommitmentsSection_compliancePortal.graphql";
import { SecurityCommitmentGroupListItem } from "./SecurityCommitmentGroupListItem";
import { securityCommitments } from "./variants";

// @throwOnFieldError makes a field error in this fragment throw at the read
// below, where the section's ErrorBoundary contains it. See
// contrib/claude/error-handling.md.
const securityCommitmentsSectionFragment = graphql`
  fragment SecurityCommitmentsSection_compliancePortal on CompliancePortal @throwOnFieldError {
    commitmentGroups(first: 100) {
      edges {
        node {
          id
          # Same args as SecurityCommitmentGroupListItem_group so Relay merges
          # this into one fetch; used only to drop groups with no cards.
          commitments(first: 100) {
            edges {
              node {
                id
              }
            }
          }
          ...SecurityCommitmentGroupListItem_group
        }
      }
    }
  }
`;

interface SecurityCommitmentsSectionProps {
  compliancePortalKey: SecurityCommitmentsSection_compliancePortal$key;
}

// "Security Commitments" section: stacked groups, each a header above a grid of
// commitment cards. Wraps its data-reading content in a boundary so a load
// failure degrades to an inline error instead of taking down the page.
export function SecurityCommitmentsSection({ compliancePortalKey }: SecurityCommitmentsSectionProps) {
  return (
    <ErrorBoundary
      fallback={(
        <div className="w-full py-8">
          <InlineErrorCard onRetry={() => window.location.reload()} />
        </div>
      )}
    >
      <SecurityCommitmentsSectionContent compliancePortalKey={compliancePortalKey} />
    </ErrorBoundary>
  );
}

function SecurityCommitmentsSectionContent({ compliancePortalKey }: SecurityCommitmentsSectionProps) {
  const data = useFragment(securityCommitmentsSectionFragment, compliancePortalKey);
  const slots = securityCommitments();

  // Groups with no cards render nothing, so filter them out here to keep the
  // section eyebrow on the first visible group and to hide the whole section
  // (no empty padded gap) when nothing will render.
  const groups = data.commitmentGroups.edges
    .map(edge => edge.node)
    .filter(node => node.commitments.edges.length > 0);

  if (groups.length === 0) {
    return null;
  }

  return (
    <section className={slots.root()}>
      {groups.map((group, index) => (
        <SecurityCommitmentGroupListItem
          key={group.id}
          groupKey={group}
          showEyebrow={index === 0}
        />
      ))}
    </section>
  );
}
