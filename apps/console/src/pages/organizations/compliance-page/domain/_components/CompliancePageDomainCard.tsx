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

import {
  getCustomDomainStatusBadgeLabel,
  getCustomDomainStatusBadgeVariant,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Button, Card } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainCardFragment$key } from "#/__generated__/core/CompliancePageDomainCardFragment.graphql";

import { CompliancePageDomainDialog } from "./CompliancePageDomainDialog";
import { DeleteCompliancePageDomainDialog } from "./DeleteCompliancePageDomainDialog";

const fragment = graphql`
  fragment CompliancePageDomainCardFragment on CustomDomain {
    domain
    sslStatus
    provisioningError
    canDelete: permission(action: "core:custom-domain:delete")
    ...CompliancePageDomainDialogFragment
  }
`;

export function CompliancePageDomainCard(props: { fKey: CompliancePageDomainCardFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const domain = useFragment<CompliancePageDomainCardFragment$key>(fragment, fKey);

  return (
    <Card>
      <div className="p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div>
              <div className="font-medium mb-1">{domain.domain}</div>
              <div className="text-sm text-txt-secondary">
                {domain.sslStatus === "ACTIVE"
                  ? __("Verified")
                  : domain.provisioningError
                    ? domain.provisioningError
                    : __("Pending verification")}
              </div>
            </div>
            <Badge
              variant={getCustomDomainStatusBadgeVariant(domain.sslStatus)}
            >
              {getCustomDomainStatusBadgeLabel(domain.sslStatus, __)}
            </Badge>
          </div>

          <div className="flex items-center gap-2">
            <CompliancePageDomainDialog fKey={domain}>
              <Button variant="secondary">{__("View Details")}</Button>
            </CompliancePageDomainDialog>

            {domain.canDelete && (
              <DeleteCompliancePageDomainDialog domain={domain.domain}>
                <Button variant="danger">{__("Delete")}</Button>
              </DeleteCompliancePageDomainDialog>
            )}
          </div>
        </div>
      </div>
    </Card>
  );
}
