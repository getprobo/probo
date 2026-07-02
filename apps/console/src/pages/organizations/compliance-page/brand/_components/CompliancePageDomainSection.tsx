// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Button, Card, IconPlusLarge } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainSection_trustCenter$key } from "#/__generated__/core/CompliancePageDomainSection_trustCenter.graphql";

import { CompliancePageDomainCard } from "../../domain/_components/CompliancePageDomainCard";
import { NewCompliancePageDomainDialog } from "../../domain/_components/NewCompliancePageDomainDialog";

const fragment = graphql`
  fragment CompliancePageDomainSection_trustCenter on TrustCenter {
    id
    canCreateCustomDomain: permission(action: "core:custom-domain:create")
    customDomain {
      ...CompliancePageDomainCardFragment
    }
  }
`;

export function CompliancePageDomainSection(props: {
  trustCenterKey: CompliancePageDomainSection_trustCenter$key;
}) {
  const { trustCenterKey } = props;
  const { __ } = useTranslate();

  const trustCenter = useFragment(fragment, trustCenterKey);

  return (
    <div className="space-y-4">
      <div>
        <h3 className="text-base font-medium">{__("Custom Domain")}</h3>
        <p className="text-sm text-txt-tertiary">
          {__("Add your own domain to make your compliance page more professional")}
        </p>
      </div>

      {trustCenter.customDomain
        ? (
            <CompliancePageDomainCard
              fKey={trustCenter.customDomain}
              trustCenterId={trustCenter.id}
            />
          )
        : (
            <Card padded>
              <div className="text-center py-8">
                <h4 className="text-lg font-semibold mb-2">
                  {__("No custom domain configured")}
                </h4>
                <p className="text-txt-tertiary mb-4">
                  {__(
                    "Connect a custom domain so visitors reach your compliance page on your own hostname.",
                  )}
                </p>
                <div className="flex justify-center">
                  {trustCenter.canCreateCustomDomain && (
                    <NewCompliancePageDomainDialog trustCenterId={trustCenter.id}>
                      <Button icon={IconPlusLarge}>{__("Add Domain")}</Button>
                    </NewCompliancePageDomainDialog>
                  )}
                </div>
              </div>
            </Card>
          )}
    </div>
  );
}
