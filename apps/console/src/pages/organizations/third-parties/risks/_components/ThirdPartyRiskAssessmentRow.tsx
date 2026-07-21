// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { formatDate } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Markdown,
  RiskBadge,
  Td,
  Tr,
} from "@probo/ui";
import { clsx } from "clsx";
import { graphql, useFragment } from "react-relay";

import type { ThirdPartyRiskAssessmentRow_assessment$key } from "#/__generated__/core/ThirdPartyRiskAssessmentRow_assessment.graphql";

const riskAssessmentRowFragment = graphql`
  fragment ThirdPartyRiskAssessmentRow_assessment on ThirdPartyRiskAssessment {
    id
    createdAt
    expiresAt
    dataSensitivity
    businessImpact
    notes
  }
`;

interface ThirdPartyRiskAssessmentRowProps {
  assessmentKey: ThirdPartyRiskAssessmentRow_assessment$key;
  isExpanded: boolean;
  onClick: (id: string) => void;
}

export function ThirdPartyRiskAssessmentRow(props: ThirdPartyRiskAssessmentRowProps) {
  const { __ } = useTranslate();
  const assessment = useFragment(riskAssessmentRowFragment, props.assessmentKey);
  const { relativeDateFormat } = useTranslate();
  const isExpired = new Date(assessment.expiresAt) < new Date();

  return (
    <>
      <Tr
        className={clsx(
          isExpired && "opacity-50",
          "cursor-pointer",
          props.isExpanded && "border-none",
        )}
        onClick={() => props.onClick(assessment.id)}
      >
        <Td>
          <span className="text-xs text-txt-secondary ml-1">
            {formatDate(assessment.createdAt)}
          </span>
        </Td>
        <Td>
          <div className="flex items-center gap-2">
            {relativeDateFormat(assessment.expiresAt)}
            {isExpired && <Badge variant="neutral">{__("Expired")}</Badge>}
          </div>
        </Td>
        <Td>
          <RiskBadge level={assessment.dataSensitivity} />
        </Td>
        <Td>
          <RiskBadge level={assessment.businessImpact} />
        </Td>
      </Tr>
      {props.isExpanded && (
        <Tr className={clsx("border-none", isExpired && "opacity-50")}>
          <Td colSpan={4} className="whitespace-normal align-top">
            <div className="space-y-2 max-w-4xl">
              <div>
                {__("Notes")}
                :
              </div>
              {assessment.notes
                ? (
                    <div className="overflow-x-auto">
                      <div className="prose prose-sm max-w-none [&_.prose]:max-w-none">
                        <Markdown content={assessment.notes} />
                      </div>
                    </div>
                  )
                : null}
            </div>
          </Td>
        </Tr>
      )}
    </>
  );
}
