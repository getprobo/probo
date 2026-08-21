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

import { TrashIcon } from "@phosphor-icons/react";
import { getCertificateProvisioningErrorMessage } from "@probo/helpers";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDomainCardFragment$key } from "#/__generated__/core/CompliancePortalDomainCardFragment.graphql";

import {
  customDomainCardTone,
  CustomDomainStatusIcon,
  showsCustomDomainDns,
} from "../_lib/customDomainCardTone";
import { domainCardCallout } from "../_lib/domainCardCallout";
import { domainCard, hostingCard } from "../variants";

import { CompliancePortalDomainDnsRecordListItem } from "./CompliancePortalDomainDnsRecordListItem";
import { DeleteCompliancePortalDomainDialog } from "./DeleteCompliancePortalDomainDialog";

const fragment = graphql`
  fragment CompliancePortalDomainCardFragment on CustomDomain {
    domain
    managed
    certificate {
      status
      expiresAt
      provisioningError
    }
    dnsRecords {
      type
      name
      ...CompliancePortalDomainDnsRecordListItem_dnsRecord
    }
    canDelete: permission(action: "compliance-portal:custom-domain:delete")
    ...DeleteCompliancePortalDomainDialog_customDomain
  }
`;

interface CompliancePortalDomainCardProps {
  customDomainKey: CompliancePortalDomainCardFragment$key;
}

export function CompliancePortalDomainCard({ customDomainKey }: CompliancePortalDomainCardProps) {
  const { t, i18n } = useTranslation("organizations/compliance-portals");
  const domain = useFragment(fragment, customDomainKey);
  const sslStatus = domain.certificate?.status ?? "PENDING";
  const provisioningErrorMessage = getCertificateProvisioningErrorMessage(
    domain.certificate?.provisioningError,
    t,
  );
  const showDns = showsCustomDomainDns(domain.managed) && domain.dnsRecords.length > 0;
  const tone = customDomainCardTone(sslStatus);
  const { frame, header, wash, fade, icon: iconSlot, control: controlSlot, body } = hostingCard({
    tone,
    wide: showDns,
  });
  const { copy, identity, lead, subtitle, dns } = domainCard();
  const expiresLabel = domain.certificate?.expiresAt
    ? new Intl.DateTimeFormat(i18n.language, { dateStyle: "medium" }).format(
        new Date(domain.certificate.expiresAt),
      )
    : null;
  const callout = domainCardCallout(
    sslStatus,
    provisioningErrorMessage,
    expiresLabel,
    t,
  );

  return (
    <Card variant="ghost" size={3} padding="none" className={frame()}>
      <div className={header()}>
        <div className={wash()} />
        <div className={fade()} />
        <div className={lead()}>
          <div className={iconSlot()}>
            <CustomDomainStatusIcon status={sslStatus} />
          </div>
          <Text size={2} weight="medium" color={tone} className={subtitle()}>
            {callout.title}
          </Text>
        </div>
        {domain.canDelete && (
          <div className={controlSlot()}>
            <DeleteCompliancePortalDomainDialog customDomainKey={domain}>
              <IconButton
                variant="soft"
                color="red"
                aria-label={t("domainCard.actions.delete")}
              >
                <TrashIcon />
              </IconButton>
            </DeleteCompliancePortalDomainDialog>
          </div>
        )}
      </div>
      <div className={body()}>
        <div className={copy()}>
          <div className={identity()}>
            <Text size={4} weight="medium" highContrast>
              {domain.domain}
            </Text>
            {domain.managed && (
              <Badge variant="soft" color="neutral">{t("domainCard.managed")}</Badge>
            )}
          </div>
          <Text size={2} color="neutral">
            {callout.description}
          </Text>
          {callout.detail != null && (
            <Text size={1} color="neutral">
              {callout.detail}
            </Text>
          )}
        </div>
        {showDns && (
          <div className={dns()}>
            {callout.dns != null && (
              <Text size={2} color="faint">
                {callout.dns}
              </Text>
            )}
            {domain.dnsRecords.map(dnsRecord => (
              <CompliancePortalDomainDnsRecordListItem
                key={`${dnsRecord.type}:${dnsRecord.name}`}
                dnsRecordKey={dnsRecord}
              />
            ))}
          </div>
        )}
      </div>
    </Card>
  );
}
