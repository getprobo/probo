import { useTranslate } from "@probo/i18n";
import { Button, Card, Badge, IconPlusLarge } from "@probo/ui";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { graphql } from "relay-runtime";
import type { CustomDomainManagerDeleteMutation } from "./__generated__/CustomDomainManagerDeleteMutation.graphql";
import { CreateCustomDomainDialog } from "./CreateCustomDomainDialog";
import { DeleteCustomDomainDialog } from "./DeleteCustomDomainDialog";
import { DomainDetailsDialog } from "./DomainDetailsDialog";
import { IfAuthorized } from "/permissions/IfAuthorized";

const deleteCustomDomainMutation = graphql`
  mutation CustomDomainManagerDeleteMutation($input: DeleteCustomDomainInput!) {
    deleteCustomDomain(input: $input) {
      deletedCustomDomainId
    }
  }
`;

interface CustomDomainManagerProps {
  organizationId: string;
  customDomain:
    | {
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
        readonly createdAt?: string | null;
        readonly updatedAt?: string | null;
        readonly sslExpiresAt?: string | null;
      }
    | null
    | undefined;
}

export function CustomDomainManager({
  organizationId,
  customDomain,
}: CustomDomainManagerProps) {
  const { __ } = useTranslate();

  const [deleteCustomDomain] =
    useMutationWithToasts<CustomDomainManagerDeleteMutation>(
      deleteCustomDomainMutation,
      {
        successMessage: __("Domain deleted successfully"),
        errorMessage: __("Failed to delete domain"),
      }
    );

  const domain = customDomain;

  const getStatusBadge = (domain: any) => {
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

  const handleDeleteDomain = async () => {
    return deleteCustomDomain({
      variables: {
        input: { organizationId },
      },
      updater: (store) => {
        // Update the cache by setting customDomain to null
        const organizationRecord = store.get(organizationId);
        if (organizationRecord) {
          organizationRecord.setValue(null, "customDomain");
        }
      },
    });
  };

  if (!domain) {
    return (
      <Card padded>
        <div className="text-center py-8">
          <h3 className="text-lg font-semibold mb-2">
            {__("No custom domain configured")}
          </h3>
          <p className="text-txt-tertiary mb-4">
            {__(
              "Add your own domain to make your trust center more professional"
            )}
          </p>
          <div className="flex justify-center">
            <IfAuthorized entity="Organization" action="update">
              <CreateCustomDomainDialog organizationId={organizationId}>
                <Button icon={IconPlusLarge}>{__("Add Domain")}</Button>
              </CreateCustomDomainDialog>
            </IfAuthorized>
          </div>
        </div>
      </Card>
    );
  }

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
                  : __("Pending verification")}
              </div>
            </div>
            {getStatusBadge(domain)}
          </div>

          <div className="flex items-center gap-2">
            <DomainDetailsDialog domain={domain}>
              <Button variant="secondary">{__("View Details")}</Button>
            </DomainDetailsDialog>

            <IfAuthorized entity="Organization" action="delete">
              <DeleteCustomDomainDialog
                domainName={domain.domain}
                onConfirm={handleDeleteDomain}
              >
                <Button variant="danger">{__("Delete")}</Button>
              </DeleteCustomDomainDialog>
            </IfAuthorized>
          </div>
        </div>
      </div>
    </Card>
  );
}
