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

import { dateFormat } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconSquareBehindSquare2,
  Td,
  Tr,
  useDialogRef,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { RiskAnalysisListItem_riskAnalysis$key } from "#/__generated__/core/RiskAnalysisListItem_riskAnalysis.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { ForkRiskAnalysisDialog } from "./ForkRiskAnalysisDialog";
import { formatMatrixSize } from "./matrixSize";

const riskAnalysisListItemFragment = graphql`
  fragment RiskAnalysisListItem_riskAnalysis on RiskAnalysis {
    id
    name
    description
    period {
      start
      end
    }
    matrixSize {
      rows
      cols
    }
    createdAt
    organization {
      id
      canCreateRiskAnalysis: permission(action: "core:risk-analysis:create")
    }
    ...ForkRiskAnalysisDialog_riskAnalysis
  }
`;

interface RiskAnalysisListItemProps {
  riskAnalysisKey: RiskAnalysisListItem_riskAnalysis$key;
  connectionId: string;
}

export function RiskAnalysisListItem({
  riskAnalysisKey,
  connectionId,
}: RiskAnalysisListItemProps) {
  const { i18n, t } = useTranslation();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();
  const riskAnalysis = useFragment(riskAnalysisListItemFragment, riskAnalysisKey);
  const canFork = riskAnalysis.organization?.canCreateRiskAnalysis ?? false;

  return (
    <>
      {canFork && (
        <ForkRiskAnalysisDialog
          riskAnalysisKey={riskAnalysis}
          connectionId={connectionId}
          dialogRef={dialogRef}
        />
      )}
      <Tr
        to={`/organizations/${organizationId}/risk-management/risk-analyses/${riskAnalysis.id}/treatment-plans`}
      >
        <Td className="w-px whitespace-nowrap font-medium">{riskAnalysis.name}</Td>
        <Td className="text-txt-secondary max-w-md">
          {riskAnalysis.description || "—"}
        </Td>
        <Td className="w-px whitespace-nowrap text-txt-secondary">
          {riskAnalysis.period
            ? `${riskAnalysis.period.start ? dateFormat(i18n.language, riskAnalysis.period.start) : "—"} – ${riskAnalysis.period.end ? dateFormat(i18n.language, riskAnalysis.period.end) : "—"}`
            : "—"}
        </Td>
        <Td className="w-px whitespace-nowrap text-txt-secondary">
          {formatMatrixSize(riskAnalysis.matrixSize.rows, riskAnalysis.matrixSize.cols)}
        </Td>
        <Td className="w-px whitespace-nowrap text-txt-secondary">
          {dateFormat(i18n.language, riskAnalysis.createdAt)}
        </Td>
        {canFork && (
          <Td noLink width={50} className="w-px text-end">
            <div onClick={event => event.stopPropagation()}>
              <ActionDropdown>
                <DropdownItem
                  icon={IconSquareBehindSquare2}
                  onClick={() => dialogRef.current?.open()}
                >
                  {t("riskAnalysesPage.actions.fork")}
                </DropdownItem>
              </ActionDropdown>
            </div>
          </Td>
        )}
      </Tr>
    </>
  );
}
