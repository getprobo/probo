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

import { PlusIcon, StampIcon, XIcon } from "@phosphor-icons/react";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, usePaginationFragment } from "react-relay";

import type { DocumentApprovalsSection_document$key } from "#/__generated__/core/DocumentApprovalsSection_document.graphql";
import type { DocumentApprovalsSectionDecisions_quorum$data } from "#/__generated__/core/DocumentApprovalsSectionDecisions_quorum.graphql";
import type { DocumentApprovalsSectionDecisions_quorum$key } from "#/__generated__/core/DocumentApprovalsSectionDecisions_quorum.graphql";
import type { DocumentApprovalsSectionDecisionsRefetchQuery } from "#/__generated__/core/DocumentApprovalsSectionDecisionsRefetchQuery.graphql";
import { PeopleMenu } from "#/components/people/PeopleMenu";

import { useUpdateDocument } from "../_lib/useUpdateDocument";
import { documentApprovals } from "../variants";

const documentApprovalsPageSize = 10;

const documentApprovalsSectionFragment = graphql`
  fragment DocumentApprovalsSection_document on Document {
    id
    status
    canUpdate: permission(action: "core:document:update")
    defaultApprovers {
      id
      fullName
    }
    pendingApproval: versions(first: 1, filter: { statuses: [PENDING_APPROVAL] }) {
      edges {
        node {
          approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                status
                ...DocumentApprovalsSectionDecisions_quorum
              }
            }
          }
        }
      }
    }
  }
`;

const documentApprovalsDecisionsFragment = graphql`
  fragment DocumentApprovalsSectionDecisions_quorum on DocumentVersionApprovalQuorum
  @refetchable(queryName: "DocumentApprovalsSectionDecisionsRefetchQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 10 }
    after: { type: "CursorKey" }
  ) {
    decisions(
      first: $first
      after: $after
      orderBy: { field: CREATED_AT, direction: ASC }
    ) @connection(key: "DocumentApprovalsSection_decisions") {
      edges {
        node {
          id
          state
          approver {
            id
            fullName
          }
        }
      }
    }
  }
`;

type DecisionState = DocumentApprovalsSectionDecisions_quorum$data[
  "decisions"
]["edges"][number]["node"]["state"];

interface DocumentApprovalsSectionProps {
  documentKey: DocumentApprovalsSection_document$key;
}

function stateColor(state: DecisionState) {
  switch (state) {
    case "APPROVED":
      return "green";
    case "REJECTED":
      return "red";
    case "PENDING":
      return "amber";
    case "VOIDED":
      return "neutral";
  }
}

export function DocumentApprovalsSection({ documentKey }: DocumentApprovalsSectionProps) {
  const document = useFragment(documentApprovalsSectionFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const [updateDocument, isUpdating] = useUpdateDocument();
  const [visibleCount, setVisibleCount] = useState(documentApprovalsPageSize);
  const { root, title, titleLead, titleIcon, rows, row, name } = documentApprovals();
  const quorum = document.pendingApproval.edges[0]?.node.approvalQuorums.edges[0]?.node ?? null;
  const pending = quorum?.status === "PENDING";
  const canEdit = document.canUpdate && document.status !== "ARCHIVED" && !pending;
  const savedApprovers = document.defaultApprovers;
  const approverIds = savedApprovers.map(person => person.id);
  const shownApprovers = savedApprovers.slice(0, visibleCount);

  function save(defaultApproverIds: string[]) {
    return updateDocument({
      variables: {
        input: {
          id: document.id,
          defaultApproverIds,
        },
      },
    });
  }

  return (
    <Card variant="soft" size={2}>
      <div className={root()}>
        <div className={title()}>
          <div className={titleLead()}>
            <StampIcon className={titleIcon()} aria-hidden />
            <Heading level={2} size={2} weight="medium">
              {t("details.approvals.title")}
            </Heading>
          </div>
          {canEdit && (
            <Dropdown>
              <DropdownTrigger
                disabled={isUpdating}
                render={(
                  <IconButton
                    size={1}
                    variant="ghost"
                    color="neutral"
                    disabled={isUpdating}
                    aria-label={t("details.approvals.add")}
                  />
                )}
              >
                <PlusIcon />
              </DropdownTrigger>
              <DropdownPopup align="end">
                <PeopleMenu
                  value={null}
                  emptyValue={null}
                  disabled={isUpdating}
                  onValueChange={(approverId) => {
                    if (approverId == null || approverIds.includes(approverId)) {
                      return;
                    }
                    void save([...approverIds, approverId]).then(() => {
                      setVisibleCount(approverIds.length + 1);
                    }).catch(() => {
                      // Error toast is already shown by useMutation.
                    });
                  }}
                />
              </DropdownPopup>
            </Dropdown>
          )}
        </div>
        <div className={rows()}>
          {pending && quorum != null
            ? <DocumentApprovalDecisions quorumKey={quorum} />
            : shownApprovers.length > 0
              ? (
                  <>
                    {shownApprovers.map(person => (
                      <div key={person.id} className={row()}>
                        <Text size={2} className={name()} title={person.fullName}>
                          {person.fullName}
                        </Text>
                        {canEdit && (
                          <IconButton
                            size={1}
                            variant="ghost"
                            color="neutral"
                            disabled={isUpdating}
                            aria-label={t("details.approvals.remove", { name: person.fullName })}
                            onClick={() => {
                              void save(approverIds.filter(id => id !== person.id)).catch(() => {
                                // Error toast is already shown by useMutation.
                              });
                            }}
                          >
                            <XIcon />
                          </IconButton>
                        )}
                      </div>
                    ))}
                    {savedApprovers.length > shownApprovers.length && (
                      <Button
                        variant="ghost"
                        color="neutral"
                        size={1}
                        className="self-start"
                        onClick={() => {
                          setVisibleCount(count => count + documentApprovalsPageSize);
                        }}
                      >
                        {t("details.approvals.loadMore")}
                      </Button>
                    )}
                  </>
                )
              : (
                  <Text size={1} color="faint">{t("details.approvals.none")}</Text>
                )}
        </div>
      </div>
    </Card>
  );
}

function DocumentApprovalDecisions({
  quorumKey,
}: {
  quorumKey: DocumentApprovalsSectionDecisions_quorum$key;
}) {
  const { t } = useTranslation("organizations/documents");
  const {
    data: quorum,
    loadNext,
    hasNext,
    isLoadingNext,
  } = usePaginationFragment<
    DocumentApprovalsSectionDecisionsRefetchQuery,
    DocumentApprovalsSectionDecisions_quorum$key
  >(documentApprovalsDecisionsFragment, quorumKey);
  const { row, name, trailing } = documentApprovals();
  const decisions = quorum.decisions.edges.map(edge => edge.node);

  return (
    <>
      {decisions.length > 0
        ? decisions.map(decision => (
            <div key={decision.id} className={row()}>
              <Text size={2} className={name()} title={decision.approver.fullName}>
                {decision.approver.fullName}
              </Text>
              <div className={trailing()}>
                <Badge size={1} variant="soft" color={stateColor(decision.state)}>
                  {t(`details.approvals.states.${decision.state}`)}
                </Badge>
              </div>
            </div>
          ))
        : (
            <Text size={1} color="faint">{t("details.approvals.none")}</Text>
          )}
      {hasNext && (
        <Button
          variant="ghost"
          color="neutral"
          size={1}
          className="self-start"
          loading={isLoadingNext}
          onClick={() => loadNext(documentApprovalsPageSize)}
        >
          {t("details.approvals.loadMore")}
        </Button>
      )}
    </>
  );
}
