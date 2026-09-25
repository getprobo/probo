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

import { GitForkIcon } from "@phosphor-icons/react";
import {
  ActionDropdown,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import type { ComponentProps } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskAnalysisActions_riskAnalysis$key } from "#/__generated__/core/RiskAnalysisActions_riskAnalysis.graphql";
import type { RiskAnalysisActionsDeleteMutation } from "#/__generated__/core/RiskAnalysisActionsDeleteMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { ForkRiskAnalysisDialog } from "./ForkRiskAnalysisDialog";
import { UpdateRiskAnalysisDialog } from "./UpdateRiskAnalysisDialog";

const riskAnalysisActionsFragment = graphql`
  fragment RiskAnalysisActions_riskAnalysis on RiskAnalysis {
    id
    canUpdate: permission(action: "risk-management:risk-analysis:update")
    canDelete: permission(action: "risk-management:risk-analysis:delete")
    organization {
      id
      canCreateRiskAnalysis: permission(action: "risk-management:risk-analysis:create")
    }
    ...ForkRiskAnalysisDialog_riskAnalysis
    ...UpdateRiskAnalysisDialog_riskAnalysis
  }
`;

const deleteMutation = graphql`
  mutation RiskAnalysisActionsDeleteMutation(
    $input: DeleteRiskAnalysisInput!
    $connections: [ID!]!
  ) {
    deleteRiskAnalysis(input: $input) {
      deletedRiskAnalysisId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskAnalysisActionsProps {
  riskAnalysisKey: RiskAnalysisActions_riskAnalysis$key;
  connectionId: string;
  variant?: ComponentProps<typeof ActionDropdown>["variant"];
  onDeleted?: () => void;
}

export function RiskAnalysisActions({
  riskAnalysisKey,
  connectionId,
  variant,
  onDeleted,
}: RiskAnalysisActionsProps) {
  const { t } = useTranslation();
  const confirm = useConfirm();
  const forkDialogRef = useDialogRef();
  const editDialogRef = useDialogRef();
  const riskAnalysis = useFragment(riskAnalysisActionsFragment, riskAnalysisKey);
  const [deleteRiskAnalysis] = useMutation<RiskAnalysisActionsDeleteMutation>(
    deleteMutation,
    { errorToast: t("riskAnalysisDetailPage.errors.delete") },
  );
  const canFork = riskAnalysis.organization?.canCreateRiskAnalysis ?? false;

  if (!canFork && !riskAnalysis.canUpdate && !riskAnalysis.canDelete) {
    return null;
  }

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteRiskAnalysis({
          variables: {
            input: { riskAnalysisId: riskAnalysis.id },
            connections: [connectionId],
          },
        });
        onDeleted?.();
      },
      { message: t("riskAnalysisDetailPage.deleteConfirmation") },
    );
  };

  return (
    <>
      {canFork && (
        <ForkRiskAnalysisDialog
          riskAnalysisKey={riskAnalysis}
          connectionId={connectionId}
          dialogRef={forkDialogRef}
        />
      )}
      {riskAnalysis.canUpdate && (
        <UpdateRiskAnalysisDialog
          riskAnalysisKey={riskAnalysis}
          dialogRef={editDialogRef}
        />
      )}
      <ActionDropdown variant={variant}>
        {canFork && (
          <DropdownItem
            icon={GitForkIcon}
            onSelect={() => forkDialogRef.current?.open()}
          >
            {t("riskAnalysesPage.actions.fork")}
          </DropdownItem>
        )}
        {riskAnalysis.canUpdate && (
          <DropdownItem
            icon={IconPencil}
            onSelect={() => editDialogRef.current?.open()}
          >
            {t("riskAnalysisDetailPage.actions.edit")}
          </DropdownItem>
        )}
        {riskAnalysis.canDelete && (
          <DropdownItem
            variant="danger"
            icon={IconTrashCan}
            onSelect={handleDelete}
          >
            {t("riskAnalysisDetailPage.actions.delete")}
          </DropdownItem>
        )}
      </ActionDropdown>
    </>
  );
}
