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
  IconChevronDown,
  IconChevronRight,
  IconPencil,
  IconTrashCan,
  Td,
  Tr,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useRelayEnvironment } from "react-relay";
import { Link, useParams } from "react-router";

import type { TreatmentPlanListItem_treatmentPlan$key } from "#/__generated__/core/TreatmentPlanListItem_treatmentPlan.graphql";
import type { TreatmentPlanListItemDeleteMutation } from "#/__generated__/core/TreatmentPlanListItemDeleteMutation.graphql";
import { updateStoreCounter } from "#/hooks/useMutationWithIncrement";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

import { type MatrixSize } from "../../_components/matrixSize";

import { TreatmentPlanMeasureList } from "./TreatmentPlanMeasureList";
import { TreatmentPlanProgressBar } from "./TreatmentPlanProgressBar";
import { TreatmentPlanScoreTags } from "./TreatmentPlanScoreTags";
import { UpdateTreatmentPlanDialog } from "./UpdateTreatmentPlanDialog";

export const treatmentPlanListItemFragment = graphql`
  fragment TreatmentPlanListItem_treatmentPlan on TreatmentPlan
  @argumentDefinitions(asOf: { type: "Datetime", defaultValue: null }) {
    id
    treatment
    inherentLikelihood
    inherentImpact
    residualLikelihood
    residualImpact
    category
    owner {
      fullName
    }
    risk {
      id
      name
    }
    canUpdate: permission(action: "risk-management:treatment-plan:update")
    canDelete: permission(action: "risk-management:treatment-plan:delete")
    progress {
      done
      inProgress
      notImplemented
      total
    }
    ...TreatmentPlanMeasureList_meta
    ...TreatmentPlanMeasureList_treatmentPlan @arguments(asOf: $asOf)
    ...UpdateTreatmentPlanDialog_treatmentPlan
  }
`;

const deleteMutation = graphql`
  mutation TreatmentPlanListItemDeleteMutation(
    $input: DeleteTreatmentPlanInput!
    $connections: [ID!]!
  ) {
    deleteTreatmentPlan(input: $input) {
      deletedTreatmentPlanId @deleteEdge(connections: $connections)
    }
  }
`;

interface TreatmentPlanListItemProps {
  treatmentPlanKey: TreatmentPlanListItem_treatmentPlan$key;
  connectionId: string;
  matrixSize: MatrixSize;
  readOnly?: boolean;
  onChanged?: () => void;
}

export function TreatmentPlanListItem({
  treatmentPlanKey,
  connectionId,
  matrixSize,
  readOnly = false,
  onChanged,
}: TreatmentPlanListItemProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { riskAnalysisId } = useParams<{ riskAnalysisId: string }>();
  const confirm = useConfirm();
  const formDialogRef = useDialogRef();
  const [expanded, setExpanded] = useState(false);
  const treatmentPlan = useFragment(treatmentPlanListItemFragment, treatmentPlanKey);
  const relayEnv = useRelayEnvironment();
  const [deleteTreatmentPlan] = useMutation<TreatmentPlanListItemDeleteMutation>(deleteMutation);
  const progress = treatmentPlan.progress;
  const inherentLikelihood = treatmentPlan.inherentLikelihood;
  const inherentImpact = treatmentPlan.inherentImpact;
  const residualLikelihood = treatmentPlan.residualLikelihood;
  const residualImpact = treatmentPlan.residualImpact;

  const onDelete = () => {
    confirm(
      async () => {
        await deleteTreatmentPlan({
          variables: {
            input: { treatmentPlanId: treatmentPlan.id },
            connections: [connectionId],
          },
        });
        if (riskAnalysisId) {
          updateStoreCounter(relayEnv, riskAnalysisId, "treatmentPlans(first:0)", -1);
        }
        onChanged?.();
      },
      {
        message: t("treatmentPlanListItem.deleteConfirmation", {
          name: treatmentPlan.risk.name,
        }),
      },
    );
  };

  return (
    <>
      {!readOnly && (
        <UpdateTreatmentPlanDialog
          dialogRef={formDialogRef}
          treatmentPlanKey={treatmentPlan}
          matrixSize={matrixSize}
        />
      )}
      <Tr
        className="cursor-pointer"
        onClick={() => setExpanded(open => !open)}
      >
        <Td className="font-medium">
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="shrink-0 text-txt-secondary"
              aria-expanded={expanded}
              aria-label={expanded
                ? t("treatmentPlanListItem.actions.collapse")
                : t("treatmentPlanListItem.actions.expand")}
              onClick={(event) => {
                event.stopPropagation();
                setExpanded(open => !open);
              }}
            >
              {expanded
                ? <IconChevronDown size={16} />
                : <IconChevronRight size={16} />}
            </button>
            <Link
              className="hover:underline"
              to={`/organizations/${organizationId}/risk-management/risks/${treatmentPlan.risk.id}`}
              onClick={event => event.stopPropagation()}
            >
              {treatmentPlan.risk.name}
            </Link>
          </div>
        </Td>
        <Td className="w-px whitespace-nowrap pr-6">{treatmentPlan.category}</Td>
        <Td className="w-px whitespace-nowrap pr-6">
          {t(`formRiskDialog.treatments.${treatmentPlan.treatment.toLowerCase()}`)}
        </Td>
        <Td className="w-px whitespace-nowrap pr-6">
          {treatmentPlan.owner.fullName}
        </Td>
        <Td className="w-px whitespace-nowrap pr-6">
          <TreatmentPlanScoreTags
            inherentLikelihood={inherentLikelihood}
            inherentImpact={inherentImpact}
            residualLikelihood={residualLikelihood}
            residualImpact={residualImpact}
            matrixSize={matrixSize}
          />
        </Td>
        <Td className="w-px whitespace-nowrap pr-6">
          <TreatmentPlanProgressBar
            done={progress.done}
            inProgress={progress.inProgress}
            notImplemented={progress.notImplemented}
            total={progress.total}
          />
        </Td>
        <Td noLink width={48} className="w-px text-end">
          {!readOnly && (treatmentPlan.canUpdate || treatmentPlan.canDelete) && (
            <div onClick={event => event.stopPropagation()}>
              <ActionDropdown>
                {treatmentPlan.canUpdate && (
                  <DropdownItem
                    icon={IconPencil}
                    onClick={() => formDialogRef.current?.open()}
                  >
                    {t("treatmentPlanListItem.actions.edit")}
                  </DropdownItem>
                )}
                {treatmentPlan.canDelete && (
                  <DropdownItem
                    variant="danger"
                    icon={IconTrashCan}
                    onClick={onDelete}
                  >
                    {t("treatmentPlanListItem.actions.delete")}
                  </DropdownItem>
                )}
              </ActionDropdown>
            </div>
          )}
        </Td>
      </Tr>
      {expanded && (
        <Tr>
          <Td colSpan={7} className="bg-subtle !py-3">
            <div className="pl-6">
              <TreatmentPlanMeasureList
                treatmentPlanKey={treatmentPlan}
                onChanged={onChanged}
              />
            </div>
          </Td>
        </Tr>
      )}
    </>
  );
}
