import { useState } from "react";
import { useTranslate } from "@probo/i18n";
import { graphql, useLazyLoadQuery, useMutation } from "react-relay";
import { useOrganizationId } from "/hooks/useOrganizationId";
import {
  Button,
  Card,
  Badge,
  Field,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
  useToast,
} from "@probo/ui";
import type { CustomDomainManagerQuery } from "./__generated__/CustomDomainManagerQuery.graphql";
import type { CustomDomainManagerCreateMutation } from "./__generated__/CustomDomainManagerCreateMutation.graphql";
import type { CustomDomainManagerDeleteMutation } from "./__generated__/CustomDomainManagerDeleteMutation.graphql";

const customDomainsQuery = graphql`
  query CustomDomainManagerQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        customDomains(first: 100) {
          edges {
            node {
              id
              domain
              sslStatus
              isActive
              dnsRecords {
                type
                name
                value
                ttl
                purpose
              }
              createdAt
              updatedAt
              verifiedAt
              sslExpiresAt
            }
          }
        }
      }
    }
  }
`;

const createCustomDomainMutation = graphql`
  mutation CustomDomainManagerCreateMutation($input: CreateCustomDomainInput!) {
    createCustomDomain(input: $input) {
      customDomainEdge {
        node {
          id
          domain
          sslStatus
          isActive
          dnsRecords {
            type
            name
            value
            ttl
            purpose
          }
          createdAt
          updatedAt
          verifiedAt
          sslExpiresAt
        }
      }
    }
  }
`;

const deleteCustomDomainMutation = graphql`
  mutation CustomDomainManagerDeleteMutation($input: DeleteCustomDomainInput!) {
    deleteCustomDomain(input: $input) {
      deletedCustomDomainId
    }
  }
`;

