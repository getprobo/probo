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
  Button,
  IconPlusLarge,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { LinkedRiskListItem_analysis$key } from "#/__generated__/core/LinkedRiskListItem_analysis.graphql";
import type { LinkedRiskListItem_risk$key } from "#/__generated__/core/LinkedRiskListItem_risk.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import type { MatrixSize } from "../../_components/matrixSize";

import { CreateTreatmentPlanDialog } from "./CreateTreatmentPlanDialog";

export const linkedRiskListItemFragment = graphql`
  fragment LinkedRiskListItem_risk on Risk {
    id
    name
    category
  }
`;

export const linkedRiskListItemAnalysisFragment = graphql`
  fragment LinkedRiskListItem_analysis on RiskAnalysis {
    canCreateTreatmentPlan: permission(action: "risk-management:treatment-plan:create")
  }
`;

interface LinkedRiskListItemProps {
  riskKey: LinkedRiskListItem_risk$key;
  analysisKey: LinkedRiskListItem_analysis$key;
  connectionId: string;
  matrixSize: MatrixSize;
  onCreated?: () => void;
}

export function LinkedRiskListItem({
  riskKey,
  analysisKey,
  connectionId,
  matrixSize,
  onCreated,
}: LinkedRiskListItemProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const risk = useFragment(linkedRiskListItemFragment, riskKey);
  const analysis = useFragment(linkedRiskListItemAnalysisFragment, analysisKey);

  return (
    <Tr to={`/organizations/${organizationId}/risk-management/risks/${risk.id}`}>
      <Td className="font-medium">
        <div className="flex items-center gap-2">
          <span className="inline-block w-4 shrink-0" />
          {risk.name}
        </div>
      </Td>
      <Td className="w-px whitespace-nowrap">{risk.category}</Td>
      <Td colSpan={4} className="text-sm text-txt-tertiary">
        {t("linkedRiskListItem.empty")}
      </Td>
      <Td noLink width={48} className="w-px text-end">
        {analysis.canCreateTreatmentPlan && (
          <CreateTreatmentPlanDialog
            connectionId={connectionId}
            matrixSize={matrixSize}
            riskId={risk.id}
            onCompleted={onCreated}
          >
            <Button icon={IconPlusLarge} variant="secondary">
              {t("linkedRiskListItem.actions.create")}
            </Button>
          </CreateTreatmentPlanDialog>
        )}
      </Td>
    </Tr>
  );
}
