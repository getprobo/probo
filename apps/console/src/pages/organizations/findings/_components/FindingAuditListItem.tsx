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

import { getAuditStateVariant } from "@probo/helpers";
import {
  Badge,
  Button,
  IconTrashCan,
  Td,
  Tr,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { FindingAuditListItem_auditEdge$key } from "#/__generated__/core/FindingAuditListItem_auditEdge.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const findingAuditListItemFragment = graphql`
  fragment FindingAuditListItem_auditEdge on AuditEdge {
    referenceId
    node {
      id
      name
      state
      framework {
        name
      }
    }
  }
`;

interface FindingAuditListItemProps {
  auditEdgeKey: FindingAuditListItem_auditEdge$key;
  canUnlink: boolean;
  onUnlink: (auditId: string) => void;
}

export function FindingAuditListItem({
  auditEdgeKey,
  canUnlink,
  onUnlink,
}: FindingAuditListItemProps) {
  const auditEdge = useFragment(findingAuditListItemFragment, auditEdgeKey);
  const organizationId = useOrganizationId();
  const { t } = useTranslation();
  const audit = auditEdge.node;

  return (
    <Tr to={`/organizations/${organizationId}/governance/audits/${audit.id}`}>
      <Td>
        <div className="flex flex-col">
          <div className="font-medium">{audit.framework?.name}</div>
          {audit.name && (
            <div className="text-sm text-txt-secondary">{audit.name}</div>
          )}
        </div>
      </Td>
      <Td>{auditEdge.referenceId}</Td>
      <Td>
        <Badge color={getAuditStateVariant(audit.state)}>
          {audit.state.replace(/_/g, " ")}
        </Badge>
      </Td>
      {canUnlink && (
        <Td noLink width={50} className="text-end">
          <Button
            variant="secondary"
            icon={IconTrashCan}
            onClick={() => onUnlink(audit.id)}
          >
            {t("linkedAuditsCard.actions.unlink")}
          </Button>
        </Td>
      )}
    </Tr>
  );
}
