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
import { Card } from "@probo/ui/src/v2/Card/Card";
import { Dropdown } from "@probo/ui/src/v2/Dropdown/Dropdown";
import { DropdownPopup } from "@probo/ui/src/v2/Dropdown/DropdownPopup";
import { DropdownTrigger } from "@probo/ui/src/v2/Dropdown/DropdownTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentApprovalsSection_document$data } from "#/__generated__/core/DocumentApprovalsSection_document.graphql";
import type { DocumentApprovalsSection_document$key } from "#/__generated__/core/DocumentApprovalsSection_document.graphql";
import { PeopleMenu } from "#/components/people/PeopleMenu";

import { useUpdateDocument } from "../_lib/useUpdateDocument";
import { documentApprovals } from "../variants";

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
                decisions(first: 100, orderBy: { field: CREATED_AT, direction: ASC }) {
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
            }
          }
        }
      }
    }
  }
`;

type DecisionState = DocumentApprovalsSection_document$data[
  "pendingApproval"
]["edges"][number]["node"]["approvalQuorums"]["edges"][number]["node"]["decisions"]["edges"][number]["node"]["state"];

interface ApprovalRow {
  id: string;
  name: string;
  state: DecisionState | null;
}

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
  const { root, title, titleLead, titleIcon, rows, row, name, trailing } = documentApprovals();
  const quorum = document.pendingApproval.edges[0]?.node.approvalQuorums.edges[0]?.node ?? null;
  const pending = quorum?.status === "PENDING";
  const canEdit = document.canUpdate && document.status !== "ARCHIVED" && !pending;
  const decisions = quorum?.decisions.edges.map(edge => edge.node) ?? [];
  const savedApprovers = document.defaultApprovers;
  const approverIds = savedApprovers.map(person => person.id);
  const people: ApprovalRow[] = pending
    ? decisions.map(decision => ({
        id: decision.id,
        name: decision.approver.fullName,
        state: decision.state,
      }))
    : savedApprovers.map(person => ({
        id: person.id,
        name: person.fullName,
        state: null,
      }));

  function save(defaultApproverIds: string[]) {
    void updateDocument({
      variables: {
        input: {
          id: document.id,
          defaultApproverIds,
        },
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
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
                    save([...approverIds, approverId]);
                  }}
                />
              </DropdownPopup>
            </Dropdown>
          )}
        </div>
        <div className={rows()}>
          {people.length > 0
            ? people.map(person => (
                <div key={person.id} className={row()}>
                  <Text size={2} className={name()} title={person.name}>
                    {person.name}
                  </Text>
                  <div className={trailing()}>
                    {person.state != null && (
                      <Badge size={1} variant="soft" color={stateColor(person.state)}>
                        {t(`details.approvals.states.${person.state}`)}
                      </Badge>
                    )}
                    {canEdit && (
                      <IconButton
                        size={1}
                        variant="ghost"
                        color="neutral"
                        disabled={isUpdating}
                        aria-label={t("details.approvals.remove", { name: person.name })}
                        onClick={() => {
                          save(approverIds.filter(id => id !== person.id));
                        }}
                      >
                        <XIcon />
                      </IconButton>
                    )}
                  </div>
                </div>
              ))
            : (
                <Text size={1} color="faint">{t("details.approvals.none")}</Text>
              )}
        </div>
      </div>
    </Card>
  );
}
