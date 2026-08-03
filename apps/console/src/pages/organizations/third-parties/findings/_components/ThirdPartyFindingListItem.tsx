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

import { getStatusVariant } from "@probo/helpers";
import { Badge, Td, Tr } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";
import { Link } from "react-router";

import type { ThirdPartyFindingListItem_finding$key } from "#/__generated__/core/ThirdPartyFindingListItem_finding.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const findingListItemFragment = graphql`
  fragment ThirdPartyFindingListItem_finding on Finding {
    id
    referenceId
    description
    status
    priority
  }
`;

export function ThirdPartyFindingListItem({
  findingKey,
}: {
  findingKey: ThirdPartyFindingListItem_finding$key;
}) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const finding = useFragment(findingListItemFragment, findingKey);

  return (
    <Tr>
      <Td>
        <Link
          to={`/organizations/${organizationId}/findings/${finding.id}`}
          className="font-mono underline"
        >
          {finding.referenceId}
        </Link>
      </Td>
      <Td>{finding.description ?? t("thirdPartyFindingsPage.noDescription")}</Td>
      <Td>
        <Badge variant={getStatusVariant(finding.status)}>
          {t(`findingsPage.status.${finding.status.toLowerCase()}`)}
        </Badge>
      </Td>
      <Td>{t(`findingsPage.priority.${finding.priority.toLowerCase()}`)}</Td>
    </Tr>
  );
}
