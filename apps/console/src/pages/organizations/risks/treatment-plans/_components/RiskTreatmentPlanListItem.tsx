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

import {
  ActionDropdown,
  DropdownItem,
  IconCircleCheck,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";
import { useParams } from "react-router";

import type { RiskTreatmentPlanListItem_treatmentPlan$key } from "#/__generated__/core/RiskTreatmentPlanListItem_treatmentPlan.graphql";
import type { RiskTreatmentPlanListItemDeleteMutation } from "#/__generated__/core/RiskTreatmentPlanListItemDeleteMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { TreatmentPlanMeasuresDialog } from "../../risk-analyses/treatment-plans/_components/TreatmentPlanMeasuresDialog";
import { TreatmentPlanScoreTags } from "../../risk-analyses/treatment-plans/_components/TreatmentPlanScoreTags";
import { UpdateTreatmentPlanDialog } from "../../risk-analyses/treatment-plans/_components/UpdateTreatmentPlanDialog";

export const riskTreatmentPlanListItemFragment = graphql`
  fragment RiskTreatmentPlanListItem_treatmentPlan on TreatmentPlan {
    id
    treatment
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    owner {
      fullName
    }
    riskAnalysis {
      id
      name
      matrixSize {
        rows
        cols
      }
    }
    canUpdate: permission(action: "risk-management:treatment-plan:update")
    canDelete: permission(action: "risk-management:treatment-plan:delete")
    ...UpdateTreatmentPlanDialog_treatmentPlan
  }
`;

const deleteMutation = graphql`
  mutation RiskTreatmentPlanListItemDeleteMutation(
    $input: DeleteTreatmentPlanInput!
    $connections: [ID!]!
  ) {
    deleteTreatmentPlan(input: $input) {
      deletedTreatmentPlanId @deleteEdge(connections: $connections)
    }
  }
`;

interface RiskTreatmentPlanListItemProps {
  treatmentPlanKey: RiskTreatmentPlanListItem_treatmentPlan$key;
  connectionId: string;
}

export function RiskTreatmentPlanListItem({
  treatmentPlanKey,
  connectionId,
}: RiskTreatmentPlanListItemProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { riskId } = useParams<{ riskId: string }>();
  const confirm = useConfirm();
  const formDialogRef = useDialogRef();
  const measuresDialogRef = useDialogRef();
  const treatmentPlan = useFragment(riskTreatmentPlanListItemFragment, treatmentPlanKey);
  const relayEnv = useRelayEnvironment();
  const [deleteTreatmentPlan]
    = useMutation<RiskTreatmentPlanListItemDeleteMutation>(deleteMutation);

  const onDelete = () => {
    confirm(
      async () => {
        await deleteTreatmentPlan({
          variables: {
            input: { treatmentPlanId: treatmentPlan.id },
            connections: [connectionId],
          },
        });
        if (riskId) {
          updateStoreCounter(relayEnv, riskId, "treatmentPlans(first:0)", -1);
        }
      },
      {
        message: t("riskTreatmentPlanListItem.deleteConfirmation", {
          name: treatmentPlan.riskAnalysis.name,
        }),
      },
    );
  };

  const analysisMatrix = treatmentPlan.riskAnalysis.matrixSize;
  const matrixSize = analysisMatrix
    ? { rows: analysisMatrix.rows, cols: analysisMatrix.cols }
    : null;

  return (
    <>
      {matrixSize && (
        <UpdateTreatmentPlanDialog
          dialogRef={formDialogRef}
          treatmentPlanKey={treatmentPlan}
          matrixSize={matrixSize}
        />
      )}
      <TreatmentPlanMeasuresDialog
        dialogRef={measuresDialogRef}
        treatmentPlanId={treatmentPlan.id}
      />
      <Tr to={`/organizations/${organizationId}/risk-management/risk-analyses/${treatmentPlan.riskAnalysis.id}/treatment-plans`}>
        <Td className="font-medium">{treatmentPlan.riskAnalysis.name}</Td>
        <Td>
          {t(`formRiskDialog.treatments.${treatmentPlan.treatment.toLowerCase()}`)}
        </Td>
        <Td>
          {treatmentPlan.owner.fullName}
        </Td>
        <Td>
          <TreatmentPlanScoreTags
            inherentLikelihood={treatmentPlan.inherentLikelihood}
            inherentImpact={treatmentPlan.inherentImpact}
            residualLikelihood={treatmentPlan.residualLikelihood}
            residualImpact={treatmentPlan.residualImpact}
            matrixSize={matrixSize ?? { rows: 5, cols: 5 }}
          />
        </Td>
        <Td noLink width={50} className="text-end">
          {(treatmentPlan.treatment !== "ACCEPTED" || treatmentPlan.canUpdate || treatmentPlan.canDelete) && (
            <ActionDropdown>
              {treatmentPlan.treatment !== "ACCEPTED" && (
                <DropdownItem
                  icon={IconCircleCheck}
                  onClick={() => measuresDialogRef.current?.open()}
                >
                  {t("riskTreatmentPlanListItem.actions.measures")}
                </DropdownItem>
              )}
              {treatmentPlan.canUpdate && matrixSize && (
                <DropdownItem
                  icon={IconPencil}
                  onClick={() => formDialogRef.current?.open()}
                >
                  {t("riskTreatmentPlanListItem.actions.edit")}
                </DropdownItem>
              )}
              {treatmentPlan.canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  onClick={onDelete}
                >
                  {t("riskTreatmentPlanListItem.actions.delete")}
                </DropdownItem>
              )}
            </ActionDropdown>
          )}
        </Td>
      </Tr>
    </>
  );
}
