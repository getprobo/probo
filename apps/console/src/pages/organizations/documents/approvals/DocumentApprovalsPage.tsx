// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { Badge, Spinner } from "@probo/ui";
import { Suspense } from "react";
import { type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentApprovalsPage_versionFragment$key } from "#/__generated__/core/DocumentApprovalsPage_versionFragment.graphql";
import type { DocumentApprovalsPageQuery } from "#/__generated__/core/DocumentApprovalsPageQuery.graphql";

import { DocumentApprovalList } from "./_components/DocumentApprovalList";

export const documentApprovalsPageQuery = graphql`
  query DocumentApprovalsPageQuery($documentId: ID! $versionId: ID! $versionSpecified: Boolean!) {
    # We use this on /documents/:documentId
    document: node(id: $documentId) @skip(if: $versionSpecified) {
      __typename
      ... on Document {
        lastVersion: versions(
          first: 1
          orderBy: { field: CREATED_AT, direction: DESC }
        ) {
          edges {
            node {
              ...DocumentApprovalList_versionFragment
              ...DocumentApprovalsPage_versionFragment
            }
          }
        }
      }
    }
    # We use this on /documents/:documentId/versions/:versionId
    version: node(id: $versionId) @include(if: $versionSpecified) {
      __typename
      ...DocumentApprovalList_versionFragment
      ...DocumentApprovalsPage_versionFragment
    }
  }
`;

const versionFragment = graphql`
  fragment DocumentApprovalsPage_versionFragment on DocumentVersion {
    approvalQuorums(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          id
          status
          createdAt
          decisions(first: 100, orderBy: { field: CREATED_AT, direction: ASC }) {
            edges {
              node {
                id
                approver {
                  fullName
                }
                state
                comment
                decidedAt
                createdAt
              }
            }
          }
        }
      }
    }
  }
`;

export function DocumentApprovalsPage(props: {
  queryRef: PreloadedQuery<DocumentApprovalsPageQuery>;
  onRefetch: () => void;
}) {
  const { queryRef, onRefetch } = props;

  const { document, version } = usePreloadedQuery<DocumentApprovalsPageQuery>(
    documentApprovalsPageQuery,
    queryRef,
  );

  if ((version && version.__typename !== "DocumentVersion") || (document && document.__typename !== "Document")) {
    throw new Error("invalid type for node");
  }
  if (!document && !version) {
    throw new Error("no document or version specified");
  }

  const lastVersionNode = document?.lastVersion.edges[0]?.node;
  const approvalListRef = version ?? lastVersionNode;
  const versionFragmentRef = (version ?? lastVersionNode) as DocumentApprovalsPage_versionFragment$key | null;
  if (!approvalListRef || !versionFragmentRef) {
    throw new Error("no version found");
  }

  return (
    <Suspense fallback={<Spinner centered />}>
      <DocumentApprovalsPageContent
        approvalListRef={approvalListRef}
        versionFragmentRef={versionFragmentRef}
        onRefetch={onRefetch}
      />
    </Suspense>
  );
}

function DocumentApprovalsPageContent(props: {
  approvalListRef: Parameters<typeof DocumentApprovalList>[0]["versionFragmentRef"];
  versionFragmentRef: DocumentApprovalsPage_versionFragment$key;
  onRefetch: () => void;
}) {
  const { approvalListRef, versionFragmentRef, onRefetch } = props;
  const { __, dateTimeFormat } = useTranslate();

  const versionData = useFragment(versionFragment, versionFragmentRef);
  const quorumEdges = versionData.approvalQuorums?.edges ?? [];
  const pastQuorums = quorumEdges.slice(1);

  return (
    <div className="space-y-8">
      <DocumentApprovalList versionFragmentRef={approvalListRef} onRefetch={onRefetch} />

      {pastQuorums.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-sm font-medium text-txt-secondary">{__("Previous approval requests")}</h3>
          {pastQuorums.map(({ node: quorum }) => (
            <div key={quorum.id} className="border border-border-solid rounded-lg p-4">
              <div className="flex items-center gap-2 mb-3">
                <Badge variant={quorum.status === "APPROVED" ? "success" : "danger"}>
                  {quorum.status === "APPROVED" ? __("Approved") : __("Rejected")}
                </Badge>
                <span className="text-xs text-txt-secondary">
                  {dateTimeFormat(quorum.createdAt)}
                </span>
              </div>
              <div className="divide-y divide-border-solid">
                {quorum.decisions.edges.map(({ node: decision }) => (
                  <div key={decision.id} className="flex items-center gap-3 py-2">
                    <div className="space-y-0.5">
                      <div className="text-sm text-txt-primary font-medium">
                        {decision.approver.fullName}
                      </div>
                      <div className="text-xs text-txt-secondary">
                        {decision.decidedAt && (decision.state === "APPROVED" || decision.state === "REJECTED") && dateTimeFormat(decision.decidedAt)}
                      </div>
                      {decision.comment && (
                        <div className="text-xs text-txt-secondary italic">{decision.comment}</div>
                      )}
                    </div>
                    <div className="ml-auto">
                      <Badge variant={decision.state === "APPROVED" ? "success" : decision.state === "REJECTED" ? "danger" : "warning"}>
                        {decision.state === "APPROVED" ? __("Approved") : decision.state === "REJECTED" ? __("Rejected") : __("Pending")}
                      </Badge>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
