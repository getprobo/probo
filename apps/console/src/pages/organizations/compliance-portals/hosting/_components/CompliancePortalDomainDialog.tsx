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

import { CheckCircleIcon, CopyIcon, WarningCircleIcon } from "@phosphor-icons/react";
import {
  getCertificateProvisioningErrorMessage,
  getCustomDomainStatusBadgeLabel,
} from "@probo/helpers";
import { useToast } from "@probo/ui";
import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import type { ReactElement } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalDomainDialogFragment$key } from "#/__generated__/core/CompliancePortalDomainDialogFragment.graphql";

import { customDomainBadgeColor } from "../_lib/customDomainBadgeColor";
import { domainDetailsDialog } from "../variants";

const fragment = graphql`
  fragment CompliancePortalDomainDialogFragment on CustomDomain {
    domain
    certificate {
      status
      expiresAt
      provisioningError
    }
    dnsRecords {
      type
      name
      value
      ttl
      purpose
    }
  }
`;

interface CompliancePortalDomainDialogProps {
  customDomainKey: CompliancePortalDomainDialogFragment$key;
  children: ReactElement;
}

export function CompliancePortalDomainDialog({
  customDomainKey,
  children,
}: CompliancePortalDomainDialogProps) {
  const { t, i18n } = useTranslation("organizations/compliance-portals");
  const { toast } = useToast();
  const {
    titleRow, body, calloutBody, dns, records, record, recordHeader, recordField, recordValue, code,
  } = domainDetailsDialog();

  const domain = useFragment<CompliancePortalDomainDialogFragment$key>(fragment, customDomainKey);
  const sslStatus = domain.certificate?.status ?? "PENDING";
  const expiresAt = domain.certificate?.expiresAt;
  const provisioningErrorMessage = getCertificateProvisioningErrorMessage(
    domain.certificate?.provisioningError,
    t,
  );

  function copyToClipboard(text: string) {
    void navigator.clipboard.writeText(text);
    toast({
      title: t("domainDialog.messages.copied"),
      description: t("domainDialog.messages.valueCopied"),
      variant: "success",
    });
  }

  const expiresLabel = expiresAt
    ? new Intl.DateTimeFormat(i18n.language, { dateStyle: "medium" }).format(new Date(expiresAt))
    : null;

  return (
    <Dialog>
      <DialogTrigger render={children} />
      <DialogPopup>
        <DialogHeader>
          <div className={titleRow()}>
            <DialogTitle>{domain.domain}</DialogTitle>
            <Badge variant="soft" color={customDomainBadgeColor(sslStatus)}>
              {getCustomDomainStatusBadgeLabel(sslStatus, t)}
            </Badge>
          </div>
        </DialogHeader>
        <DialogBody>
          <div className={body()}>
            {sslStatus === "ACTIVE"
              ? (
                  <Callout color="green" icon={<CheckCircleIcon weight="fill" />}>
                    <div className={calloutBody()}>
                      <Text size={2} weight="medium" color="current">
                        {t("domainDialog.active.title")}
                      </Text>
                      <Text size={2} color="current">
                        {t("domainDialog.active.description")}
                      </Text>
                      {expiresLabel && (
                        <Text size={1} color="current">
                          {t("domainDialog.sslExpires")}
                          {" "}
                          {expiresLabel}
                        </Text>
                      )}
                    </div>
                  </Callout>
                )
              : (
                  <div className={dns()}>
                    {provisioningErrorMessage && (
                      <Callout color="red" icon={<WarningCircleIcon weight="fill" />}>
                        <div className={calloutBody()}>
                          <Text size={2} weight="medium" color="current">
                            {t("domainDialog.provisioningError")}
                          </Text>
                          <Text size={2} color="current">{provisioningErrorMessage}</Text>
                        </div>
                      </Callout>
                    )}

                    <div className={calloutBody()}>
                      <Heading level={3} size={3} weight="medium">
                        {t("domainDialog.dns.title")}
                      </Heading>
                      <Text size={2} color="faint">
                        {t("domainDialog.dns.description")}
                      </Text>
                    </div>

                    <div className={records()}>
                      {domain.dnsRecords?.map((dnsRecord, index) => (
                        <div key={index} className={record()}>
                          <div className={recordHeader()}>
                            <Text size={2} weight="medium">{dnsRecord.type}</Text>
                            <Badge variant="soft" color="neutral">{dnsRecord.purpose}</Badge>
                          </div>
                          <div className={recordField()}>
                            <Text size={1} color="faint">{t("domainDialog.dns.name")}</Text>
                            <div className={recordValue()}>
                              <Code className={code()}>{dnsRecord.name}</Code>
                              <IconButton
                                variant="soft"
                                color="neutral"
                                aria-label={t("domainDialog.actions.copy")}
                                onClick={() => copyToClipboard(dnsRecord.name)}
                              >
                                <CopyIcon />
                              </IconButton>
                            </div>
                          </div>
                          <div className={recordField()}>
                            <Text size={1} color="faint">{t("domainDialog.dns.value")}</Text>
                            <div className={recordValue()}>
                              <Code className={code()}>{dnsRecord.value}</Code>
                              <IconButton
                                variant="soft"
                                color="neutral"
                                aria-label={t("domainDialog.actions.copy")}
                                onClick={() => copyToClipboard(dnsRecord.value)}
                              >
                                <CopyIcon />
                              </IconButton>
                            </div>
                          </div>
                          {dnsRecord.ttl && (
                            <Text size={1} color="faint">
                              {t("domainDialog.dns.ttl", { ttl: dnsRecord.ttl })}
                            </Text>
                          )}
                        </div>
                      ))}
                    </div>

                    {sslStatus === "PENDING" && (
                      <Callout color="neutral">
                        <Text size={2} color="current">
                          {t("domainDialog.pendingDescription")}
                        </Text>
                      </Callout>
                    )}
                  </div>
                )}
          </div>
        </DialogBody>
      </DialogPopup>
    </Dialog>
  );
}
