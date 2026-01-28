import { formatDate, getAuditStateLabel, getAuditStateVariant, getTrustCenterVisibilityOptions } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Field, Option, Td, Tr } from "@probo/ui";
import { useCallback } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageAuditListItem_auditFragment$key } from "#/__generated__/core/CompliancePageAuditListItem_auditFragment.graphql";
import type { CompliancePageAuditListItem_compliancePageFragment$key } from "#/__generated__/core/CompliancePageAuditListItem_compliancePageFragment.graphql";
import type { CompliancePageAuditListItem_updateAuditVisibilityMutation } from "#/__generated__/core/CompliancePageAuditListItem_updateAuditVisibilityMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const compliancePageFragment = graphql`
  fragment CompliancePageAuditListItem_compliancePageFragment on TrustCenter {
    canUpdate: permission(action: "core:trust-center:update")
  }
`;

const auditFragment = graphql`
  fragment CompliancePageAuditListItem_auditFragment on Audit {
    id
    name
    framework {
      name
    }
    validUntil
    state
    trustCenterVisibility
  }
`;

const updateAuditVisibilityMutation = graphql`
  mutation CompliancePageAuditListItem_updateAuditVisibilityMutation($input: UpdateAuditInput!) {
    updateAudit(input: $input) {
      audit {
        ...CompliancePageAuditListItem_auditFragment
      }
    }
  }
`;

export function CompliancePageAuditListItem(props: {
  auditFragmentRef: CompliancePageAuditListItem_auditFragment$key;
  compliancePageFragmentRef: CompliancePageAuditListItem_compliancePageFragment$key;
}) {
  const { auditFragmentRef, compliancePageFragmentRef } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();

  const compliancePage = useFragment<CompliancePageAuditListItem_compliancePageFragment$key>(
    compliancePageFragment,
    compliancePageFragmentRef,
  );
  const audit = useFragment<CompliancePageAuditListItem_auditFragment$key>(auditFragment, auditFragmentRef);

  const [updateAuditVisibility, isUpdatingAuditVisibility] = useMutationWithToasts<
    CompliancePageAuditListItem_updateAuditVisibilityMutation
  >(
    updateAuditVisibilityMutation,
    {
      successMessage: __("Audit visibility updated successfully."),
      errorMessage: __("Failed to update audit visibility"),
    },
  );
  const handleVisibilityChange = useCallback(
    async (value: string) => {
      const stringValue = typeof value === "string" ? value : "";
      const typedValue = stringValue as "NONE" | "PRIVATE" | "PUBLIC";
      await updateAuditVisibility({
        variables: {
          input: {
            id: audit.id,
            trustCenterVisibility: typedValue,
          },
        },
      });
    },
    [audit.id, updateAuditVisibility],
  );

  const visibilityOptions = getTrustCenterVisibilityOptions(__);
  const validUntilFormatted = audit.validUntil
    ? formatDate(audit.validUntil)
    : __("No expiry");

  return (
    <Tr to={`/organizations/${organizationId}/audits/${audit.id}`}>
      <Td>
        <div className="flex gap-4 items-center">{audit.framework.name}</div>
      </Td>
      <Td>{audit.name || __("Untitled")}</Td>
      <Td>{validUntilFormatted}</Td>
      <Td>
        <Badge variant={getAuditStateVariant(audit.state)}>
          {getAuditStateLabel(__, audit.state)}
        </Badge>
      </Td>
      <Td noLink width={130} className="pr-0">
        <Field
          type="select"
          value={audit.trustCenterVisibility}
          onValueChange={value => void handleVisibilityChange(value)}
          disabled={isUpdatingAuditVisibility || !compliancePage.canUpdate}
          className="w-[105px]"
        >
          {visibilityOptions.map(option => (
            <Option key={option.value} value={option.value}>
              <div className="flex items-center justify-between w-full">
                <Badge variant={option.variant}>{option.label}</Badge>
              </div>
            </Option>
          ))}
        </Field>
      </Td>
    </Tr>
  );
}
