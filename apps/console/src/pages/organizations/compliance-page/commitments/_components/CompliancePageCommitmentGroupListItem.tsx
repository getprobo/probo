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

import { Button, Card, Dialog, DialogContent, DialogFooter, IconChevronDown, IconChevronUp, IconPencil, IconPlusLarge, IconTrashCan, Spinner, Table, Tbody, Td, Th, Thead, Tr, useDialogRef } from "@probo/ui";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageCommitmentGroupListItemDeleteMutation } from "#/__generated__/core/CompliancePageCommitmentGroupListItemDeleteMutation.graphql";
import type { CompliancePageCommitmentGroupListItemFragment$data, CompliancePageCommitmentGroupListItemFragment$key } from "#/__generated__/core/CompliancePageCommitmentGroupListItemFragment.graphql";
import type { CompliancePageCommitmentGroupListItemUpdateRankMutation } from "#/__generated__/core/CompliancePageCommitmentGroupListItemUpdateRankMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CompliancePageCommitmentDialog, type CompliancePageCommitmentDialogRef } from "./CompliancePageCommitmentDialog";
import { CompliancePageCommitmentListItem } from "./CompliancePageCommitmentListItem";

const deleteGroupMutation = graphql`
  mutation CompliancePageCommitmentGroupListItemDeleteMutation(
    $input: DeleteCompliancePortalCommitmentGroupInput!
  ) {
    deleteCompliancePortalCommitmentGroup(input: $input) {
      deletedCompliancePortalCommitmentGroupId
    }
  }
`;

const updateCommitmentRankMutation = graphql`
  mutation CompliancePageCommitmentGroupListItemUpdateRankMutation(
    $input: UpdateCompliancePortalCommitmentInput!
  ) {
    updateCompliancePortalCommitment(input: $input) {
      compliancePortalCommitment {
        id
        rank
      }
    }
  }
`;

const fragment = graphql`
  fragment CompliancePageCommitmentGroupListItemFragment on CompliancePortalCommitmentGroup {
    id
    title
    description
    canUpdate: permission(action: "core:compliance-portal-commitment-group:update")
    canDelete: permission(action: "core:compliance-portal-commitment-group:delete")
    canCreateCommitment: permission(action: "core:compliance-portal-commitment:create")
    commitments(first: 100, orderBy: { field: RANK, direction: ASC }) {
      edges {
        node {
          id
          rank
          ...CompliancePageCommitmentListItemFragment
        }
      }
    }
  }
`;

