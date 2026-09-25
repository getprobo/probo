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

import { Td, Tr } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { TransferImpactAssessmentListItem_transferImpactAssessment$key } from "#/__generated__/core/TransferImpactAssessmentListItem_transferImpactAssessment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const transferImpactAssessmentListItemFragment = graphql`
  fragment TransferImpactAssessmentListItem_transferImpactAssessment on TransferImpactAssessment {
    id
    dataSubjects
    transfer
    localLawRisk
    processingActivity {
      id
      name
    }
  }
`;

interface TransferImpactAssessmentListItemProps {
  transferImpactAssessmentKey: TransferImpactAssessmentListItem_transferImpactAssessment$key;
}

export function TransferImpactAssessmentListItem({
  transferImpactAssessmentKey,
}: TransferImpactAssessmentListItemProps) {
  const organizationId = useOrganizationId();
  const tia = useFragment(
    transferImpactAssessmentListItemFragment,
    transferImpactAssessmentKey,
  );

  const activityUrl
    = `/organizations/${organizationId}/privacy/processing-activities/${tia.processingActivity.id}#tia`;

  return (
    <Tr to={activityUrl}>
      <Td>
        <span className="font-semibold">
          {tia.processingActivity.name}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.dataSubjects || "-"}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.transfer || "-"}
        </span>
      </Td>
      <Td>
        <span className="text-sm text-txt-secondary line-clamp-2">
          {tia.localLawRisk || "-"}
        </span>
      </Td>
    </Tr>
  );
}
