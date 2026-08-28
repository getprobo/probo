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

import { graphql, useFragment } from "react-relay";

import type { ApprovalDashboardCard_viewer$key } from "#/__generated__/core/ApprovalDashboardCard_viewer.graphql";

import { DashboardCard } from "./DashboardCard";

const approvalDashboardCardFragment = graphql`
  fragment ApprovalDashboardCard_viewer on Viewer
  @argumentDefinitions(organizationId: { type: "ID!" })
  @throwOnFieldError {
    pendingApprovals: approvableDocuments(
      organizationId: $organizationId
      first: 1
      filter: { approvalStates: [PENDING] }
    ) {
      totalCount
      edges {
        node {
          id
        }
      }
    }
    approvedDocuments: approvableDocuments(
      organizationId: $organizationId
      filter: { approvalStates: [APPROVED] }
    ) {
      totalCount
    }
  }
`;

export interface ApprovalDashboardCardProps {
  viewerKey: ApprovalDashboardCard_viewer$key;
  wash?: boolean;
}

export function ApprovalDashboardCard({
  viewerKey,
  wash = false,
}: ApprovalDashboardCardProps) {
  const viewer = useFragment(approvalDashboardCardFragment, viewerKey);

  return (
    <DashboardCard
      kind="approvals"
      pendingCount={viewer.pendingApprovals.totalCount}
      completedCount={viewer.approvedDocuments.totalCount}
      firstPendingId={viewer.pendingApprovals.edges[0]?.node.id ?? null}
      wash={wash}
    />
  );
}
