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