function DomainDetailsDialog({
  domain,
  children,
  onDelete,
  isDeletingDomain,
}: {
  domain: any;
  children: React.ReactNode;
  onDelete: (domainId: string, domainName: string) => void;
  isDeletingDomain: boolean;
}) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const getStatusBadge = (domain: any) => {
    if (domain.sslStatus === "ACTIVE") {
      return <Badge variant="success">{__("Active")}</Badge>;
    }
    if (domain.sslStatus === "PROVISIONING" || domain.sslStatus === "RENEWING") {
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
          <div className="bg-level-2 rounded-lg p-4">
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
                  {__("Your custom domain is verified and SSL certificate is active")}
                </p>
                {domain.sslExpiresAt && (
                  <p className="text-xs text-txt-tertiary mt-2">
                    {__("SSL expires")} {new Date(domain.sslExpiresAt).toLocaleDateString()}
                  </p>
                )}
              </div>
            </div>
          </div>
        ) : (
          <div>
            <h4 className="font-medium mb-3">{__("DNS Configuration")}</h4>
            <p className="text-sm text-txt-secondary mb-4">
              {__("Add these DNS records to your domain to complete verification")}
            </p>

            <div className="space-y-3">
              {domain.dnsRecords?.map((record: any, index: number) => (
                <div key={index} className="bg-level-2 rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm font-medium">{record.type}</span>
                    <Badge variant="neutral">{record.purpose}</Badge>
                  </div>
                  <div className="space-y-2">
                    <div>
                      <label className="text-xs text-txt-tertiary">{__("Name")}</label>
                      <div className="flex items-center gap-2 mt-1">
                        <code className="flex-1 text-sm bg-level-3 px-2 py-1 rounded">
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
                      <label className="text-xs text-txt-tertiary">{__("Value")}</label>
                      <div className="flex items-center gap-2 mt-1">
                        <code className="flex-1 text-sm bg-level-3 px-2 py-1 rounded break-all">
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
              <div className="bg-level-2 rounded-lg p-4 mt-4">
                <p className="text-sm">
                  {__("After adding the DNS records, verification will happen automatically. This may take a few minutes to propagate.")}
                </p>
              </div>
            )}
          </div>
        )}

        <div className="pt-4 border-t border-border">
          <Button
            variant="danger"
            onClick={() => {
              dialogRef.current?.close();
              onDelete(domain.id, domain.domain);
            }}
            disabled={isDeletingDomain}
          >
            {__("Delete Domain")}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export function CustomDomainManager() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  const data = useLazyLoadQuery<CustomDomainManagerQuery>(
    customDomainsQuery,
    { organizationId },
    { fetchPolicy: "network-only" }
  );

  const [createCustomDomain, isCreatingDomain] =
    useMutation<CustomDomainManagerCreateMutation>(createCustomDomainMutation);
  const [deleteCustomDomain, isDeletingDomain] =
    useMutation<CustomDomainManagerDeleteMutation>(deleteCustomDomainMutation);

  const [newDomain, setNewDomain] = useState("");

  const domains =
    data.organization?.customDomains?.edges?.map((edge) => edge.node) || [];

  const validateDomain = (domainInput: string): boolean => {
    const domainRegex =
      /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i;
    return domainRegex.test(domainInput);
  };

  const handleAddDomain = () => {
    const trimmedDomain = newDomain
      .trim()
      .toLowerCase()
      .replace(/^https?:\/\//, "")
      .replace(/\/$/, "");

    if (!validateDomain(trimmedDomain)) {
      toast({
        title: __("Invalid Domain"),
        description: __(
          "Please enter a valid domain (e.g., compliance.example.com)"
        ),
        variant: "error",
      });
      return;
    }

    createCustomDomain({
      variables: {
        input: {
          organizationId,
          domain: trimmedDomain,
        },
      },
      onCompleted: () => {
        setNewDomain("");
        dialogRef.current?.close();

        toast({
          title: __("Domain Added Successfully"),
          description: __(
            "Configure the DNS records below to verify and activate your domain"
          ),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Failed to Add Domain"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const handleDeleteDomain = (domainId: string, domainName: string) => {
    if (!confirm(`${__("Are you sure you want to delete")} ${domainName}?`)) {
      return;
    }

    deleteCustomDomain({
      variables: {
        input: { domainId },
      },
      onCompleted: () => {
        toast({
          title: __("Domain Deleted"),
          description: __("The domain has been removed"),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Failed to Delete Domain"),
          description: error.message,
          variant: "error",
        });
      },
    });
  };

  const getStatusBadge = (domain: any) => {
    if (domain.sslStatus === "ACTIVE") {
      return <Badge variant="success">{__("Active")}</Badge>;
    }
    if (domain.sslStatus === "PROVISIONING" || domain.sslStatus === "RENEWING") {
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

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-semibold">{__("Custom Domains")}</h2>
          <p className="text-sm text-txt-secondary mt-1">
            {__("Use your own domain for your trust center")}
          </p>
        </div>
        <Button onClick={() => dialogRef.current?.open()}>
          {__("Add Domain")}
        </Button>
      </div>

      {domains.length === 0 ? (
        <Card padded>
          <div className="text-center py-12">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-level-2 rounded-full mb-4">
              <svg
                className="w-8 h-8 text-txt-tertiary"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
              </svg>
            </div>
            <h3 className="text-lg font-medium mb-2">
              {__("No custom domains yet")}
            </h3>
            <p className="text-sm text-txt-secondary mb-6">
              {__(
                "Add your own domain to make your trust center more professional"
              )}
            </p>
            <Button onClick={() => dialogRef.current?.open()}>
              {__("Add Your First Domain")}
            </Button>
          </div>
        </Card>
      ) : (
        <Card>
          <div className="divide-y divide-border">
            {domains.map((domain: any) => (
              <DomainDetailsDialog
                key={domain.id}
                domain={domain}
                onDelete={handleDeleteDomain}
                isDeletingDomain={isDeletingDomain}
              >
                <div className="p-4 cursor-pointer hover:bg-level-1 transition-colors">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-medium mb-1">{domain.domain}</div>
                      {getStatusBadge(domain)}
                    </div>
                    <svg
                      className="w-5 h-5 text-txt-tertiary"
                      fill="none"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </div>
              </DomainDetailsDialog>
            ))}
          </div>
        </Card>
      )}

      <Dialog
        ref={dialogRef}
        onClose={() => {
          setNewDomain("");
        }}
      >
        <DialogContent className="max-w-md" padded>
          <div className="space-y-4">
            <div>
              <h3 className="text-lg font-semibold mb-2">
                {__("Add Custom Domain")}
              </h3>
              <p className="text-sm text-txt-secondary">
                {__(
                  "Enter your domain and we'll generate the DNS records you need to add"
                )}
              </p>
            </div>

            <Field
              label={__("Domain")}
              name="domain"
              value={newDomain}
              onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                setNewDomain(e.target.value)
              }
              placeholder="compliance.example.com"
              help={__("Enter without http:// or https://")}
              autoFocus
            />

            <div className="bg-level-2 rounded-lg p-3">
              <p className="text-xs text-txt-secondary">
                <strong>{__("Examples:")}</strong> compliance.example.com,
                trust.example.com
              </p>
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            onClick={handleAddDomain}
            disabled={isCreatingDomain || !newDomain.trim()}
          >
            {isCreatingDomain ? __("Adding...") : __("Add Domain")}
          </Button>
        </DialogFooter>
      </Dialog>
    </div>
  );
}
