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
  getCertificateProvisioningErrorMessage,
  getCustomDomainStatusBadgeLabel,
} from "@probo/helpers";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDomainCardFragment$key } from "#/__generated__/core/CompliancePortalDomainCardFragment.graphql";

import { customDomainBadgeColor } from "../_lib/customDomainBadgeColor";
import { domainCard } from "../variants";

import { CompliancePortalDomainDialog } from "./CompliancePortalDomainDialog";
import { DeleteCompliancePortalDomainDialog } from "./DeleteCompliancePortalDomainDialog";

const fragment = graphql`
  fragment CompliancePortalDomainCardFragment on CustomDomain {
    domain
    managed
    certificate {
      status
      provisioningError
    }
    canDelete: permission(action: "compliance-portal:custom-domain:delete")
    ...CompliancePortalDomainDialogFragment
    ...DeleteCompliancePortalDomainDialog_customDomain
  }
`;

interface CompliancePortalDomainCardProps {
  customDomainKey: CompliancePortalDomainCardFragment$key;
}

export function CompliancePortalDomainCard({ customDomainKey }: CompliancePortalDomainCardProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { identity, actions } = domainCard();

  const domain = useFragment<CompliancePortalDomainCardFragment$key>(fragment, customDomainKey);
  const sslStatus = domain.certificate?.status ?? "PENDING";
  const provisioningErrorMessage = getCertificateProvisioningErrorMessage(
    domain.certificate?.provisioningError,
    t,
  );

  return (
    <ListItem>
      <ListItemContent>
        <div className={identity()}>
          <Text size={2} weight="medium" highContrast>{domain.domain}</Text>
          {domain.managed && (
            <Badge variant="soft" color="neutral">{t("domainCard.managed")}</Badge>
          )}
          <Badge variant="soft" color={customDomainBadgeColor(sslStatus)}>
            {getCustomDomainStatusBadgeLabel(sslStatus, t)}
          </Badge>
        </div>
        <Text size={1} color="neutral">
          {sslStatus === "ACTIVE"
            ? t("domainCard.status.active")
            : provisioningErrorMessage
              ? provisioningErrorMessage
              : t("domainCard.status.pending")}
        </Text>
      </ListItemContent>

      <div className={actions()}>
        <CompliancePortalDomainDialog customDomainKey={domain}>
          <Button variant="soft" color="neutral">
            {t("domainCard.actions.viewDetails")}
          </Button>
        </CompliancePortalDomainDialog>

        {domain.canDelete && (
          <DeleteCompliancePortalDomainDialog customDomainKey={domain}>
            <Button variant="solid" color="red">
              {t("domainCard.actions.delete")}
            </Button>
          </DeleteCompliancePortalDomainDialog>
        )}
      </div>
    </ListItem>
  );
}
