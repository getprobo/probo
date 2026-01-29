import { useTranslate } from "@probo/i18n";
import { Button, Card, Checkbox, Spinner, useToast } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageStatusSectionFragment$key } from "#/__generated__/core/CompliancePageStatusSectionFragment.graphql";
import { useUpdateTrustCenterMutation } from "#/hooks/graph/TrustCenterGraph";

const fragment = graphql`
  fragment CompliancePageStatusSectionFragment on Organization {
    customDomain {
      domain
    }
    compliancePage: trustCenter {
      id
      active
      canUpdate: permission(action: "core:trust-center:update")
    }
  }
`;

export function CompliancePageStatusSection(props: {
  fragmentRef: CompliancePageStatusSectionFragment$key;
}) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const organization = useFragment<CompliancePageStatusSectionFragment$key>(
    fragment,
    fragmentRef,
  );

  const compliancePageUrl = organization.compliancePage?.id
    ? organization.customDomain?.domain
      ? `https://${organization.customDomain.domain}`
      : `${window.location.origin}/trust/${organization.compliancePage.id}`
    : null;

  const [updateCompliancePage, isUpdating] = useUpdateTrustCenterMutation();

  const handleToggleActive = async (active: boolean) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: __("Error"),
        description: __("Compliance page not found"),
        variant: "error",
      });
      return;
    }

    await updateCompliancePage({
      variables: {
        input: {
          trustCenterId: organization.compliancePage.id,
          active,
        },
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">{__("Compliance Page Status")}</h2>
        {isUpdating && <Spinner />}
      </div>
      <Card padded className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="font-medium">{__("Activate Compliance Page")}</h3>
            <p className="text-sm text-txt-tertiary">
              {__(
                "Make your compliance page publicly accessible to build customer confidence",
              )}
            </p>
          </div>
          <Checkbox
            checked={!!organization.compliancePage?.active}
            onChange={checked => void handleToggleActive(checked)}
            disabled={!organization.compliancePage?.canUpdate}
          />
        </div>

        {organization.compliancePage?.active && compliancePageUrl && (
          <div className="mt-4 p-4 bg-accent-light rounded-lg border border-accent">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="font-medium text-accent-dark">
                  {__("Your compliance page is live!")}
                </h4>
                <p className="text-sm text-accent-dark mt-1">
                  {__("Your customers can now access your compliance page at:")}
                </p>
                <a
                  href={compliancePageUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm font-mono text-accent underline hover:no-underline"
                >
                  {compliancePageUrl}
                </a>
              </div>
              <Button
                variant="secondary"
                onClick={() =>
                  window.open(compliancePageUrl, "_blank", "noopener,noreferrer")}
              >
                {__("View")}
              </Button>
            </div>
          </div>
        )}

        {!organization.compliancePage?.active && (
          <div className="mt-4 p-4 bg-tertiary rounded-lg border border-border-solid">
            <h4 className="font-medium text-txt-secondary">
              {__("Compliance page is inactive")}
            </h4>
            <p className="text-sm text-txt-tertiary mt-1">
              {__(
                "Your compliance page is currently not accessible to the public. Enable it to start sharing your compliance status.",
              )}
            </p>
          </div>
        )}
      </Card>
    </div>
  );
}
