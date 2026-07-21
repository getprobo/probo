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

import { Badge, Button, IconCrossLargeX, useConfirm } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentApprovalList_versionFragment$key } from "#/__generated__/core/DocumentApprovalList_versionFragment.graphql";
import type { DocumentApprovalList_voidMutation } from "#/__generated__/core/DocumentApprovalList_voidMutation.graphql";

import { DocumentApprovalListItem } from "./DocumentApprovalListItem";

const versionFragment = graphql`
  fragment DocumentApprovalList_versionFragment on DocumentVersion {
    id
    approvalQuorums(first: 100, orderBy: { field: CREATED_AT, direction: DESC }) {
      edges {
        node {
          status
          decisions(first: 100, orderBy: { field: CREATED_AT, direction: ASC })
            @connection(key: "DocumentApprovalList_decisions") {
            edges {
              node {
                id
                ...DocumentApprovalListItemFragment
              }
            }
          }
        }
      }
    }
  }
`;

const voidMutation = graphql`
  mutation DocumentApprovalList_voidMutation(
    $input: VoidDocumentVersionApprovalInput!
  ) {
    voidDocumentVersionApproval(input: $input) {
      documentVersion {
        id
        status
        major
        minor
        ...DocumentApprovalList_versionFragment
      }
      approvalQuorum {
        id
        status
      }
    }
  }
`;

export function DocumentApprovalList(props: {
  versionFragmentRef: DocumentApprovalList_versionFragment$key;
}) {
  const { versionFragmentRef } = props;
  const { t } = useTranslation();

  const version = useFragment(versionFragment, versionFragmentRef);

  const lastQuorum = version.approvalQuorums?.edges?.[0]?.node ?? null;
  const isPending = lastQuorum?.status === "PENDING";
  const edges = lastQuorum?.decisions?.edges ?? [];

  const [voidApproval, isVoiding]
    = useMutation<DocumentApprovalList_voidMutation>(voidMutation);
  const confirm = useConfirm();

  const handleVoid = () => {
    confirm(
      () =>
        new Promise<void>((resolve, reject) => {
          voidApproval({
            variables: {
              input: { documentVersionId: version.id },
            },
            onCompleted: (_, errors) => {
              if (errors?.length) {
                reject(new Error(errors[0].message));
              } else {
                resolve();
              }
            },
            onError: err => reject(err),
          });
        }),
      {
        message: t("documentApprovalList.confirmation.void"),
        label: t("documentApprovalList.actions.void"),
        variant: "danger",
      },
    );
  };

  const statusVariant = {
    PENDING: "warning",
    APPROVED: "success",
    REJECTED: "danger",
    VOIDED: "neutral",
  } as const;

  const statusLabel = {
    PENDING: t("documentApprovalList.status.pending"),
    APPROVED: t("documentApprovalList.status.approved"),
    REJECTED: t("documentApprovalList.status.rejected"),
    VOIDED: t("documentApprovalList.status.voided"),
  } as const;

  return (
    <div>
      {lastQuorum && (
        <div className="flex items-center justify-between mb-4">
          <Badge variant={statusVariant[lastQuorum.status]}>
            {statusLabel[lastQuorum.status]}
          </Badge>
          {isPending && (
            <Button
              variant="quaternary"
              icon={IconCrossLargeX}
              onClick={handleVoid}
              disabled={isVoiding}
            >
              {t("documentApprovalList.actions.cancel")}
            </Button>
          )}
        </div>
      )}

      {edges.length === 0
        ? (
            <div className="text-sm text-txt-secondary text-center py-8">
              {t("documentApprovalList.empty")}
            </div>
          )
        : (
            <div className="divide-y divide-border-solid">
              {edges.map(({ node }) => (
                <DocumentApprovalListItem
                  key={node.id}
                  fragmentRef={node}
                />
              ))}
            </div>
          )}
    </div>
  );
}
