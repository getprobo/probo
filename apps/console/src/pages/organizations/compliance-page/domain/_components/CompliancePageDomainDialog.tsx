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
  getCustomDomainStatusBadgeVariant,
} from "@probo/helpers";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { PropsWithChildren } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageDomainDialogFragment$key } from "#/__generated__/core/CompliancePageDomainDialogFragment.graphql";

const fragment = graphql`
  fragment CompliancePageDomainDialogFragment on CustomDomain {
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

type CompliancePageDomainDialogProps = PropsWithChildren<{ fKey: CompliancePageDomainDialogFragment$key }>;

export function CompliancePageDomainDialog(props: CompliancePageDomainDialogProps) {
  const { children, fKey } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const dialogRef = useDialogRef();
  const { toast } = useToast();

  const copyToClipboard = (text: string) => {
    void navigator.clipboard.writeText(text);
    toast({
      title: t("domainDialog.messages.copied"),
      description: t("domainDialog.messages.valueCopied"),
      variant: "success",
    });
  };

  const domain = useFragment<CompliancePageDomainDialogFragment$key>(fragment, fKey);
  const sslStatus = domain.certificate?.status ?? "PENDING";
  const expiresAt = domain.certificate?.expiresAt;
  const provisioningErrorMessage = getCertificateProvisioningErrorMessage(
    domain.certificate?.provisioningError,
    t,
  );

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={(
        <div className="flex items-center gap-3">
          <span>{domain.domain}</span>
          <Badge variant={getCustomDomainStatusBadgeVariant(sslStatus)}>
            {getCustomDomainStatusBadgeLabel(sslStatus, t)}
          </Badge>
        </div>
      )}
    >
      <DialogContent padded className="space-y-6">
        {sslStatus === "ACTIVE"
          ? (
              <div className="bg-subtle rounded-lg p-4">
                <div className="flex items-start">
                  <svg
                    className="w-5 h-5 text-green-500 mt-0.5 mr-3 shrink-0"
                    fill="none"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <div>
                    <p className="font-medium mb-1">{t("domainDialog.active.title")}</p>
                    <p className="text-sm text-txt-secondary">
                      {t("domainDialog.active.description")}
                    </p>
                    {expiresAt && (
                      <p className="text-xs text-txt-tertiary mt-2">
                        {t("domainDialog.sslExpires")}
                        {" "}
                        {new Date(expiresAt).toLocaleDateString()}
                      </p>
                    )}
                  </div>
                </div>
              </div>
            )
          : (
              <div>
                {provisioningErrorMessage && (
                  <div className="bg-danger-subtle text-danger rounded-lg p-4 mb-4">
                    <p className="text-sm font-medium mb-1">{t("domainDialog.provisioningError")}</p>
                    <p className="text-sm">{provisioningErrorMessage}</p>
                  </div>
                )}

                <h4 className="font-medium mb-3">{t("domainDialog.dns.title")}</h4>
                <p className="text-sm text-txt-secondary mb-4">
                  {t("domainDialog.dns.description")}
                </p>

                <div className="space-y-3">
                  {domain.dnsRecords?.map((record, index) => (
                    <div key={index} className="bg-subtle rounded-lg p-4">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">{record.type}</span>
                        <Badge variant="neutral">{record.purpose}</Badge>
                      </div>
                      <div className="space-y-2">
                        <div>
                          <label className="text-xs text-txt-tertiary">
                            {t("domainDialog.dns.name")}
                          </label>
                          <div className="flex items-center gap-2 mt-1">
                            <code className="flex-1 text-sm bg-subtle px-2 py-1 rounded">
                              {record.name}
                            </code>
                            <Button
                              variant="secondary"
                              onClick={() => copyToClipboard(record.name)}
                            >
                              {t("domainDialog.actions.copy")}
                            </Button>
                          </div>
                        </div>
                        <div>
                          <label className="text-xs text-txt-tertiary">
                            {t("domainDialog.dns.value")}
                          </label>
                          <div className="flex items-center gap-2 mt-1">
                            <code className="flex-1 text-sm bg-subtle px-2 py-1 rounded break-all">
                              {record.value}
                            </code>
                            <Button
                              variant="secondary"
                              onClick={() => copyToClipboard(record.value)}
                            >
                              {t("domainDialog.actions.copy")}
                            </Button>
                          </div>
                        </div>
                        {record.ttl && (
                          <div className="text-xs text-txt-tertiary">
                            {t("domainDialog.dns.ttl", { ttl: record.ttl })}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>

                {sslStatus === "PENDING" && (
                  <div className="bg-subtle rounded-lg p-4 mt-4">
                    <p className="text-sm">
                      {t("domainDialog.pendingDescription")}
                    </p>
                  </div>
                )}
              </div>
            )}
      </DialogContent>
    </Dialog>
  );
}