export function CompliancePageCommitmentGroupListItem(props: {
  fragmentRef: CompliancePageCommitmentGroupListItemFragment$key;
  onEdit: (group: CompliancePageCommitmentGroupListItemFragment$data) => void;
  onChanged: () => void;
  isFirst: boolean;
  isLast: boolean;
  isReordering: boolean;
  onMoveUp: () => void;
  onMoveDown: () => void;
}) {
  const { fragmentRef, onEdit, onChanged, isFirst, isLast, isReordering, onMoveUp, onMoveDown } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const group = useFragment<CompliancePageCommitmentGroupListItemFragment$key>(fragment, fragmentRef);
  const commitmentDialogRef = useRef<CompliancePageCommitmentDialogRef>(null);
  const deleteDialogRef = useDialogRef();

  const [deleteGroup, isDeleting] = useMutationWithToasts<CompliancePageCommitmentGroupListItemDeleteMutation>(
    deleteGroupMutation,
    { successMessage: t("commitmentGroupListItem.messages.deleted"), errorMessage: t("commitmentGroupListItem.errors.delete") },
  );
  const [updateCommitmentRank, isReorderingCommitment] = useMutationWithToasts<
    CompliancePageCommitmentGroupListItemUpdateRankMutation
  >(
    updateCommitmentRankMutation,
    { successMessage: t("commitmentGroupListItem.messages.orderUpdated"), errorMessage: t("commitmentGroupListItem.errors.updateOrder") },
  );

  const commitments = group.commitments.edges.map(edge => edge.node);

  const moveCommitment = async (index: number, direction: "up" | "down") => {
    const target = direction === "up" ? commitments[index - 1] : commitments[index + 1];
    if (!target) {
      return;
    }

    await updateCommitmentRank({
      variables: { input: { id: commitments[index].id, rank: target.rank } },
      onSuccess: onChanged,
    });
  };

  const handleDelete = async () => {
    await deleteGroup({
      variables: { input: { id: group.id } },
      onSuccess: () => {
        deleteDialogRef.current?.close();
        onChanged();
      },
    });
  };

  return (
    <Card className="space-y-4 p-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h3 className="text-base font-medium">{group.title}</h3>
          <p className="text-sm text-txt-tertiary">{group.description}</p>
        </div>
        <div className="flex gap-2">
          {group.canUpdate && (
            <>
              <Button
                variant="secondary"
                icon={IconChevronUp}
                disabled={isFirst || isReordering}
                onClick={onMoveUp}
              />
              <Button
                variant="secondary"
                icon={IconChevronDown}
                disabled={isLast || isReordering}
                onClick={onMoveDown}
              />
              <Button variant="secondary" icon={IconPencil} onClick={() => onEdit(group)} />
            </>
          )}
          {group.canDelete && (
            <>
              <Button variant="danger" icon={IconTrashCan} onClick={() => deleteDialogRef.current?.open()} />
              <Dialog
                ref={deleteDialogRef}
                title={t("commitmentGroupListItem.delete.title")}
                className="max-w-md"
              >
                <DialogContent padded>
                  <p className="text-txt-secondary">
                    {t("commitmentGroupListItem.delete.description", {
                      title: group.title,
                    })}
                  </p>
                  <p className="text-txt-secondary mt-2">
                    {t("commitmentGroupListItem.delete.warning")}
                  </p>
                </DialogContent>
                <DialogFooter>
                  <Button
                    variant="danger"
                    onClick={() => void handleDelete()}
                    disabled={isDeleting}
                    icon={isDeleting ? Spinner : IconTrashCan}
                  >
                    {isDeleting
                      ? t("commitmentGroupListItem.actions.deleting")
                      : t("commitmentGroupListItem.actions.delete")}
                  </Button>
                </DialogFooter>
              </Dialog>
            </>
          )}
        </div>
      </div>

      <Table>
        <Thead>
          <Tr>
            <Th>{t("commitmentGroupListItem.columns.icon")}</Th>
            <Th>{t("commitmentGroupListItem.columns.title")}</Th>
            <Th>{t("commitmentGroupListItem.columns.description")}</Th>
            <Th></Th>
          </Tr>
        </Thead>
        <Tbody>
          {commitments.length === 0 && (
            <Tr>
              <Td colSpan={4} className="text-center text-txt-secondary">
                {t("commitmentGroupListItem.empty")}
              </Td>
            </Tr>
          )}
          {commitments.map((commitment, index) => (
            <CompliancePageCommitmentListItem
              key={commitment.id}
              fragmentRef={commitment}
              onEdit={c => commitmentDialogRef.current?.openEdit(c)}
              onChanged={onChanged}
              isFirst={index === 0}
              isLast={index === commitments.length - 1}
              isReordering={isReorderingCommitment}
              onMoveUp={() => void moveCommitment(index, "up")}
              onMoveDown={() => void moveCommitment(index, "down")}
            />
          ))}
        </Tbody>
      </Table>

      {group.canCreateCommitment && (
        <Button
          variant="secondary"
          icon={IconPlusLarge}
          onClick={() => commitmentDialogRef.current?.openCreate(group.id)}
        >
          {t("commitmentGroupListItem.actions.addCommitment")}
        </Button>
      )}

      <CompliancePageCommitmentDialog ref={commitmentDialogRef} onChanged={onChanged} />
    </Card>
  );
}
