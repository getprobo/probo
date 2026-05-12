// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  IconCircleCheck,
  IconCircleX,
  IconClock,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { DocumentApprovalListItemFragment$key } from "#/__generated__/core/DocumentApprovalListItemFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
  fragment DocumentApprovalListItemFragment on DocumentVersionApprovalDecision {
    id
    approver {
      fullName
    }
    state
    comment
    decidedAt
    createdAt
    canApprove: permission(action: "core:document-version:approve")
    canReject: permission(action: "core:document-version:reject")
    documentVersion {
      id
      document {
        id
      }
    }
  }
`;

export function DocumentApprovalListItem(props: {
  fragmentRef: DocumentApprovalListItemFragment$key;
}) {
  const { fragmentRef } = props;
  const { __, dateTimeFormat } = useTranslate();
  const organizationId = useOrganizationId();

  const decision = useFragment(fragment, fragmentRef);

  const isPending = decision.state === "PENDING";
  const isApproved = decision.state === "APPROVED";
  const isRejected = decision.state === "REJECTED";
  const isVoided = decision.state === "VOIDED";

  const reviewUrl = `/organizations/${organizationId}/employee/approvals/${decision.documentVersion.document.id}`;

  return (
    <div className="flex gap-3 items-center py-3">
      <div className="space-y-1">
        <div className="text-sm text-txt-primary font-medium">
          {decision.approver.fullName}
        </div>
        <div className="text-xs text-txt-secondary flex items-center gap-1">
          {isApproved && <IconCircleCheck size={16} className="text-txt-accent" />}
          {isRejected && <IconCircleX size={16} className="text-txt-danger" />}
          {isPending && <IconClock size={16} />}
          {isVoided && <IconClock size={16} className="text-txt-secondary" />}
          <span>
            {isPending && sprintf(__("Requested on %s"), dateTimeFormat(decision.createdAt))}
            {isApproved && sprintf(__("Approved on %s"), dateTimeFormat(decision.decidedAt))}
            {isRejected && sprintf(__("Rejected on %s"), dateTimeFormat(decision.decidedAt))}
            {isVoided && sprintf(__("Requested on %s"), dateTimeFormat(decision.createdAt))}
          </span>
        </div>
        {decision.comment && (
          <div className="text-xs text-txt-secondary italic">
            {decision.comment}
          </div>
        )}
      </div>
      <div className="ml-auto flex items-center gap-2">
        {isApproved && (
          <Badge variant="success">{__("Approved")}</Badge>
        )}
        {isRejected && (
          <Badge variant="danger">{__("Rejected")}</Badge>
        )}
        {isPending && (decision.canApprove || decision.canReject) && (
          <Button variant="secondary" to={reviewUrl} target="_blank">
            {__("Review")}
          </Button>
        )}
        {isPending && !decision.canApprove && !decision.canReject && (
          <Badge variant="warning">{__("Pending")}</Badge>
        )}
        {isVoided && (
          <Badge variant="neutral">{__("Voided")}</Badge>
        )}
      </div>
    </div>
  );
}
