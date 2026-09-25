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

import { CopyIcon } from "@phosphor-icons/react";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDomainDnsRecordListItem_dnsRecord$key } from "#/__generated__/core/CompliancePortalDomainDnsRecordListItem_dnsRecord.graphql";

import { domainCard } from "../variants";

const dnsRecordFragment = graphql`
  fragment CompliancePortalDomainDnsRecordListItem_dnsRecord on DNSRecordInstruction {
    type
    name
    value
    ttl
  }
`;

interface CompliancePortalDomainDnsRecordListItemProps {
  dnsRecordKey: CompliancePortalDomainDnsRecordListItem_dnsRecord$key;
}

export function CompliancePortalDomainDnsRecordListItem({
  dnsRecordKey,
}: CompliancePortalDomainDnsRecordListItemProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { toast } = useToast();
  const dnsRecord = useFragment(dnsRecordFragment, dnsRecordKey);
  const { record, recordHeader, recordField, recordValue, code } = domainCard();

  async function copyToClipboard(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      toast({
        title: t("domainCard.messages.copyFailed"),
        description: t("domainCard.messages.valueCopyFailed"),
        variant: "error",
      });
      return;
    }
    toast({
      title: t("domainCard.messages.copied"),
      description: t("domainCard.messages.valueCopied"),
      variant: "success",
    });
  }

  return (
    <div className={record()}>
      <div className={recordHeader()}>
        <Text size={2} weight="medium">{dnsRecord.type}</Text>
        <Badge variant="soft" color="neutral">
          {t("domainCard.dns.purposePointToServers")}
        </Badge>
      </div>
      <div className={recordField()}>
        <Text size={1} color="faint">{t("domainCard.dns.name")}</Text>
        <div className={recordValue()}>
          <Code size={2} className={code()}>{dnsRecord.name}</Code>
          <IconButton
            size={1}
            variant="soft"
            color="neutral"
            aria-label={t("domainCard.dns.copy")}
            onClick={() => void copyToClipboard(dnsRecord.name)}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
      <div className={recordField()}>
        <Text size={1} color="faint">{t("domainCard.dns.value")}</Text>
        <div className={recordValue()}>
          <Code size={2} className={code()}>{dnsRecord.value}</Code>
          <IconButton
            size={1}
            variant="soft"
            color="neutral"
            aria-label={t("domainCard.dns.copy")}
            onClick={() => void copyToClipboard(dnsRecord.value)}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
      <Text size={1} color="faint">
        {t("domainCard.dns.ttl", { ttl: String(dnsRecord.ttl) })}
      </Text>
    </div>
  );
}
