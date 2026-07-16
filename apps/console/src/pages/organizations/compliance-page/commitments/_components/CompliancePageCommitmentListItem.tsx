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

import { useTranslate } from "@probo/i18n";
import { Badge, Button, IconPencil, IconTrashCan, Spinner, Td, Tr } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageCommitmentListItemDeleteMutation } from "#/__generated__/core/CompliancePageCommitmentListItemDeleteMutation.graphql";
import type { CompliancePageCommitmentListItemFragment$data, CompliancePageCommitmentListItemFragment$key } from "#/__generated__/core/CompliancePageCommitmentListItemFragment.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { COMMITMENT_ICON_LABELS } from "../_lib/commitmentIcons";

const deleteCommitmentMutation = graphql`
  mutation CompliancePageCommitmentListItemDeleteMutation(
    $input: DeleteCompliancePortalCommitmentInput!
  ) {
    deleteCompliancePortalCommitment(input: $input) {
      deletedCompliancePortalCommitmentId
    }
  }
`;

const fragment = graphql`
  fragment CompliancePageCommitmentListItemFragment on CompliancePortalCommitment {
    id
    icon
    eyebrow
    title
    description
    canUpdate: permission(action: "core:compliance-portal-commitment:update")
    canDelete: permission(action: "core:compliance-portal-commitment:delete")
  }
`;

export function CompliancePageCommitmentListItem(props: {
  fragmentRef: CompliancePageCommitmentListItemFragment$key;
  onEdit: (commitment: CompliancePageCommitmentListItemFragment$data) => void;
  onChanged: () => void;
}) {
  const { fragmentRef, onEdit, onChanged } = props;

  const { __ } = useTranslate();
  const commitment = useFragment<CompliancePageCommitmentListItemFragment$key>(fragment, fragmentRef);

  const [deleteCommitment, isDeleting] = useMutationWithToasts<CompliancePageCommitmentListItemDeleteMutation>(
    deleteCommitmentMutation,
    { successMessage: __("Commitment deleted successfully"), errorMessage: __("Failed to delete commitment") },
  );

  const handleDelete = async () => {
    await deleteCommitment({
      variables: { input: { id: commitment.id } },
      onSuccess: onChanged,
    });
  };

  const iconLabel = COMMITMENT_ICON_LABELS[commitment.icon];

  return (
    <Tr>
      <Td>
        <Badge variant="neutral">{iconLabel}</Badge>
      </Td>
      <Td>
        <div className="flex flex-col">
          {commitment.eyebrow && (
            <span className="text-xs text-txt-tertiary">{commitment.eyebrow}</span>
          )}
          <span className="font-medium">{commitment.title}</span>
        </div>
      </Td>
      <Td>
        <span className="text-txt-secondary line-clamp-2">{commitment.description}</span>
      </Td>
      <Td noLink width={120} className="text-end">
        <div className="flex gap-2 justify-end">
          {commitment.canUpdate && (
            <Button variant="secondary" icon={IconPencil} onClick={() => onEdit(commitment)} />
          )}
          {commitment.canDelete && (
            <Button
              variant="danger"
              icon={isDeleting ? Spinner : IconTrashCan}
              disabled={isDeleting}
              onClick={() => void handleDelete()}
            />
          )}
        </div>
      </Td>
    </Tr>
  );
}
