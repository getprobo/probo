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

import { dateTimeFormat } from "@probo/i18n";
import { Badge, Spinner } from "@probo/ui";
import { Suspense } from "react";
import { useTranslation } from "react-i18next";
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
}) {
  const { queryRef } = props;

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
      />
    </Suspense>
  );
}

function DocumentApprovalsPageContent(props: {
  approvalListRef: Parameters<typeof DocumentApprovalList>[0]["versionFragmentRef"];
  versionFragmentRef: DocumentApprovalsPage_versionFragment$key;
}) {
  const { approvalListRef, versionFragmentRef } = props;
  const { t, i18n } = useTranslation();

  const versionData = useFragment(versionFragment, versionFragmentRef);
  const quorumEdges = versionData.approvalQuorums?.edges ?? [];
  const pastQuorums = quorumEdges.slice(1);

  return (
    <div className="space-y-8">
      <DocumentApprovalList versionFragmentRef={approvalListRef} />

      {pastQuorums.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-sm font-medium text-txt-secondary">
            {t("documentApprovalsPage.previousRequests")}
          </h3>
          {pastQuorums.map(({ node: quorum }) => (
            <div key={quorum.id} className="border border-border-solid rounded-lg p-4">
              <div className="flex items-center gap-2 mb-3">
                <Badge variant={quorum.status === "APPROVED" ? "success" : quorum.status === "VOIDED" ? "neutral" : "danger"}>
                  {quorum.status === "APPROVED"
                    ? t("documentApprovalsPage.status.approved")
                    : quorum.status === "VOIDED"
                      ? t("documentApprovalsPage.status.voided")
                      : t("documentApprovalsPage.status.rejected")}
                </Badge>
                <span className="text-xs text-txt-secondary">
                  {dateTimeFormat(i18n.language, quorum.createdAt)}
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
                        {decision.decidedAt
                          && (decision.state === "APPROVED" || decision.state === "REJECTED")
                          && dateTimeFormat(i18n.language, decision.decidedAt)}
                      </div>
                      {decision.comment && (
                        <div className="text-xs text-txt-secondary italic">{decision.comment}</div>
                      )}
                    </div>
                    <div className="ml-auto">
                      <Badge variant={decision.state === "APPROVED" ? "success" : decision.state === "REJECTED" ? "danger" : decision.state === "VOIDED" ? "neutral" : "warning"}>
                        {decision.state === "APPROVED"
                          ? t("documentApprovalsPage.status.approved")
                          : decision.state === "REJECTED"
                            ? t("documentApprovalsPage.status.rejected")
                            : decision.state === "VOIDED"
                              ? t("documentApprovalsPage.status.voided")
                              : t("documentApprovalsPage.status.pending")}
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
