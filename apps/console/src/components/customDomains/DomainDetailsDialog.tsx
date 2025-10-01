import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  Badge,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { ReactNode } from "react";

interface DomainDetailsDialogProps {
  children: ReactNode;
  domain: {
    readonly id: string;
    readonly domain: string;
    readonly sslStatus: string;
    readonly dnsRecords?:
      | readonly {
          readonly type: string;
          readonly name: string;
          readonly value: string;
          readonly ttl?: number;
          readonly purpose: string;
        }[]
      | null;
    readonly verifiedAt?: string | null;
    readonly sslExpiresAt?: string | null;
  };
}

export function DomainDetailsDialog({
  children,
  domain,
}: DomainDetailsDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const getStatusBadge = (domain: DomainDetailsDialogProps["domain"]) => {
    if (domain.sslStatus === "ACTIVE") {
      return <Badge variant="success">{__("Active")}</Badge>;
    }
    if (
      domain.sslStatus === "PROVISIONING" ||
      domain.sslStatus === "RENEWING"
    ) {
      return <Badge variant="warning">{__("Provisioning")}</Badge>;
    }
    if (domain.sslStatus === "PENDING") {
      return <Badge variant="warning">{__("Pending")}</Badge>;
    }
    if (domain.sslStatus === "FAILED") {
      return <Badge variant="danger">{__("Failed")}</Badge>;
    }
    if (domain.sslStatus === "EXPIRED") {
      return <Badge variant="danger">{__("Expired")}</Badge>;
    }
    return <Badge variant="neutral">{__("Unknown")}</Badge>;
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    toast({
      title: __("Copied"),
      description: __("Value copied to clipboard"),
      variant: "success",
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={
        <div className="flex items-center gap-3">
          <span>{domain.domain}</span>
          {getStatusBadge(domain)}
        </div>
      }
    >
      <DialogContent padded className="space-y-6">
        {domain.verifiedAt && (
          <p className="text-sm text-txt-secondary">
            {__("Verified")} {new Date(domain.verifiedAt).toLocaleDateString()}
          </p>
        )}

        {domain.sslStatus === "ACTIVE" ? (
          <div className="bg-subtle rounded-lg p-4">
            <div className="flex items-start">
              <svg
                className="w-5 h-5 text-green-500 mt-0.5 mr-3 flex-shrink-0"
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
                <p className="font-medium mb-1">{__("Domain is active")}</p>
                <p className="text-sm text-txt-secondary">
                  {__(
                    "Your custom domain is verified and SSL certificate is active"
                  )}
                </p>
                {domain.sslExpiresAt && (
                  <p className="text-xs text-txt-tertiary mt-2">
                    {__("SSL expires")}{" "}
                    {new Date(domain.sslExpiresAt).toLocaleDateString()}
                  </p>
                )}
              </div>
            </div>
          </div>
        ) : (
          <div>
            <h4 className="font-medium mb-3">{__("DNS Configuration")}</h4>
            <p className="text-sm text-txt-secondary mb-4">
              {__(
                "Add these DNS records to your domain to complete verification"
              )}
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
                        {__("Name")}
                      </label>
                      <div className="flex items-center gap-2 mt-1">
                        <code className="flex-1 text-sm bg-subtle px-2 py-1 rounded">
                          {record.name}
                        </code>
                        <Button
                          variant="secondary"
                          onClick={() => copyToClipboard(record.name)}
                        >
                          {__("Copy")}
                        </Button>
                      </div>
                    </div>
                    <div>
                      <label className="text-xs text-txt-tertiary">
                        {__("Value")}
                      </label>
                      <div className="flex items-center gap-2 mt-1">
                        <code className="flex-1 text-sm bg-subtle px-2 py-1 rounded break-all">
                          {record.value}
                        </code>
                        <Button
                          variant="secondary"
                          onClick={() => copyToClipboard(record.value)}
                        >
                          {__("Copy")}
                        </Button>
                      </div>
                    </div>
                    {record.ttl && (
                      <div className="text-xs text-txt-tertiary">
                        TTL: {record.ttl}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>

            {domain.sslStatus === "PENDING" && (
              <div className="bg-subtle rounded-lg p-4 mt-4">
                <p className="text-sm">
                  {__(
                    "After adding the DNS records, verification will happen automatically. This may take a few minutes to propagate."
                  )}
                </p>
              </div>
            )}
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
