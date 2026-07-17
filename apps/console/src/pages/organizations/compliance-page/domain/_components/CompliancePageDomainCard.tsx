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
    id
    domain
    managed
    sslStatus
    provisioningError
    canDelete: permission(action: "compliance-portal:custom-domain:delete")
    ...CompliancePageDomainDialogFragment
  }
`;

export function CompliancePageDomainCard(props: {
  fKey: CompliancePageDomainCardFragment$key;
}) {
  const { fKey } = props;

  const { __ } = useTranslate();

  const domain = useFragment<CompliancePageDomainCardFragment$key>(fragment, fKey);

  return (
    <Card padded>
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-medium">{domain.domain}</span>
            {domain.managed && (
              <Badge variant="neutral">{__("Managed")}</Badge>
            )}
            <Badge variant={getCustomDomainStatusBadgeVariant(domain.sslStatus)}>
              {getCustomDomainStatusBadgeLabel(domain.sslStatus, __)}
            </Badge>
          </div>
          <p className="text-sm text-txt-secondary">
            {domain.sslStatus === "ACTIVE"
              ? __("Verified and serving traffic")
              : domain.provisioningError
                ? domain.provisioningError
                : __("Pending DNS verification")}
          </p>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <CompliancePageDomainDialog fKey={domain}>
            <Button variant="secondary">{__("View details")}</Button>
          </CompliancePageDomainDialog>

          {domain.canDelete && (
            <DeleteCompliancePageDomainDialog
              domain={domain.domain}
              customDomainId={domain.id}
            >
              <Button variant="danger">{__("Delete")}</Button>
            </DeleteCompliancePageDomainDialog>
          )}
        </div>
      </div>
    </Card>
  );
}
