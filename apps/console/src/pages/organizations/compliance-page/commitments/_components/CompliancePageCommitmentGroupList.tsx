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

import { Button, Card, IconPlusLarge } from "@probo/ui";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { useRefetchableFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageCommitmentGroupListFragment$key } from "#/__generated__/core/CompliancePageCommitmentGroupListFragment.graphql";
import type { CompliancePageCommitmentGroupListRefetchQuery } from "#/__generated__/core/CompliancePageCommitmentGroupListRefetchQuery.graphql";
import type { CompliancePageCommitmentGroupListUpdateRankMutation } from "#/__generated__/core/CompliancePageCommitmentGroupListUpdateRankMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CompliancePageCommitmentGroupDialog, type CompliancePageCommitmentGroupDialogRef } from "./CompliancePageCommitmentGroupDialog";
import { CompliancePageCommitmentGroupListItem } from "./CompliancePageCommitmentGroupListItem";

const fragment = graphql`
  fragment CompliancePageCommitmentGroupListFragment on CompliancePortal
  @refetchable(queryName: "CompliancePageCommitmentGroupListRefetchQuery") {
    commitmentGroups(first: 100, orderBy: { field: RANK, direction: ASC }) {
      edges {
        node {
          id
          rank
          ...CompliancePageCommitmentGroupListItemFragment
        }
      }
    }
  }
`;

const updateRankMutation = graphql`
  mutation CompliancePageCommitmentGroupListUpdateRankMutation(
    $input: UpdateCompliancePortalCommitmentGroupInput!
  ) {
    updateCompliancePortalCommitmentGroup(input: $input) {
      compliancePortalCommitmentGroup {
        id
        rank
      }
    }
  }
`;

export function CompliancePageCommitmentGroupList(props: {
  fragmentRef: CompliancePageCommitmentGroupListFragment$key;
  compliancePortalId: string;
  canCreate: boolean;
}) {
  const { fragmentRef, compliancePortalId, canCreate } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const dialogRef = useRef<CompliancePageCommitmentGroupDialogRef>(null);

  const [data, refetch] = useRefetchableFragment<
    CompliancePageCommitmentGroupListRefetchQuery,
    CompliancePageCommitmentGroupListFragment$key
  >(fragment, fragmentRef);

  const [updateRank, isReordering] = useMutationWithToasts<CompliancePageCommitmentGroupListUpdateRankMutation>(
    updateRankMutation,
    { successMessage: t("commitmentGroupList.messages.orderUpdated"), errorMessage: t("commitmentGroupList.errors.updateOrder") },
  );

  const onChanged = () => refetch({}, { fetchPolicy: "network-only" });

  const groups = data.commitmentGroups.edges.map(edge => edge.node);

  const moveGroup = async (index: number, direction: "up" | "down") => {
    const target = direction === "up" ? groups[index - 1] : groups[index + 1];
    if (!target) {
      return;
    }

    await updateRank({
      variables: { input: { id: groups[index].id, rank: target.rank } },
      onSuccess: onChanged,
    });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-base font-medium">
            {t("commitmentGroupList.title")}
          </h2>
          <p className="text-sm text-txt-tertiary">
            {t("commitmentGroupList.description")}
          </p>
        </div>
        {canCreate && (
          <Button icon={IconPlusLarge} onClick={() => dialogRef.current?.openCreate(compliancePortalId)}>
            {t("commitmentGroupList.actions.addGroup")}
          </Button>
        )}
      </div>

      {groups.length === 0
        ? (
            <Card className="p-6 text-center text-sm text-txt-secondary">
              {t("commitmentGroupList.empty")}
            </Card>
          )
        : (
            <div className="space-y-6">
              {groups.map((group, index) => (
                <CompliancePageCommitmentGroupListItem
                  key={group.id}
                  fragmentRef={group}
                  onEdit={g => dialogRef.current?.openEdit(g)}
                  onChanged={onChanged}
                  isFirst={index === 0}
                  isLast={index === groups.length - 1}
                  isReordering={isReordering}
                  onMoveUp={() => void moveGroup(index, "up")}
                  onMoveDown={() => void moveGroup(index, "down")}
                />
              ))}
            </div>
          )}

      <CompliancePageCommitmentGroupDialog ref={dialogRef} onChanged={onChanged} />
    </div>
  );
}
